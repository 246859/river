package riverdb

import (
	"github.com/246859/river/entry"
	"github.com/246859/river/index"
	"github.com/246859/river/wal"
	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"io"
	"path"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrKeyNotFound    = errors.New("key not found")
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

// Open returns a river db database
func Open(options Options, opts ...Option) (*DB, error) {
	// apply options
	for _, opt := range opts {
		opt(&options)
	}

	// check options
	err := options.check()
	if err != nil {
		return nil, err
	}

	// try to lock dir
	fileLockPath := path.Join(options.Dir, lockName)
	filelock := flock.New(fileLockPath)
	success, err := filelock.TryRLock()
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, ErrDirAlreadyUsed
	}

	db := &DB{
		fu:         filelock,
		serializer: entry.BinaryEntry{},
		option:     options,
	}

	// initialize memory index
	db.index = index.BtreeIndex(48, options.Compare)

	// initialize transaction manager
	tx, err := newTx()
	if err != nil {
		return nil, err
	}
	db.tx = tx

	// load data from dir
	data, err := db.loadData()
	if err != nil {
		return nil, err
	}
	db.data = data

	// load index
	memIndex, err := db.loadIndex()
	if err != nil {
		return nil, err
	}
	db.index = memIndex

	return db, nil
}

// DB represents a db instance, which stores wal file in a specific data directory
type DB struct {
	data  *wal.Wal
	hint  *wal.Wal
	index index.Index

	// transaction manager
	tx *tx

	// db status
	closed  atomic.Bool // db is closed
	merging atomic.Bool // db is merging

	// read-write lock
	mu sync.Mutex
	fu *flock.Flock

	serializer entry.Serializer
	option     Options
}

// Get returns value match the given key
func (db *DB) Get(key Key) (Value, error) {
	tx, err := db.Begin(true)
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
// puts a key-value into db. overwrite value if key already exists.
// if ttl <= 0, the key will never expired
func (db *DB) Put(key Key, value Value, ttl time.Duration) error {
	tx, err := db.Begin(false)
	if err != nil {
		return err
	}
	err = tx.Put(key, value, ttl)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Del remove the value match the give key from db
func (db *DB) Del(key Key) error {
	tx, err := db.Begin(false)
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
	if err != nil {
		return 0, err
	}
	ttl, err := tx.TTL(key)
	if err != nil {
		return ttl, err
	}
	return ttl, tx.Commit()
}

// Range iterates over all the keys that match the given RangeOption
// and call handler for each key-value
func (db *DB) Range(option index.RangeOption, handler RangeHandler) error {
	tx, err := db.Begin(true)
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
	bytes, err := db.data.Read(hintToWalPos(hint.Hint))

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

	if entry.IsExpired(hint.Hint.TTL) {
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
		if out {
			return it.Close()
		}

		if entry.IsExpired(hint.Hint.TTL) {
			continue
		}

		if !handler(hint.Key) {
			return nil
		}
	}
}

// Sync syncs written buffer to disk
func (db *DB) Sync() error {
	if db.closed.Load() {
		return ErrDBClosed
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	return db.data.Sync()
}

// Close closes db
func (db *DB) Close() error {
	if db.closed.Load() {
		return ErrDBClosed
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	// close the files
	err := db.data.Close()
	if err != nil {
		return err
	}
	if db.hint != nil {
		err := db.hint.Close()
		if err != nil {
			return err
		}
	}

	// close index
	if err = db.index.Close(); err != nil {
		return err
	}

	// release the file lock
	if err := db.fu.Unlock(); err != nil {
		return err
	}
	if err = db.fu.Close(); err != nil {
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
	db.data = nil
	db.hint = nil
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
			Key:  entries[i].Key,
			Hint: walPosToHint(pos),
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

func (db *DB) loadIndex() (index.Index, error) {
	return nil, nil
}

func (db *DB) loadData() (*wal.Wal, error) {
	options := db.option
	// wal data
	dataDir := path.Join(options.Dir, dataName)
	datawal, err := wal.Open(wal.Option{
		DataDir:        dataDir,
		MaxFileSize:    options.MaxSize,
		Ext:            dataName,
		BlockCache:     options.BlockCache,
		FsyncPerWrite:  options.Fsync,
		FsyncThreshold: options.FsyncThreshold,
	})

	if err != nil {
		return nil, err
	}
	return datawal, nil
}

func (db *DB) loadIndexFromHint() error {
	return nil
}

// load index from data files
func (db *DB) loadIndexFromData() error {
	var (
		data         = db.data
		memIndex     = db.index
		txnSequences = make(map[int64][]entry.EntryHint, 1<<10)
	)

	// get the wal iterator
	it, err := data.Iterator(0, wal.ChunkPos{})
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

		// process record data
		switch record.Type {
		// transaction commit flag
		case entry.TxnCommitEntryType:
			expectedLen := loadLenInKey(record.Key)
			txnSequences[record.TxId] = make([]entry.EntryHint, 0, expectedLen)
		// normal data flag
		case entry.DataEntryType, entry.DeletedEntryType:
			txnSequences[record.TxId] = append(txnSequences[record.TxId], entry.EntryHint{Entry: record, Hint: walPosToHint(pos)})
		// transaction finished flag
		case entry.TxnFinishedEntryType:
			seqs := txnSequences[record.TxId]
			for _, en := range seqs {
				var idxErr error

				switch en.Type {
				case entry.DataEntryType:
					// skip if is expired
					if entry.IsExpired(en.Entry.TTL) {
						continue
					}
					idxErr = memIndex.Put(index.Hint{Key: en.Key, Hint: en.Hint})
				case entry.DeletedEntryType:
					_, idxErr = memIndex.Del(en.Key)
				}

				if idxErr != nil {
					return idxErr
				}
			}
		}

	}
}
