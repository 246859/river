package riverdb

import (
	"github.com/246859/river/entry"
	"github.com/246859/river/index"
	"github.com/246859/river/pkg/str"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
	"time"
)

var ErrBatchClosed = errors.New("batch is closed")

type BatchOption struct {
	// size of per batch
	Size int64
	// call Fsync after per batch has been written
	SyncPerBatch bool
	// call Fsync after all batch finished, not recommended enabled both SyncPerBatch and SyncOnFlush simultaneously
	// if both of SyncPerBatch and SyncOnFlush is false, batch will apply sync rules in db.options
	SyncOnFlush bool
}

// Record is a user-oriented struct representing an entry data in db
type Record struct {
	K   Key
	V   Value
	TTL time.Duration
}

// Batch provides ability to write or delete entries in batch which is update-only,
// it will split given records into several batches, and run a new goroutine for each batch to handle in batchTxn.
// each batch records will be stored in memory temporarily, then they will be written to the database in batch
// finally update db index if commit succeeds.
func (db *DB) Batch(opt BatchOption) (*Batch, error) {
	if opt.Size <= 0 {
		return nil, errors.New("invalid batch size")
	}

	return &Batch{
		opt: opt,
		db:  db,
		pendingPool: sync.Pool{New: func() any {
			return make(map[string]*entry.Entry, opt.Size)
		}},
	}, nil
}

// Batch operator
type Batch struct {
	db  *DB
	opt BatchOption

	mu sync.Mutex
	wg sync.WaitGroup

	err      atomic.Value
	effected atomic.Int64

	pendingPool sync.Pool

	closed bool
}

// Error returns the error occurred in the batch operation
func (ba *Batch) Error() error {
	val := ba.err.Load()
	if val == nil {
		return nil
	}
	return val.(error)
}

// Effected returns how many records has been effected
func (ba *Batch) Effected() int64 {
	return ba.effected.Load()
}

func (ba *Batch) writeAll(rs []Record, et entry.EType) error {
	if err := ba.Error(); err != nil {
		return err
	}

	lrs := int64(len(rs))
	for lrs > 0 {
		// cut records into per batch by BatchOption.Size
		var batchRecord []Record
		if int64(len(rs)) < ba.opt.Size {
			batchRecord = rs
			lrs -= int64(len(rs))
		} else {
			batchRecord = rs[:ba.opt.Size:ba.opt.Size]
			rs = rs[ba.opt.Size:]
			lrs -= ba.opt.Size
		}

		// run a new goroutine to handle the batch in batchTxn
		ba.wg.Add(1)
		go func() {
			defer ba.wg.Done()

			// start a txn
			txn, err := ba.db.Begin(false)
			if err != nil {
				return
			}

			bt := batchTxn{txn: txn, ba: ba}
			// get pending buffer from pool
			bt.pendingBuf = ba.pendingPool.Get().(map[string]*entry.Entry)
			defer func() {
				clear(bt.pendingBuf)
				ba.pendingPool.Put(bt.pendingBuf)
			}()

			if err = bt.write(batchRecord, et); err != nil {
				ba.err.Store(err)
				_ = bt.rollback()
				return
			}

			if err = bt.commit(ba.opt.SyncPerBatch); err != nil {
				ba.err.Store(err)
				_ = bt.rollback()
				return
			}
		}()
	}

	return ba.Error()
}

// WriteAll writes all given records to db in batchTxn
func (ba *Batch) WriteAll(records []Record) error {
	if len(records) == 0 {
		return nil
	}

	ba.mu.Lock()
	defer ba.mu.Unlock()

	if ba.closed {
		return ErrBatchClosed
	}
	return ba.writeAll(records, entry.DataEntryType)
}

// DeleteAll delete all given key matching records from db in batchTxn
func (ba *Batch) DeleteAll(keys []Key) error {
	if len(keys) == 0 {
		return nil
	}

	ba.mu.Lock()
	defer ba.mu.Unlock()

	if ba.closed {
		return ErrBatchClosed
	}

	var rs []Record
	for _, key := range keys {
		rs = append(rs, Record{K: key})
	}
	return ba.writeAll(rs, entry.DeletedEntryType)
}

// Flush close Batch, then waiting for the remaining batchTxn to complete,
// and call db.Sync to Flush db finally.
func (ba *Batch) Flush() error {
	ba.mu.Lock()
	defer ba.mu.Unlock()
	ba.wg.Wait()
	ba.closed = true

	if ba.opt.SyncOnFlush {
		// manually sync
		if err := ba.db.Sync(); err != nil {
			return err
		}
	}

	return ba.Error()
}

// batchTxn wrap Txn to perform batch operations
type batchTxn struct {
	txn        *Txn
	ba         *Batch
	pendingBuf map[string]*entry.Entry
}

func (b *batchTxn) putAll(ens []*entry.Entry, needSync bool) error {
	db := b.txn.db
	p := b.txn.pending
	all, err := db.writeAll(ens, needSync)
	if err != nil {
		return err
	}
	for i, pos := range all {
		if err := p.memIndex.Put(index.Hint{ChunkPos: pos, TTL: ens[i].TTL, Key: ens[i].Key, Meta: ens[i].Type}); err != nil {
			return err
		}
		b.txn.trackWrite(ens[i].Key)
	}
	return nil
}

func (b *batchTxn) write(record []Record, et entry.EType) error {
	for _, r := range record {
		if r.K == nil {
			return ErrNilKey
		}

		e := &entry.Entry{
			Type:  et,
			Key:   r.K,
			Value: r.V,
			TTL:   entry.NewTTL(r.TTL),
			TxId:  b.txn.id.Int64(),
		}

		if e.Type == entry.DataEntryType {
			if r.TTL == 0 {
				e.TTL = 0
			} else if r.TTL < 0 {
				ttl, err := b.txn.ttl(r.K)
				if errors.Is(err, ErrKeyNotFound) {
					e.TTL = 0
				} else if err == nil {
					e.TTL = ttl
				} else {
					return err
				}
			}
		}
		b.pendingBuf[str.BytesToString(r.K)] = e
	}
	return nil
}

func (b *batchTxn) rollback() error {
	return b.txn.RollBack()
}

func (b *batchTxn) commit(needSync bool) error {
	es := make([]*entry.Entry, 0, len(b.pendingBuf))
	for _, e := range b.pendingBuf {
		es = append(es, e)
	}

	if b.ba.closed {
		return ErrBatchClosed
	}

	if err := b.putAll(es, needSync); err != nil {
		return err
	}

	if err := b.txn.Commit(); err != nil {
		return err
	}
	b.ba.effected.Add(int64(len(es)))
	return nil
}