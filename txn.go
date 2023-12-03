package riverdb

import (
	"cmp"
	"github.com/246859/containers/heaps"
	"github.com/246859/river/entry"
	"github.com/246859/river/index"
	"github.com/bwmarrin/snowflake"
	"github.com/pkg/errors"
	"io"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrTxnClosed   = errors.New("transaction is closed")
	ErrTxnReadonly = errors.New("transaction is read-only")
	ErrTxnConflict = errors.New("transaction is conflict")
)

func newTx(db *DB) (*tx, error) {
	nodeTs := time.Now().UnixNano() % (1 << 10)
	// nodeTs should be in [0, 1023]
	node, err := snowflake.NewNode(nodeTs)
	if err != nil {
		return nil, err
	}

	tx := &tx{
		db:   db,
		node: node,
		active: heaps.NewBinaryHeap[*Txn](200, func(a, b *Txn) int {
			return cmp.Compare(a.startedTs, b.startedTs)
		}),
		committed: make([]*Txn, 0, 200),
		pendingPool: sync.Pool{New: func() any {
			return &pendingWrite{
				memIndex: index.BtreeIndex(32, db.option.Compare)}
		}},
	}

	tx.ts.Store(time.Now().UnixNano())

	return tx, nil
}

// tx represents transaction manager
type tx struct {
	db *DB

	pendingPool sync.Pool

	ts atomic.Int64

	// snowflake id generator
	node *snowflake.Node

	// record of active transactions
	active heaps.Heap[*Txn]
	amu    sync.Mutex

	// record of committed txn
	committed []*Txn
	cmu       sync.Mutex
}

func (tx *tx) newTs() int64 {
	return tx.ts.Add(1)
}

func (tx *tx) generateID() snowflake.ID {
	return tx.node.Generate()
}

func (tx *tx) hasConflict(txn *Txn) bool {
	if len(txn.reads) == 0 {
		return false
	}

	for _, committedTxn := range tx.committed {
		// committed before cur txn started
		if committedTxn.committedTs <= txn.startedTs {
			continue
		}

		// check if read keys are modified by other txn after txn start
		for _, read := range txn.reads {
			_, exist := committedTxn.writes[read]
			if exist {
				return true
			}
		}
	}

	return false
}

// clean redundant transactions in tx.committed
// must be called in lock
func (tx *tx) cleanCommitted() {
	// the earliest active transaction
	early, has := tx.active.Peek()
	if has && early != nil {
		// remove committed transaction whose commitTs less than or equal to the earliest startTs
		tmp := tx.committed[:0]
		for _, txn := range tx.committed {
			if txn.committedTs > early.startedTs {
				tmp = append(tmp, txn)
			}
		}
		tx.committed = tmp
	}
}

func (tx *tx) begin(db *DB, readonly bool) *Txn {
	txn := newTxn(db, readonly)
	txn.id = tx.generateID()
	txn.startedTs = tx.newTs()

	tx.amu.Lock()
	tx.active.Push(txn)
	tx.amu.Unlock()

	return txn
}

func (tx *tx) discardTxn(txn *Txn, fail bool) {
	tx.amu.Lock()
	// remove from active
	i := slices.Index(tx.active.Values(), txn)
	if i > -1 {
		tx.active.Remove(i)
	}
	txn.discard(fail)
	tx.amu.Unlock()
}

func (tx *tx) commit(txn *Txn) error {

	if txn.readonly {
		tx.discardTxn(txn, false)
		return nil
	}

	tx.cmu.Lock()
	defer tx.cmu.Unlock()

	// conflict check
	if tx.hasConflict(txn) {
		return ErrTxnConflict
	}

	tx.cleanCommitted()

	// if txn have written data
	if txn.pending.Len() > 0 {
		db := txn.db

		// write a flag to mark this transaction is committed
		if err := txn.pending.Flag(entry.TxnCommitEntryType); err != nil {
			return err
		}

		// update index
		if err := txn.pending.CommitMemIndex(); err != nil {
			return err
		}

		txn.committedTs = tx.newTs()
		// append to committed
		tx.committed = append(tx.committed, txn)

		// notify watcher
		if db.watcher != nil {
			err := txn.pending.Iterate(RangeOptions{}, func(et entry.EType, hint index.Hint) error {
				event := &Event{Value: hint.Key}
				switch et {
				case entry.DataEntryType:
					if !db.watcher.expected(PutEvent) {
						return nil
					}
					event.Type = PutEvent
				case entry.DeletedEntryType:
					if !db.watcher.expected(DelEvent) {
						return nil
					}
					event.Type = DelEvent
				default:
					return nil
				}
				db.watcher.push(event)
				return nil
			})
			if err != nil {
				return err
			}
		}
	}

	tx.discardTxn(txn, false)

	return nil
}

