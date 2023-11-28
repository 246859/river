package riverdb

import (
	"github.com/246859/river/file"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDB_Watch_Mixed(t *testing.T) {
	db, closeDB, err := testDB(DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	type record struct {
		k   []byte
		v   []byte
		ttl time.Duration
	}

	testkv := testRandKV()

	var samples []record
	for i := 0; i < 300; i++ {
		samples = append(samples, record{
			k:   testkv.testUniqueBytes(100),
			v:   testkv.testBytes(10 * file.KB),
			ttl: 0,
		})
	}

	for _, sample := range samples {
		err := db.Put(sample.k, sample.v, sample.ttl)
		assert.Nil(t, err)
	}

	for _, r := range samples[:100] {
		err := db.Del(r.k)
		assert.Nil(t, err)
	}

	err = db.Merge(true)
	assert.Nil(t, err)

	watch, err := db.Watch()
	assert.Nil(t, err)

	go func() {
		i := 0
		for event := range watch {
			t.Log(event.Type)
			assert.Contains(t, []EventType{PutEvent, DelEvent, MergeEvent}, event.Type)
			i++
		}
		t.Log(i)

		t.Log("watcher closed")
	}()
	time.Sleep(time.Second * 4)
}

func TestDB_Watch_Backup(t *testing.T) {
	opt := DefaultOptions
	opt.WatchEvents = append(opt.WatchEvents, BackupEvent, RecoverEvent)
	db, closeDB, err := testDB(opt)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	dest := filepath.Join(os.TempDir(), "temp.zip")
	err = db.Backup(dest)
	assert.Nil(t, err)
	err = db.Recover(dest)
	assert.Nil(t, err)

	go func() {
		watch, err := db.Watch()
		assert.Nil(t, err)
		i := 0
		for event := range watch {
			assert.Contains(t, []EventType{BackupEvent, RecoverEvent}, event.Type)
			i++
		}
		t.Log(i)
		t.Log("watcher closed")
	}()

	time.Sleep(time.Second * 2)
}
