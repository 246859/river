package riverdb

import (
	"github.com/dstgo/filebox"
	"os"
)

// Backup use tar gzip to compress data wal files to dest path
func (db *DB) Backup(destpath string) error {
	db.opmu.Lock()
	defer db.opmu.Unlock()

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
	return nil
}

// Recover recovers wal files from specified targz archive.
// it will purge current data, and overwrite by the backup.
func (db *DB) Recover(srcpath string) error {
	db.opmu.Lock()
	defer db.opmu.Unlock()

	// check src path
	if _, err := os.Stat(srcpath); err != nil {
		return err
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.closeWal(); err != nil {
		return err
	}

	if err := os.RemoveAll(db.option.dataDir); err != nil {
		return err
	}

	// untar to db root dir
	if err := filebox.Unzip(srcpath, db.option.dataDir); err != nil {
		return err
	}

	// reload data
	if err := db.loadData(); err != nil {
		return err
	}

	// reload index
	if err := db.loadIndex(); err != nil {
		return err
	}

	return nil
}
