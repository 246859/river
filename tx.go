package riverdb

import (
	"cmp"
	"encoding/binary"
	"github.com/246859/containers/heaps"
	"github.com/246859/river/entry"
	"github.com/246859/river/index"
	"github.com/246859/river/pkg/str"
	"github.com/246859/river/wal"
	"github.com/bwmarrin/snowflake"
	"github.com/google/btree"
	"github.com/pkg/errors"
	"regexp"
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

func newTx() (*tx, error) {
	nodeTs := time.Now().UnixNano() % (1 << 10)
	// nodeTs should be in [0, 1023]
	node, err := snowflake.NewNode(nodeTs)
	if err != nil {
		return nil, err
	}

	tx := &tx{
		node: node,
		active: heaps.NewBinaryHeap[*Txn](200, func(a, b *Txn) int {
			return cmp.Compare(a.startedTs, b.startedTs)
		}),
		committed: make([]*Txn, 0, 200),
	}

	tx.ts.Store(time.Now().UnixNano())

	return tx, nil
}

// tx represents transaction manager
type tx struct {
	db *DB

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
	if txn.pending.len() > 0 {
		db := txn.db
		var pendingEntry []*entry.Entry

		// collect pending entries
		txn.pending.iterate(func(item *entry.Entry) bool {
			item.TxId = txn.id.Int64()
			pendingEntry = append(pendingEntry, item)
			return true
		})

		// wrap transaction sequence
		pendingEntry = wrapTxnSequence(txn, pendingEntry)

		txn.committedTs = tx.newTs()
		// append to committed
		tx.committed = append(tx.committed, txn)

		db.mu.Lock()
		// do writes
		err := db.doWrites(pendingEntry)
		if err != nil {
			db.mu.Unlock()
			return err
		}
		db.mu.Unlock()

		// notify watcher
		if db.watcher != nil {
			for _, e := range pendingEntry {
				event := &Event{Value: e}
				switch e.Type {
				case entry.DataEntryType:
					event.Type = PutEvent
				case entry.DeletedEntryType:
					event.Type = DelEvent
				default:
					continue
				}
				db.watcher.push(event)
			}
		}
	}

	tx.discardTxn(txn, false)

	return nil
}

func (tx *tx) rollback(txn *Txn) {
	db := txn.db
	// notify watcher
	if db.watcher != nil && !txn.readonly {
		txn.pending.iterate(func(item *entry.Entry) bool {
			db.watcher.push(&Event{
				Type:  RollbackEvent,
				Value: item,
			})
			return true
		})
	}
	tx.discardTxn(txn, true)
}

func keyWithLen(length int) Key {
	bytes := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(bytes, int64(length))
	return bytes
}

func loadLenInKey(key Key) int {
	l, _ := binary.Varint(key)
	return int(l)
}

func wrapTxnSequence(txn *Txn, seq []*entry.Entry) []*entry.Entry {
	bytes := keyWithLen(len(seq))
	seq = append([]*entry.Entry{{TxId: txn.id.Int64(), Type: entry.TxnCommitEntryType, Key: bytes}}, seq...)
	seq = append(seq, &entry.Entry{TxId: txn.id.Int64(), Type: entry.TxnFinishedEntryType, Key: bytes})
	return seq
}

func newPendingTree(compare index.Compare) *btree.BTreeG[*entry.Entry] {
	return btree.NewG[*entry.Entry](32, func(a, b *entry.Entry) bool {
		return compare(a.Key, b.Key) == index.Less
	})
}

func newTxn(db *DB, readonly bool) *Txn {
	tx := &Txn{
		db:       db,
		readonly: readonly,
	}

	if !tx.readonly {
		tx.writes = make(map[uint64]struct{})
		tx.pending = &btreePending{
			tree: btree.NewG[*entry.Entry](32, func(a, b *entry.Entry) bool {
				return db.option.Compare(a.Key, b.Key) == index.Less
			}),
		}
	}

	return tx
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
	pending pendingWrites

	// ts
	startedTs   int64
	committedTs int64

	closed bool
}

// Begin begins a new transaction
func (db *DB) Begin(readonly bool) (*Txn, error) {
	if db.mask.CheckAny(closed) {
		return nil, ErrDBClosed
	}

	return db.tx.begin(db, readonly), nil
}

func (txn *Txn) Commit() error {
	if txn.closed {
		return ErrTxnClosed
	}

	if txn.db.mask.CheckAny(closed) {
		return ErrDBClosed
	}

	return txn.db.tx.commit(txn)
}

func (txn *Txn) RollBack() error {
	if txn.closed {
		return ErrTxnClosed
	}

	if txn.db.mask.CheckAny(closed) {
		return ErrDBClosed
	}

	txn.db.tx.rollback(txn)
	return nil
}

func (txn *Txn) Get(key Key) (Value, error) {
	var en entry.Entry
	if key == nil {
		return en.Value, ErrNilKey
	}

	if !txn.readonly {
		// if find it from pendingWrites
		if txnEn, has := txn.pending.get(key); has && !isExpiredOrDeleted(*txnEn) {
			return txnEn.Value, nil
		}
		// no need to track reading from pendingWrites
		// for it depend on pending data that impossible to be modified by other transactions
		txn.trackRead(key)
	}

	// io access
	value, err := txn.db.get(key)
	if err != nil {
		return en.Value, err
	}
	en = value

	return en.Value, nil
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

func (txn *Txn) ttl(key Key) (int64, error) {
	if key == nil {
		return -1, ErrNilKey
	}

	if !txn.readonly {
		// if find it from pendingWrites
		if txnEn, has := txn.pending.get(key); has && !isExpiredOrDeleted(*txnEn) {
			return txnEn.TTL, nil
		}
		// no need to track reading from pendingWrites
		// for it depend on pending data that impossible to be modified by other transactions
		txn.trackRead(key)
	}

	hint, err := txn.db.getHint(key)
	if err != nil {
		return -1, err
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

	return txn.put(en)
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
	return txn.put(en)
}

func (txn *Txn) Expire(key Key, ttl time.Duration) error {
	value, err := txn.Get(key)
	if err != nil {
		return err
	}
	return txn.Put(key, value, ttl)
}

func (txn *Txn) Range(opt RangeOptions, handler RangeHandler) error {
	it, err := txn.db.index.Iterator(opt)
	if err != nil {
		return err
	}

	compare := txn.db.option.Compare

	// store index range and pending snapshot
	snapshot := newPendingTree(txn.db.option.Compare)
	// copy the elements to snapshot
	err = txn.db.ranges(it, func(key Key) bool {
		snapshot.ReplaceOrInsert(&entry.Entry{Key: key})
		return true
	})

	if err != nil {
		return err
	}

	min, hasMin := snapshot.Min()
	max, hasMax := snapshot.Max()

	var rangeErr error

	itFn := func(item *entry.Entry) bool {
		if txn.closed {
			rangeErr = ErrTxnClosed
			return false
		}
		return handler(item.Key)
	}

	if txn.pending != nil {
		// combine pending iterator and index iterator to snapshot
		txn.pending.iterate(func(item *entry.Entry) bool {
			var put bool

			// greater than or equal to min
			if opt.Min != nil && hasMin && compare(min.Key, item.Key) >= index.Equal {
				put = true
			}

			// less than or equal to max
			if opt.Max != nil && hasMax && compare(max.Key, item.Key) <= index.Equal {
				put = true
			}

			// no range
			if opt.Min == nil && opt.Max == nil {
				put = true
			}

			// pattern match
			if opt.Pattern != nil {
				put = regexp.MustCompile(str.BytesToString(opt.Pattern)).Match(item.Key)
			}

			if put {
				snapshot.ReplaceOrInsert(item)
			}

			return true
		})
	}

	if opt.Descend {
		snapshot.Descend(itFn)
	} else {
		snapshot.Ascend(itFn)
	}

	return rangeErr
}

func (txn *Txn) put(en entry.Entry) error {
	if txn.readonly {
		return ErrTxnReadonly
	}

	if en.Key == nil {
		return ErrNilKey
	}

	if en.Type != entry.DataEntryType && en.Type != entry.DeletedEntryType {
		return entry.ErrRedundantEntryType
	}

	// check size
	enSize := int64(len(en.Key) + len(en.Value))
	if enSize > txn.db.option.MaxSize {
		return wal.ErrDataExceedFile
	}

	// write to pending data
	txn.pending.put(&en)
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
	txn.db = nil
	txn.pending = nil
	txn.closed = true
}

// in transaction, data will be written into pendingWriters.
// these data will be persisted only after committed
type pendingWrites interface {
	get(key Key) (*entry.Entry, bool)
	put(entry *entry.Entry)
	del(key Key)
	len() int
	iterate(f func(item *entry.Entry) bool)
}

// btree implements of pendingWrites
type btreePending struct {
	tree *btree.BTreeG[*entry.Entry]
}

func (p *btreePending) len() int {
	return p.tree.Len()
}

func (p *btreePending) get(key Key) (*entry.Entry, bool) {
	return p.tree.Get(&entry.Entry{Key: key})
}

func (p *btreePending) put(entry *entry.Entry) {
	p.tree.ReplaceOrInsert(entry)
}

func (p *btreePending) del(key Key) {
	p.tree.Delete(&entry.Entry{Key: key})
}

func (p *btreePending) iterate(f func(item *entry.Entry) bool) {
	p.tree.Ascend(func(item *entry.Entry) bool {
		return f(item)
	})
}
