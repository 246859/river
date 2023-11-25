package riverdb

import (
	"os"
	"path/filepath"
)

// testDB return a test db for testing
func testDB(option Options) (db *DB, closeDB func() error, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}
	dir := filepath.Join(homeDir, "test_db")
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		return nil, nil, err
	}
	db, err = Open(option, WithDir(dir))
	if err != nil {
		return nil, nil, err
	}
	return db, func() error {
		if err := db.Purge(); err != nil {
			return err
		}
		if err := db.Close(); err != nil {
			return err
		}
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
		return nil
	}, nil
}
