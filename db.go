package riverdb

import (
	"github.com/246859/river/entry"
	"github.com/246859/river/file"
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
	ErrDBMerging      = errors.New("db is merging data")
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

// Open returns a river db database
func Open(options Options, opts ...Option) (*DB, error) {
	// apply options
	for _, opt := range opts {
		opt(&options)
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

	// only one process can be active at dir
	if err := db.lockDir(); err != nil {
		return nil, err
	}

	// check if is need to transfer merged data
	db.mergeOp = &mergeOP{db: db}
	if err := db.mergeOp.doTransfer(); err != nil {
		return nil, err
	}

	// load data from dir
	if err := db.loadData(); err != nil {
		return nil, err
	}

	// load index
	if err := db.loadIndex(); err != nil {
		return nil, err
	}

	// initialize transaction manager
	db.tx, err = newTx()
	if err != nil {
		return nil, err
	}

	return db, nil
}

const (
	// db is merging data
	dbMerging uint8 = 1 << iota
	// db is already closed
	dbClosed
)

// DB represents a db instance, which stores wal file in a specific data directory
type DB struct {
	data  *wal.Wal
	index index.Index

	mergeOp *mergeOP

	// transaction manager
	tx *tx

	flag uint8

	// read-write lock
	mu sync.Mutex
	fu *flock.Flock

	serializer entry.Serializer
	option     Options
}

// Get returns value match the given key
func (db *DB) Get(key Key) (Value, error) {
	tx, err := db.Begin(true)
	defer func() {
		if err != nil {
			_ = tx.RollBack()
		}
	}()

	if err != nil {
		return nil, err
	}
	value, err := tx.Get(key)
	if err != nil {
		return value, err
	}
	return value, tx.Commit()
}

// Put
// puts a key-value pair into db, overwrite value if key already exists.
// nil key is invalid, but nil value is allowed.
// if tll == 0, key will be persisted, or ttl < 0, key will apply the previous ttl.
func (db *DB) Put(key Key, value Value, ttl time.Duration) error {
	tx, err := db.Begin(false)
	defer func() {
		if err != nil {
			_ = tx.RollBack()
		}
	}()

	if err != nil {
		return err
	}
	err = tx.Put(key, value, ttl)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Del remove the key-value pair match the give key from db.
// it will return nil if key not exist
func (db *DB) Del(key Key) error {
	tx, err := db.Begin(false)
	defer func() {
		if err != nil {
			_ = tx.RollBack()
		}
	}()

	if err != nil {
		return err
	}
	err = tx.Del(key)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Expire update ttl of the specified key
// if ttl <= 0, the key will never expired
func (db *DB) Expire(key Key, ttl time.Duration) error {
	tx, err := db.Begin(false)
	defer func() {
		if err != nil {
			_ = tx.RollBack()
		}
	}()

	if err != nil {
		return err
	}
	err = tx.Expire(key, ttl)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// TTL returns left live time of the specified key
func (db *DB) TTL(key Key) (time.Duration, error) {
	tx, err := db.Begin(true)
	defer func() {
		if err != nil {
			_ = tx.RollBack()
		}
	}()

	if err != nil {
		return 0, err
	}
	ttl, err := tx.TTL(key)
	if err != nil {
		return ttl, err
	}
	return ttl, tx.Commit()
}

type RangeOptions = index.RangeOption

// Range iterates over all the keys that match the given RangeOption
// and call handler for each key-value
func (db *DB) Range(option RangeOptions, handler RangeHandler) error {
	tx, err := db.Begin(true)
	defer func() {
		if err != nil {
			_ = tx.RollBack()
		}
	}()

	if err != nil {
		return err
	}
	err = tx.Range(option, handler)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) get(key Key) (entry.Entry, error) {
	// check index
	hint, err := db.getHint(key)
	if err != nil {
		return entry.Entry{}, err
	}

	// read raw data from wal
	bytes, err := db.data.Read(hint.ChunkPos)

	if err != nil {
		return entry.Entry{}, err
	}

	data, err := db.serializer.UnMarshalEntry(bytes)
	if err != nil {
		return entry.Entry{}, err
	}

	return data, nil
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

func (db *DB) ranges(it index.Iterator, handler RangeHandler) error {
	if it == nil {
		return index.ErrNilIterator
	}
	for {
		hint, out := it.Next()

		if entry.IsExpired(hint.TTL) {
			continue
		}

		if !handler(hint.Key) {
			return nil
		}

		if out {
			return it.Close()
		}
	}
}

// Purge remove all entries from data wal
func (db *DB) Purge() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	err := db.data.Purge()
	if err != nil {
		return err
	}
	if err := db.mergeOp.clean(); err != nil {
		return err
	}

	if err = db.loadData(); err != nil {
		return err
	}
	return nil
}

// Sync syncs written buffer to disk
func (db *DB) Sync() error {
	if db.flag&dbClosed != 0 {
		return ErrDBClosed
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	return db.data.Sync()
}

func (db *DB) closeWal() error {
	// close the files
	err := db.data.Close()
	if err != nil {
		return err
	}
	return nil
}

// Close closes db
func (db *DB) Close() error {
	if db.flag&dbClosed != 0 {
		return ErrDBClosed
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	err := db.closeWal()
	if err != nil {
		return err
	}

	// close index
	if err := db.index.Close(); err != nil {
		return err
	}

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

// discard db fields for convenient to gc
// must be called in lock
func (db *DB) discard() {
	db.flag |= dbClosed
	db.data = nil
	db.index = nil
	db.tx = nil
	db.fu = nil
	db.serializer = nil
}

// doWrites write all the entries to db, then updates index in mem
func (db *DB) doWrites(entries []*entry.Entry) error {
	if entries == nil || len(entries) == 0 {
		return nil
	}

	// pending writes
	for _, ee := range entries {
		bytes, err := db.serializer.MarshalEntry(*ee)
		if err != nil {
			return err
		}
		db.data.PendingWrite(bytes)
	}

	// pending push, write data to disk
	chunkPos, err := db.data.PendingPush()
	if err != nil {
		return err
	}

	// update index
	for i, pos := range chunkPos {
		en := index.Hint{
			Key:      entries[i].Key,
			TTL:      entries[i].TTL,
			ChunkPos: pos,
		}

		var err error

		switch entries[i].Type {
		case entry.DataEntryType:
			err = db.index.Put(en)
		case entry.DeletedEntryType:
			_, err = db.index.Del(en.Key)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) loadIndex() error {

	db.index = index.BtreeIndex(64, db.option.Compare)

	// check if data dir contains merged finished-file
	fwal, err := openFinished(db.option.dataDir)
	if err != nil {
		return err
	}

	minFid, maxFid := uint32(0), db.data.ActiveFid()
	// lastFid means that the last immutable wal file id before last merge operation. we can load part of index from hint,
	// files after the lastFid maybe has been written new data, so we need to load it from data
	lastFid, err := hasFinished(fwal)
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
	err = db.loadIndexFromData(minFid, maxFid)
	if err != nil {
		return err
	}

	return nil
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

type entryhint struct {
	entry entry.Entry
	pos   wal.ChunkPos
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
		// transaction commit flag
		case entry.TxnCommitEntryType:
			expectedLen := loadLenInKey(record.Key)
			txnSequences[record.TxId] = make([]entryhint, 0, expectedLen)
		// normal data flag
		case entry.DataEntryType, entry.DeletedEntryType:
			txnSequences[record.TxId] = append(txnSequences[record.TxId], entryhint{entry: record, pos: pos})
		// transaction finished flag
		case entry.TxnFinishedEntryType:
			seqs := txnSequences[record.TxId]
			for _, en := range seqs {
				var idxErr error

				switch en.entry.Type {
				case entry.DataEntryType:
					// skip if is expired
					if entry.IsExpired(en.entry.TTL) {
						continue
					}
					idxErr = memIndex.Put(index.Hint{Key: en.entry.Key, TTL: en.entry.TTL, ChunkPos: pos})
				case entry.DeletedEntryType:
					_, idxErr = memIndex.Del(en.entry.Key)
				}

				if idxErr != nil {
					return idxErr
				}
			}
		}
	}
}

func (db *DB) lockDir() error {
	if db.flag&dbClosed != 0 {
		return ErrDBClosed
	}

	if db.fu == nil {
		lockpath := db.option.filelock
		_, err := file.OpenStdFile(lockpath, os.O_CREATE, 0644)
		if err != nil {
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
	if db.flag&dbClosed != 0 {
		return ErrDBClosed
	}

	return db.fu.Close()
}
