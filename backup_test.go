package riverdb

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestDB_Backup(t *testing.T) {
	db, err := Open(DefaultOptions, WithDir(filepath.Join(os.TempDir(), "river")))
	assert.Nil(t, err)
	assert.NotNil(t, db)

	defer func() {
		err = db.Purge()
		assert.Nil(t, err)
		err = db.Close()
		assert.Nil(t, err)
	}()

	err = db.Put([]byte("hello world!"), []byte("hahahah"), 0)
	assert.Nil(t, err)

	err = db.Backup("now.zip")
	assert.Nil(t, err)
}

func TestDB_Recover(t *testing.T) {
	db, err := Open(DefaultOptions, WithDir(filepath.Join(os.TempDir(), "river")))
	assert.Nil(t, err)
	assert.NotNil(t, db)

	defer func() {
		err = db.Purge()
		assert.Nil(t, err)
		err = db.Close()
		assert.Nil(t, err)
	}()

	err = db.Recover("now.zip")
	assert.Nil(t, err)

	value, err := db.Get([]byte("hello world!"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("hahahah"), value)
}
