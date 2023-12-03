package riverdb

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestDB_Watcher_Empty(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	watcher, err := db.Watcher()
	assert.Nil(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		listen, err := watcher.Listen()
		assert.Nil(t, err)

		for event := range listen {
			t.Log(event.Type)
		}
		t.Log("closed listener")
		wg.Done()
	}()

	time.Sleep(time.Second)
	// close listener
	err = watcher.Close()
	assert.Nil(t, err)

	wg.Wait()
}

func TestDB_Watcher_Put_Del(t *testing.T) {
	options := DefaultOptions
	options.WatchEvents = []EventType{PutEvent, DelEvent}

	db, closeDB, err := testDB(t.Name(), options)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	rs := []Record{
		{K: []byte("foo"), V: []byte("bar")},
		{K: []byte("foo1"), V: []byte("bar")},
		{K: []byte("foo2"), V: []byte("bar")},
		{K: []byte("foo3"), V: []byte("bar")},
	}

	watcher, err := db.Watcher()
	assert.Nil(t, err)

	for _, r := range rs {
		err := db.Put(r.K, r.V, r.TTL)
		assert.Nil(t, err)
		err = db.Del(r.K)
		assert.Nil(t, err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		listen, err := watcher.Listen()
		assert.Nil(t, err)

		var i int
		for event := range listen {
			t.Log(event.Type)
			assert.Contains(t, options.WatchEvents, event.Type)
			i++
		}
		t.Log("closed listener")
		assert.Equal(t, len(rs)*2, i)
		wg.Done()
	}()

	time.Sleep(time.Second)
	// close listener
	err = watcher.Close()
	assert.Nil(t, err)

	wg.Wait()
}

func TestDB_Watcher_Put_Del_Multi(t *testing.T) {
	options := DefaultOptions
	options.WatchEvents = []EventType{PutEvent, DelEvent}

	db, closeDB, err := testDB(t.Name(), options)
	assert.Nil(t, err)

	rs := []Record{
		{K: []byte("foo"), V: []byte("bar")},
		{K: []byte("foo1"), V: []byte("bar")},
		{K: []byte("foo2"), V: []byte("bar")},
		{K: []byte("foo3"), V: []byte("bar")},
	}

	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			watcher, err := db.Watcher()
			assert.Nil(t, err)
			listen, err := watcher.Listen()
			assert.Nil(t, err)

			var i int
			for event := range listen {
				t.Log(event.Type)
				assert.Contains(t, options.WatchEvents, event.Type)
				i++
			}
			t.Log("closed listener")
			assert.Equal(t, len(rs)*2, i)
		}()
	}

	for _, r := range rs {
		err := db.Put(r.K, r.V, r.TTL)
		assert.Nil(t, err)
		err = db.Del(r.K)
		assert.Nil(t, err)
	}

	time.Sleep(time.Second * 3)
	assert.Nil(t, closeDB())
	wg.Wait()
}

func TestDB_Watcher_BackUp(t *testing.T) {
	options := DefaultOptions
	options.WatchEvents = []EventType{BackupEvent, RecoverEvent}

	db, closeDB, err := testDB(t.Name(), options)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	rs := []Record{
		{K: []byte("foo"), V: []byte("bar")},
		{K: []byte("foo1"), V: []byte("bar")},
		{K: []byte("foo2"), V: []byte("bar")},
		{K: []byte("foo3"), V: []byte("bar")},
	}

	for _, r := range rs {
		err := db.Put(r.K, r.V, r.TTL)
		assert.Nil(t, err)
	}

	target := filepath.Join(os.TempDir(), "test.tar.gz")

	var wg sync.WaitGroup
	wg.Add(2)

	backWatcher, err := db.Watcher(BackupEvent)
	assert.Nil(t, err)

	go func() {
		defer wg.Done()
		listen, err := backWatcher.Listen()
		assert.Nil(t, err)

		var i int
		for event := range listen {
			t.Log(event.Type)
			assert.Contains(t, []EventType{BackupEvent}, event.Type)
			i++
		}
		t.Log("closed listener")
		assert.Equal(t, 1, i)
	}()

	recoverWatcher, err := db.Watcher(RecoverEvent)
	assert.Nil(t, err)

	go func() {
		defer wg.Done()
		listen, err := recoverWatcher.Listen()
		assert.Nil(t, err)

		var i int
		for event := range listen {
			t.Log(event.Type)
			assert.Contains(t, []EventType{RecoverEvent}, event.Type)
			i++
		}
		t.Log("closed listener")
		assert.Equal(t, 1, i)
	}()

	err = db.Backup(target)
	assert.Nil(t, err)

	err = db.Recover(target)
	assert.Nil(t, err)

	time.Sleep(time.Second)

	err = backWatcher.Close()
	assert.Nil(t, err)

	err = recoverWatcher.Close()
	assert.Nil(t, err)

	wg.Wait()
}
