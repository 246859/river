package riverdb

import (
	"context"
	"github.com/246859/river/entry"
	"github.com/246859/river/index"
	"github.com/246859/river/wal"
	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	ErrKeyNotFound    = errors.New("key not found")
	ErrNilKey         = entry.ErrNilKey
	ErrDBClosed       = errors.New("db is already closed")
	ErrDirAlreadyUsed = errors.New("dir already used by another process")
)

type (
	// Key db key type
	Key = []byte
	// Value db value type
	Value = []byte

	// RangeHandler iterate over key-value in db
	RangeHandler = func(key Key) bool
)

// Open returns a river db instance
func Open(options Options, opts ...Option) (*DB, error) {
	return OpenWithCtx(context.Background(), options, opts...)
}

// OpenWithCtx returns a river db instance with context
func OpenWithCtx(ctx context.Context, options Options, opts ...Option) (*DB, error) {
	// apply options
	for _, opt := range opts {
		opt(&options)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// check options
	opt, err := revise(options)
	if err != nil {
		return nil, err
	}

	db := &DB{
		serializer: entry.BinaryEntry{},
		option:     opt,
	}

	// context
	db.ctx, db.cancel = context.WithCancel(ctx)

	// only one process can be active at dir
	if err := db.lockDir(); err != nil {
		return nil, err
	}

	// check if is need to transfer merged data
	db.mergeOp = &mergeOP{db: db}
	if err := db.mergeOp.doTransfer(); err != nil {
		return nil, err
	}

	// load db data
	if err = db.load(); err != nil {
		return nil, err
	}

	// initialize transaction manager
	db.tx, err = newTx(db)
	if err != nil {
		return nil, err
	}

	if db.option.WatchSize > 0 {
		db.watcher = newWatcher(db.option.WatchSize, db.option.WatchEvents...)
		db.watcher.clear()
		// new goroutine to watching events
		go db.watcher.watch(db.ctx)
	}

	return db, nil
}

const (
	backing uint64 = 1 << iota
	recovering
	merging
	closed
)

// DB represents a db instance, which stores wal file in a specific data directory
type DB struct {
	ctx    context.Context
	cancel context.CancelFunc

	// wal files
	data   *wal.Wal
	hint   *wal.Wal
	finish *wal.Wal
	// mem index
	index index.Index
	// mu ensures that only one goroutine can update data or index at a time
	mu sync.RWMutex

	// event watch
	watcher *watcher

	// merge operator
	mergeOp *mergeOP

	// opmu has the responsibility of ensuring that all db operations must be synchronized
	// like backup operations, merge operations
	opmu sync.Mutex

	// transaction manager
	tx *tx

	flag BitFlag

	// db file lock
	fu *flock.Flock

	serializer entry.Serializer
	option     Options

	// how many record has been written
	numOfRecord int64
}

// Get returns value match the given key, if it expired or not found
// db will return ErrKeyNotFound. nil Key is not allowed.
func (db *DB) Get(key Key) (Value, error) {
	txn, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	// you'd better call Rollback after transaction begins
	defer txn.RollBack()
	value, err := txn.Get(key)
	if err != nil {
		return value, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return value, nil
}

// Put
// puts a key-value pair into db, overwrite value if key already exists.
// nil key is invalid, but nil value is allowed, it will be overwritten to empty []byte.
// if tll == 0, key will be persisted, or ttl < 0, key will apply the previous ttl.
func (db *DB) Put(key Key, value Value, ttl time.Duration) error {
	txn, err := db.Begin(false)
	if err != nil {
		return err
	}
	defer txn.RollBack()
	if err := txn.Put(key, value, ttl); err != nil {
		return err
	}
	if err := txn.Commit(); err != nil {
		return err
	}
	return nil
}

// Del remove the key-value pair match the give key from db.
// it will return nil if key not exist
func (db *DB) Del(key Key) error {
	txn, err := db.Begin(false)
	if err != nil {
		return err
	}
	defer txn.RollBack()
	if err := txn.Del(key); err != nil {
		return err
	}
	if err := txn.Commit(); err != nil {
		return err
	}
	return nil
}

// Expire update ttl of the specified key
// if ttl <= 0, the key will never expired
func (db *DB) Expire(key Key, ttl time.Duration) error {
	txn, err := db.Begin(false)
	if err != nil {
		return err
	}
	defer txn.RollBack()
	if err := txn.Expire(key, ttl); err != nil {
		return err
	}
	if err := txn.Commit(); err != nil {
		return err
	}
	return nil
}

// TTL returns left live time of the specified key
func (db *DB) TTL(key Key) (time.Duration, error) {
	txn, err := db.Begin(true)
	if err != nil {
		return 0, err
	}
	defer txn.RollBack()
	ttl, err := txn.TTL(key)
	if err != nil {
		return ttl, err
	}
	if err := txn.Commit(); err != nil {
		return 0, err
	}
	return ttl, nil
}

// RangeOptions is alias of index.RangeOption
type RangeOptions = index.RangeOption

// Range iterates over all the keys that match the given RangeOption
// and call handler for each key-value
func (db *DB) Range(option RangeOptions, handler RangeHandler) error {
	txn, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer txn.RollBack()
	if err := txn.Range(option, handler); err != nil {
		return err
	}
	if err := txn.Commit(); err != nil {
		return err
	}
	return nil
}

// Purge remove all entries from data wal
func (db *DB) Purge() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	purges := []interface {
		Purge() error
	}{
		db.data,
		db.hint,
		db.finish,
	}

	// purge wal files
	for _, purge := range purges {
		if purge == (*wal.Wal)(nil) {
			continue
		}
		if err := purge.Purge(); err != nil {
			return err
		}
	}

	// clear merge data
	if err := db.mergeOp.clean(); err != nil {
		return err
	}

	// clear index
	db.index.Clear()
	db.numOfRecord = 0

	return nil
}

// Sync syncs written buffer to disk
func (db *DB) Sync() error {
	if db.flag.Check(closed) {
		return ErrDBClosed
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	return db.data.Sync()
}

func (db *DB) closeWal() error {
	closes := []io.Closer{
		db.index,
		db.hint,
		db.finish,
		db.data,
	}

	for _, closer := range closes {
		if closer == (*wal.Wal)(nil) {
			continue
		}
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Close closes db
// Once the db is closed, it can no longer be used
func (db *DB) Close() error {
	if db.flag.Check(closed) {
		return ErrDBClosed
	}

	db.mu.Lock()
	defer db.mu.Unlock()
	db.cancel()

	err := db.closeWal()
	if err != nil {
		return err
	}

	if err = db.mergeOp.Close(); err != nil {
		return err
	}

	db.watcher.close()

	// release dir lock
	if err := db.unlockDir(); err != nil {
		return err
	}

	// discard db
	db.discard()

	// manually gc
	if db.option.ClosedGc {
		runtime.GC()
	}

	return nil
}

func (db *DB) lockDir() error {

	if db.fu == nil {
		lockpath := db.option.filelock
		if err := os.MkdirAll(db.option.Dir, 0755); err != nil {
			return err
		}
		fl, err := os.OpenFile(lockpath, os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		if err = fl.Close(); err != nil {
			return err
		}
		db.fu = flock.New(lockpath)
	}

	// try to lock dir
	lock, err := db.fu.TryLock()
	if err != nil {
		return err
	}

	if !lock {
		return ErrDirAlreadyUsed
	}
	return nil
}

func (db *DB) unlockDir() error {

	if err := db.fu.Unlock(); err != nil {
		return err
	}

	if err := db.fu.Close(); err != nil {
		return err
	}

	return nil
}

// discard db fields for convenient to gc
// must be called in lock
func (db *DB) discard() {
	db.flag.Store(closed)

	db.ctx = nil
	db.cancel = nil
	db.data = nil
	db.hint = nil
	db.finish = nil
	db.index = nil
	db.watcher = nil
	db.mergeOp = nil
	db.tx = nil
	db.fu = nil
	db.serializer = nil
}

type entryhint struct {
	entry entry.Entry
	pos   wal.ChunkPos
}

func (db *DB) load() error {

	// load hint wal
	if err := db.loadhint(); err != nil {
		return err
	}

	// load finished wal
	if err := db.loadfinish(); err != nil {
		return err
	}

	// load data from dir
	if err := db.loadData(); err != nil {
		return err
	}

	// load index
	if err := db.loadIndex(); err != nil {
		return err
	}

	return nil
}

func (db *DB) loadhint() error {
	hint, err := openHint(db.option.dataDir)
	if err != nil {
		return err
	}
	db.hint = hint
	return nil
}

func (db *DB) loadfinish() error {
	finish, err := openFinished(db.option.dataDir)
	if err != nil {
		return err
	}
	db.finish = finish
	return nil
}

func (db *DB) loadIndex() error {

	db.index = index.BtreeIndex(64, db.option.Compare)

	minFid, maxFid := uint32(0), db.data.ActiveFid()
	// lastFid means that the last immutable wal file id before last merge operation.
	// we can load part of index from hint which is generated in last merge operation to speed up index loading,
	// this part of hint only stores the data without actual value.
	// files after the lastFid maybe has been written new data, so we need to load it from data wal.
	lastFid, err := hasFinished(db.finish)
	if errors.Is(err, ErrMergedNotFinished) {
		minFid = 0
	} else if err != nil {
		return err
	} else {
		minFid = lastFid + 1
	}

	// if data dir has finished-file, db can load index from hint without load whole data file
	if err := db.loadIndexFromHint(); err != nil {
		return err
	}

	// we can skip the file before lastFid, their index maybe load from hint
	if err := db.loadIndexFromData(minFid, maxFid); err != nil {
		return err
	}

	return nil
}

// load index from data files
func (db *DB) loadIndexFromData(minFid, maxFid uint32) error {
	var (
		data         = db.data
		memIndex     = db.index
		txnSequences = make(map[int64][]entryhint, 1<<10)
	)

	// get the wal iterator
	it, err := data.Iterator(minFid, maxFid, wal.ChunkPos{})
	if err != nil {
		return err
	}

	// iterate over all the entries, and update index in memory
	for {
		rawData, pos, err := it.NextData()
		if err != nil {
			// all files were read finished
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return err
		}

		// unmarshal the raw data
		record, err := db.serializer.UnMarshalEntry(rawData)
		if err != nil {
			return err
		}

		// if is merged data, update index directly
		if record.TxId == mergedTxnId && record.Type == entry.DataEntryType && !entry.IsExpired(record.TTL) {
			err := memIndex.Put(index.Hint{Key: record.Key, TTL: record.TTL, ChunkPos: pos})
			if err != nil {
				return err
			}
			continue
		}

		// process record data
		switch record.Type {
		// normal data flag
		case entry.DataEntryType, entry.DeletedEntryType:
			txnSequences[record.TxId] = append(txnSequences[record.TxId], entryhint{entry: record, pos: pos})
		// transaction finished flag
		case entry.TxnCommitEntryType:
			seqs := txnSequences[record.TxId]
			for _, en := range seqs {
				var idxErr error

				switch en.entry.Type {
				case entry.DataEntryType:
					// skip if is expired
					if entry.IsExpired(en.entry.TTL) {
						continue
					}
					idxErr = memIndex.Put(index.Hint{Key: en.entry.Key, TTL: en.entry.TTL, ChunkPos: en.pos})
				case entry.DeletedEntryType:
					_, idxErr = memIndex.Del(en.entry.Key)
				}

				if idxErr != nil {
					return idxErr
				}
			}

			// release mem
			txnSequences[record.TxId] = nil
		}
		db.numOfRecord++
	}
}

func (db *DB) loadData() error {
	options := db.option
	// wal data
	datawal, err := wal.Open(wal.Option{
		DataDir:        options.dataDir,
		MaxFileSize:    options.MaxSize,
		Ext:            dataName,
		BlockCache:     options.BlockCache,
		FsyncPerWrite:  options.Fsync,
		FsyncThreshold: options.FsyncThreshold,
	})

	if err != nil {
		return err
	}
	db.data = datawal
	return nil
}

func (db *DB) get(key Key) (*entry.Entry, error) {
	// check index
	hint, err := db.getHint(key)
	if err != nil {
		return nil, err
	}

	return db.read(hint.ChunkPos)
}

func (db *DB) getHint(key Key) (index.Hint, error) {
	hint, exist := db.index.Get(key)
	if !exist {
		return hint, ErrKeyNotFound
	}

	if entry.IsExpired(hint.TTL) {
		return hint, ErrKeyNotFound
	}

	return hint, nil
}

func (db *DB) read(pos wal.ChunkPos) (*entry.Entry, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	raw, err := db.data.Read(pos)
	if err != nil {
		return nil, err
	}
	record, err := db.serializer.UnMarshalEntry(raw)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (db *DB) write(entry *entry.Entry) (wal.ChunkPos, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	raw, err := db.serializer.MarshalEntry(*entry)
	if err != nil {
		return wal.ChunkPos{}, err
	}
	pos, err := db.data.Write(raw)
	if err != nil {
		return wal.ChunkPos{}, err
	}
	db.numOfRecord++
	return pos, nil
}

func (db *DB) writeAll(entries []*entry.Entry, needSync bool) ([]wal.ChunkPos, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	rawdatas := make([][]byte, 0, len(entries))

	for _, en := range entries {
		bytes, err := db.serializer.MarshalEntry(*en)
		if err != nil {
			return nil, err
		}
		rawdatas = append(rawdatas, bytes)
	}

	all, err := db.data.WriteAll(rawdatas, needSync)
	if err != nil {
		return nil, err
	}
	return all, nil
}