func (tx *tx) rollback(txn *Txn) error {
	db := txn.db
	// notify watcher
	if !txn.readonly {
		if err := txn.pending.Flag(entry.TxnRollBackEntryType); err != nil {
			return err
		}
		if db.watcher != nil && db.watcher.expected(RollbackEvent) {
			err := txn.pending.Iterate(RangeOptions{}, func(et entry.EType, hint index.Hint) error {
				event := &Event{Value: hint.Key}
				switch et {
				case entry.DataEntryType:
					event.Type = PutEvent
				case entry.DeletedEntryType:
					event.Type = DelEvent
				default:
					return nil
				}
				db.watcher.push(event)
				return nil
			})
			if err != nil {
				return err
			}
		}
	}
	tx.discardTxn(txn, true)
	return nil
}

func newTxn(db *DB, readonly bool) *Txn {
	txn := &Txn{
		db:       db,
		readonly: readonly,
	}

	if !txn.readonly {
		txn.writes = make(map[uint64]struct{})
		pending := db.tx.pendingPool.Get().(*pendingWrite)
		pending.txn = txn
		pending.memIndex.Clear()
		txn.pending = pending
	}

	return txn
}

// Txn represents a transaction
type Txn struct {
	readonly bool
	id       snowflake.ID
	db       *DB

	// conflict tracking
	reads     []uint64
	writes    map[uint64]struct{}
	trackLock sync.Mutex

	// pending written data
	pending *pendingWrite

	// ts
	startedTs   int64
	committedTs int64

	closed bool
}

// Begin begins a new transaction
func (db *DB) Begin(readonly bool) (*Txn, error) {
	if db.flag.Check(closed) {
		return nil, ErrDBClosed
	}
	db.opmu.Lock()
	defer db.opmu.Unlock()

	return db.tx.begin(db, readonly), nil
}

// Commit once a transaction committed successfully, its affects will act on database forever.
// if crashes, db will reload transaction data from wal and update into memory index.
func (txn *Txn) Commit() error {
	if txn.closed {
		return ErrTxnClosed
	}

	if txn.db.flag.Check(closed) {
		return ErrDBClosed
	}

	return txn.db.tx.commit(txn)
}

// RollBack could promise consistence of database.
// if transaction has been committed successfully, it will return ErrTxnClosed and be ignored,
// otherwise RollBack will write a flag record to make sure data written in this transaction will never be seen,
// then discard this transaction.
func (txn *Txn) RollBack() error {
	if txn.closed {
		return ErrTxnClosed
	}

	if txn.db.flag.Check(closed) {
		return ErrDBClosed
	}

	return txn.db.tx.rollback(txn)
}

func (txn *Txn) Get(key Key) (Value, error) {
	if key == nil {
		return nil, ErrNilKey
	}

	// if not readonly, try to find it from pending-write
	if !txn.readonly {
		record, found, err := txn.pending.Get(key)
		if err != nil {
			return nil, err
		} else if found && isExpiredOrDeleted(*record) {
			return nil, ErrKeyNotFound
		} else if found {
			// no need to track reading from pendingWrites
			// because these pending data are impossible be modified by other transactions
			return record.Value, nil
		}
		txn.trackRead(key)
	}

	// read from data
	record, err := txn.db.get(key)
	if err != nil {
		return nil, err
	}

	return record.Value, nil
}

func (txn *Txn) TTL(key Key) (time.Duration, error) {
	ttl, err := txn.ttl(key)
	if err != nil {
		return 0, err
	}
	if ttl == 0 {
		return 0, nil
	}
	return entry.LeftTTl(ttl), nil
}

// ttl returns the TTL info of the given key without io accessing
func (txn *Txn) ttl(key Key) (int64, error) {
	if key == nil {
		return 0, ErrNilKey
	}

	if !txn.readonly {
		pendingHint, has := txn.pending.memIndex.Get(key)
		if has {
			if isExpiredOrDeleted(entry.Entry{Type: pendingHint.Meta.(entry.EType), TTL: pendingHint.TTL}) {
				return 0, ErrKeyNotFound
			} else {
				return pendingHint.TTL, nil
			}
		}
		txn.trackRead(key)
	}

	// read from index
	hint, err := txn.db.getHint(key)
	if err != nil {
		return 0, err
	}

	if hint.TTL <= 0 {
		return 0, nil
	} else {
		return hint.TTL, nil
	}
}

func (txn *Txn) Put(key Key, value Value, ttl time.Duration) error {
	en := entry.Entry{
		Type:  entry.DataEntryType,
		Key:   key,
		Value: value,
		TTL:   entry.NewTTL(ttl),
	}

	// if tll > 0, update the ttl
	// if ttl == 0, persistent
	// if ttl < 0, apply old ttl
	if ttl == 0 {
		en.TTL = 0
	} else if ttl < 0 {
		ttl, err := txn.ttl(key)
		if errors.Is(err, ErrKeyNotFound) {
			en.TTL = 0
		} else if err != nil {
			return err
		} else {
			en.TTL = ttl
		}
	}

	return txn.put(&en)
}

