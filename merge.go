package riverdb

import (
	"encoding/binary"
	"fmt"
	"github.com/246859/river/entry"
	"github.com/246859/river/index"
	"github.com/246859/river/types"
	"github.com/246859/river/wal"
	"github.com/pkg/errors"
	"io"
	"math"
	"os"
	"path"
	"strings"
	"time"
)

var (
	ErrMergedNotFinished = errors.New("merged not finished")
)

const (
	mergedTxnId = -1
)

// Merge clean the redundant data entry in db, shrinking the db size
// if domerge is false, it will only record of merged data, will not replace them to data dir
func (db *DB) Merge(domerge bool) error {

	if db.flag.Check(closed) {
		return ErrDBClosed
	}

	db.opmu.Lock()
	defer db.opmu.Unlock()

	db.flag.Store(merging)
	defer db.flag.Revoke(merging)

	db.mu.Lock()

	// return if db data is empty
	if db.data.IsEmpty() {
		db.mu.Unlock()
		return nil
	}

	// record last active id before rotate
	lastActiveId := db.data.ActiveFid()

	// rotate open a new wal file to write new data
	if err := db.data.Rotate(); err != nil {
		db.mu.Unlock()
		return err
	}

	// then read old data from immutable file, due to no possibility of writing conflicts occurring
	// so no need to lock db
	db.mu.Unlock()

	// reload op
	op := db.mergeOp
	err := op.reload()
	if err != nil {
		return err
	}

	// do real merge
	if err := op.doMerge(lastActiveId); err != nil {
		_ = op.Close()
		_ = op.clean()
		return err
	}
	// close merge operator
	if err = op.Close(); err != nil {
		return err
	}

	defer func() {
		// notify watcher
		if db.watcher != nil && db.watcher.expected(MergeEvent) {
			db.watcher.push(&Event{Type: MergeEvent, Value: lastActiveId})
		}
	}()

	if !domerge {
		return nil
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	// close wal data, ready to transfer merge data
	if err = db.closeWal(); err != nil {
		return err
	}

	// do real transfer
	if err = op.doTransfer(); err != nil {
		return err
	}

	// reload data
	if err = db.load(); err != nil {
		return err
	}

	return nil
}

func (db *DB) loadIndexFromHint() error {
	hint := db.hint
	if hint.IsEmpty() {
		return nil
	}

	it, err := hint.Iterator(0, hint.ActiveFid(), wal.ChunkPos{})
	if err != nil {
		return err
	}

	for {
		rawhint, _, err := it.NextData()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		// ignore error when loading hint
		db.index.Put(index.UnMarshalHint(rawhint))
		db.numOfRecord++
	}

	return nil
}

type mergeOP struct {
	db *DB

	// merged data
	merged *wal.Wal
	// hint data
	hint *wal.Wal
	// record fid of merged finished file
	finish *wal.Wal
}

func (op *mergeOP) doTransfer() error {
	// check dir if exist
	mergeDir := op.db.option.mergeDir
	if _, err := os.Stat(mergeDir); err != nil {
		return nil
	}
	defer os.RemoveAll(mergeDir)

	if op.finish == nil {
		if err := op.loadFinishedWal(); err != nil {
			return err
		}
	}

	// check if finished
	lastFid, err := hasFinished(op.finish)
	// if not finished, just return
	if errors.Is(err, ErrMergedNotFinished) {
		return nil
	} else if err != nil {
		return err
	}

	// close file, it will be moved to data dir after soon
	if err = op.finish.Close(); err != nil {
		return err
	}
	op.finish = nil

	// remove original data
	for fid := uint32(0); fid <= lastFid; fid++ {
		walfilename := wal.WalFileName(op.db.option.dataDir, fid, dataName)
		_, err := os.Stat(walfilename)
		if err != nil {
			continue
		}

		if err := os.Remove(walfilename); err != nil {
			return err
		}
	}

	dir, err := os.ReadDir(op.db.option.mergeDir)
	if err != nil {
		return err
	}

	// transfer merged data
	for _, f := range dir {
		if f.IsDir() {
			continue
		}

		fname := f.Name()
		if !strings.HasSuffix(fname, dataName) &&
			!strings.HasSuffix(fname, hintName) &&
			!strings.HasSuffix(fname, finishedName) {
			continue
		}

		err := os.Rename(path.Join(op.db.option.mergeDir, fname), path.Join(op.db.option.dataDir, fname))
		if err != nil {
			return err
		}
	}

	return nil
}

func (op *mergeOP) clean() error {
	if err := os.RemoveAll(op.db.option.mergeDir); err != nil {
		return err
	}
	return nil
}

func (op *mergeOP) load() error {
	if err := op.loadMergedWal(); err != nil {
		return err
	}
	if err := op.loadHintWal(); err != nil {
		return err
	}
	if err := op.loadFinishedWal(); err != nil {
		return err
	}
	return nil
}

func (op *mergeOP) reload() error {
	_ = op.Close()
	err := op.clean()
	if err != nil {
		return err
	}
	err = op.load()
	if err != nil {
		return err
	}
	return nil
}

func (op *mergeOP) loadMergedWal() error {
	merged, err := wal.Open(wal.Option{
		DataDir:     op.db.option.mergeDir,
		MaxFileSize: op.db.option.MaxSize,
		Ext:         dataName,
	})
	op.merged = merged
	return err
}

func (op *mergeOP) loadHintWal() error {
	hint, err := openHint(op.db.option.mergeDir)
	op.hint = hint
	return err
}

func (op *mergeOP) loadFinishedWal() error {
	finished, err := openFinished(op.db.option.mergeDir)
	op.finish = finished
	return err
}

func (op *mergeOP) doMerge(lastActiveId uint32) error {
	db := op.db

	it, err := db.data.Iterator(0, lastActiveId, wal.ChunkPos{})
	if err != nil {
		return err
	}

	txnSequences := make(map[int64][]entryhint, 1<<10)

	var i int
	// iterate over immutable files and merge redundant entries
	for {
		rawData, pos, err := it.NextData()
		if err != nil {
			// all files were read finished
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Printf("%d, %+v\n", i, pos)
			return err
		}
		i++

		// unmarshal the raw data
		record, err := db.serializer.UnMarshalEntry(rawData)
		if err != nil {
			return err
		}

		if record.TxId == mergedTxnId {
			if err := op.doWrite(entryhint{entry: record, pos: pos}); err != nil {
				return err
			}
			continue
		}

		// process record data
		switch record.Type {
		// normal data flag
		case entry.DataEntryType:
			txnSequences[record.TxId] = append(txnSequences[record.TxId], entryhint{entry: record, pos: pos})
		// transaction finished flag
		case entry.TxnCommitEntryType:
			seqs := txnSequences[record.TxId]
			for _, en := range seqs {
				en.entry.TxId = mergedTxnId
				err := op.doWrite(en)
				if err != nil {
					return err
				}
			}

			// release mem
			txnSequences[record.TxId] = nil
		}
	}

	// manual sync
	if err = op.merged.Sync(); err != nil {
		return err
	}

	// record of the last data fid before merge. By using this fid
	// it is possible to determine which intervals in the data file can be replaced once the database is restarted.
	err = op.finished(lastActiveId)
	if err != nil {
		return err
	}

	return nil
}

func (op *mergeOP) doWrite(eh entryhint) error {
	db := op.db
	hint, has := db.index.Get(eh.entry.Key)
	// expired and delete or maybe has been written to new active file
	if !has || hint.Fid != eh.pos.Fid || entry.IsExpired(hint.TTL) {
		return nil
	}

	marshalEntry, err := db.serializer.MarshalEntry(eh.entry)
	if err != nil {
		return err
	}

	newPos, err := op.merged.Write(marshalEntry)
	if err != nil {
		return err
	}

	newHint := index.Hint{Key: eh.entry.Key, TTL: eh.entry.TTL, ChunkPos: newPos}
	marshalHint := index.MarshalHint(newHint)
	if _, err := op.hint.Write(marshalHint); err != nil {
		return err
	}

	return nil
}

// record of the last merge wal fid
func (op *mergeOP) finished(lastFid uint32) error {
	bs := make([]byte, binary.MaxVarintLen32)
	binary.LittleEndian.PutUint32(bs, lastFid)
	if _, err := op.finish.Write(bs); err != nil {
		return err
	}
	return nil
}

func (op *mergeOP) Close() error {
	closes := []io.Closer{
		op.merged,
		op.hint,
		op.finish,
	}

	for _, closer := range closes {
		if closer == (*wal.Wal)(nil) {
			continue
		}
		if err := closer.Close(); err != nil {
			return err
		}
	}

	op.merged = nil
	op.hint = nil
	op.finish = nil
	return nil
}

func openHint(dir string) (*wal.Wal, error) {
	return wal.Open(wal.Option{
		DataDir:     dir,
		MaxFileSize: math.MaxInt64,
		Ext:         hintName,
	})
}

func openFinished(dir string) (*wal.Wal, error) {
	return wal.Open(wal.Option{
		DataDir:     dir,
		MaxFileSize: math.MaxInt64,
		Ext:         finishedName,
	})
}

func hasFinished(fwal *wal.Wal) (uint32, error) {
	if fwal.IsEmpty() {
		return 0, ErrMergedNotFinished
	}

	// read data from head of wal file, if merged finished, it must be valid
	bytes, err := fwal.Read(wal.ChunkPos{
		Fid:    fwal.ActiveFid(),
		Block:  0,
		Offset: 0,
		Size:   binary.MaxVarintLen32,
	})

	// if is eof which means this is a new wal instance
	if errors.Is(err, io.EOF) {
		return 0, ErrMergedNotFinished
	} else if err != nil {
		return 0, err
	}

	fid := binary.LittleEndian.Uint32(bytes)
	return fid, nil
}

// it will continue listening the event of write events, and do merge when reaching the checkpoint.
// this is a dead simple implementation, maybe can not meet your requirements, but db.Merge is public method,
// so you can disable default checkpoint and use db.Merge by yourself at the right time.
func doMergeAtCheckpoint(db *DB) {
	watcher, _ := db.Watcher("riverdb_merge_checkpoint_watcher", PutEvent, DelEvent)
	listen, _ := watcher.Listen()

	var lastMergeT time.Time
	for {
		select {
		case <-db.ctx.Done():
			// db closed
			return
		case _, ok := <-listen:
			// closed
			if !ok {
				return
			}

			if db.flag.Check(closed) {
				return
			}

			// check db stats
			stats := db.Stats()

			// no need to merge
			if stats.RecordNums < 10_000 && stats.DataSize < types.MB*10 {
				continue
			}

			// reach the checkpoint
			if float64(stats.RecordNums)/float64(stats.KeyNums) > db.option.MergeCheckpoint {
				// 5 min after last merge
				if time.Now().Sub(lastMergeT) < time.Minute*5 {
					continue
				}
				if err := db.Merge(true); err != nil {
					panic(err)
				}
				lastMergeT = time.Now()
			}

			time.Sleep(time.Millisecond * 50)
		}
	}
}
