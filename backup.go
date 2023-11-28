package riverdb

import (
	"fmt"
	"github.com/dstgo/filebox"
	"os"
	"strings"
)

// Backup use tar gzip to compress data wal files to dest path
func (db *DB) Backup(destpath string) error {
	if db.mask.CheckAny(closed) {
		return ErrDBClosed
	}

	if strings.HasPrefix(strings.ToLower(destpath), strings.ToLower(db.option.Dir)) {
		return fmt.Errorf("archive destination should not in db directory")
	}

	db.opmu.Lock()
	defer db.opmu.Unlock()

	db.mask.Store(backing)
	defer db.mask.Remove(backing)

	// check dir status
	datadir := db.option.dataDir
	_, err := os.Stat(datadir)
	if err != nil {
		return err
	}

	// rotate data wal file
	db.mu.Lock()
	if err := db.data.Rotate(); err != nil {
		db.mu.Unlock()
		return err
	}
	db.mu.Unlock()

	// tar gzip archive
	err = filebox.Zip(datadir, destpath)
	if err != nil {
		return err
	}

	// notify watcher
	if db.watcher != nil {
		db.watcher.push(&Event{Type: BackupEvent, Value: destpath})
	}
	return nil
}

// Recover recovers wal files from specified targz archive.
// it will purge current data, and overwrite by the backup.
func (db *DB) Recover(srcpath string) error {
	if db.mask.CheckAny(closed) {
		return ErrDBClosed
	}

	db.opmu.Lock()
	defer db.opmu.Unlock()

	db.mask.Store(recovering)
	defer db.mask.Remove(recovering)

	// check src path
	if _, err := os.Stat(srcpath); err != nil {
		return err
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.closeWal(); err != nil {
		return err
	}

	// remove all original data
	if err := os.RemoveAll(db.option.dataDir); err != nil {
		return err
	}

	// recover backup data
	if err := filebox.Unzip(srcpath, db.option.dataDir); err != nil {
		return err
	}

	// reload data
	if err := db.load(); err != nil {
		return err
	}

	// notify watcher
	if db.watcher != nil {
		db.watcher.push(&Event{Type: RecoverEvent, Value: srcpath})
	}
	return nil
}
