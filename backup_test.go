package riverdb

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestDB_Backup(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	err = db.Put([]byte("hello world!"), []byte("hahahah"), 0)
	assert.Nil(t, err)

	tarpath := filepath.Join(os.TempDir(), "test.tar.gz")
	err = db.Backup(filepath.Join(os.TempDir(), "test.tar.gz"))
	assert.Nil(t, err)

	_, err = os.Stat(tarpath)
	assert.Nil(t, err)
}

func TestDB_Recover(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	err = db.Put([]byte("hello world!"), []byte("hahahah"), 0)
	assert.Nil(t, err)

	err = db.Backup(filepath.Join(os.TempDir(), "test.tar.gz"))
	assert.Nil(t, err)

	err = db.Purge()
	assert.Nil(t, err)

	err = db.Recover(filepath.Join(os.TempDir(), "test.tar.gz"))
	assert.Nil(t, err)

	value, err := db.Get([]byte("hello world!"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("hahahah"), value)
}