func (txn *Txn) Del(key Key) error {
	// check if key exists
	_, err := txn.TTL(key)
	if errors.Is(err, ErrKeyNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	en := entry.Entry{
		Type: entry.DeletedEntryType,
		Key:  key,
	}
	return txn.put(&en)
}

func (txn *Txn) Expire(key Key, ttl time.Duration) error {
	value, err := txn.Get(key)
	if err != nil {
		return err
	}
	return txn.Put(key, value, ttl)
}

func (txn *Txn) Range(opt RangeOptions, handler RangeHandler) error {
	it, err := txn.db.idxIterator(opt)
	if err != nil {
		return err
	}

	snapshot := index.BtreeNIndex(32, txn.db.option.Compare)
	defer snapshot.Close()
	err = index.Ranges(it, func(hint index.Hint) error {
		if entry.IsExpired(hint.TTL) {
			return nil
		}
		return snapshot.Put(hint)
	})
	if err != nil {
		return err
	}

	if !txn.readonly {
		err = txn.pending.Iterate(opt, func(et entry.EType, hint index.Hint) error {
			switch et {
			case entry.DataEntryType:
				if !entry.IsExpired(hint.TTL) {
					return snapshot.Put(hint)
				}
			case entry.DeletedEntryType:
				_ = snapshot.Del(hint.Key)
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	sit, err := snapshot.Iterator(opt)

	return index.Ranges(sit, func(hint index.Hint) error {
		if !handler(hint.Key) {
			return io.EOF
		}
		return nil
	})
}

func (txn *Txn) put(en *entry.Entry) error {
	if txn.readonly {
		return ErrTxnReadonly
	}

	if en.Key == nil {
		return ErrNilKey
	}

	// write to pending data
	if err := txn.pending.Put(en); err != nil {
		return err
	}
	// track write
	txn.trackWrite(en.Key)

	return nil
}

func (txn *Txn) trackRead(key Key) {
	if !txn.readonly {
		hash := memHash(key)
		txn.trackLock.Lock()
		txn.reads = append(txn.reads, hash)
		txn.trackLock.Unlock()
	}
}

func (txn *Txn) trackWrite(key Key) {
	if !txn.readonly {
		hash := memHash(key)
		txn.writes[hash] = struct{}{}
	}
}
func (txn *Txn) discard(fail bool) {
	// must be called in lock
	if fail {
		txn.reads = nil
		txn.writes = nil
	}
	if !txn.readonly {
		txn.db.tx.pendingPool.Put(txn.pending)
	}
	txn.db = nil
	txn.closed = true
}

// pendingWrite maintains a temporary index, whose behavior same as db index mostly.
// all written entries in transactions, their hints will be updated into the pending index rather than db index
// only transaction is committed successfully, then pending index will be written to db index.
// if transaction has been rolled back, these pending index will be cleared for re-used by other transactions.
type pendingWrite struct {
	txn      *Txn
	memIndex index.Index
}

func (p *pendingWrite) Len() int {
	return p.memIndex.Size()
}

// Get returns the entry from pending index
func (p *pendingWrite) Get(key Key) (*entry.Entry, bool, error) {
	db := p.txn.db
	hint, has := p.memIndex.Get(key)
	if !has || hint.Meta.(entry.EType) != entry.DataEntryType {
		return nil, false, nil
	}
	record, err := db.read(hint.ChunkPos)
	if err != nil {
		return nil, false, err
	}
	// txid must be same as txn
	if record.TxId != p.txn.id.Int64() {
		return nil, false, nil
	}
	return record, true, nil
}

// Put writes a data entry to database, and record of hint in pending index
func (p *pendingWrite) Put(record *entry.Entry) error {
	db := p.txn.db
	record.TxId = p.txn.id.Int64()
	pos, err := db.write(record)
	if err != nil {
		return err
	}
	return p.memIndex.Put(index.Hint{Key: record.Key, TTL: record.TTL, ChunkPos: pos, Meta: record.Type})
}

// Flag writes an empty k-v entry which type is not normal data type
func (p *pendingWrite) Flag(t entry.EType) error {
	return p.Put(&entry.Entry{Key: []byte{}, Type: t})
}

// Iterate iterates over all entries in pending index in way of that same as index range
func (p *pendingWrite) Iterate(options RangeOptions, handle func(et entry.EType, hint index.Hint) error) error {
	it, err := p.memIndex.Iterator(options)
	if err != nil {
		return err
	}
	return index.Ranges(it, func(hint index.Hint) error {
		return handle(hint.Meta.(entry.EType), hint)
	})
}

// CommitMemIndex commits the pending write to the db index
func (p *pendingWrite) CommitMemIndex() error {
	var ws []writeIndex
	err := p.Iterate(RangeOptions{}, func(et entry.EType, hint index.Hint) error {
		ws = append(ws, writeIndex{t: et, h: hint})
		return nil
	})
	if err != nil {
		return err
	}

	// write to index
	if err := p.txn.db.writeIndex(ws...); err != nil {
		return err
	}
	return nil
}
