package riverdb

import (
	"github.com/246859/river/file"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestDB_Merge_0(t *testing.T) {
	db, closeDB, err := testDB(t, DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	err = db.Merge(false)
	assert.Nil(t, err)
}

func testDB_Merge(t *testing.T) {
	opt := DefaultOptions
	opt.MaxSize = file.MB
	opt.BlockCache = 0
	opt.FsyncThreshold = 100 * file.KB

	db, closeDB, err := testDB(t, opt)
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

	var redundantSamples []record
	for i := 0; i < 300; i++ {
		redundantSamples = append(redundantSamples, record{
			k:   testkv.testUniqueBytes(100),
			v:   testkv.testBytes(10 * file.KB),
			ttl: 0,
		})
	}

	for _, sample := range redundantSamples {
		err := db.Put(sample.k, sample.v, sample.ttl)
		assert.Nil(t, err)
	}

	// delete redundant data
	for _, sample := range redundantSamples {
		err := db.Del(sample.k)
		assert.Nil(t, err)
	}

	var validSamples []record
	for i := 0; i < 100; i++ {
		validSamples = append(validSamples, record{
			k:   testkv.testUniqueBytes(100),
			v:   testkv.testBytes(10 * file.KB),
			ttl: 0,
		})
	}

	for _, sample := range validSamples {
		err := db.Put(sample.k, sample.v, sample.ttl)
		assert.Nil(t, err)
	}

	err = db.Merge(true)
	assert.Nil(t, err)

	for _, sample := range validSamples {
		value, err := db.Get(sample.k)
		assert.Nil(t, err)
		assert.EqualValues(t, value, sample.v)
	}
}

func TestDB_Merge_1(t *testing.T) {
	for i := 0; i < 30; i++ {
		t.Log(i)
		testDB_Merge(t)
		time.Sleep(100 * time.Millisecond)
	}
}

func TestDB_Merge_2(t *testing.T) {
	db, closeDB, err := testDB(t, DefaultOptions)
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

	// redundant records
	for _, sample := range samples[:100] {
		err := db.Del(sample.k)
		assert.Nil(t, err)
	}

	var records []record
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(22)

	for i := 0; i < 20; i++ {
		i := i
		go func() {
			t.Log("writing", i)
			defer wg.Done()
			var samples1 []record
			for i := 0; i < 100; i++ {
				samples1 = append(samples1, record{
					k:   testkv.testUniqueBytes(100),
					v:   testkv.testBytes(10 * file.KB),
					ttl: 0,
				})
			}

			for _, sample := range samples1 {
				err := db.Put(sample.k, sample.v, sample.ttl)
				assert.Nil(t, err)
			}

			mu.Lock()
			records = append(records, samples1...)
			mu.Unlock()
		}()
	}

	go func() {
		defer wg.Done()
		t.Log("backing")
		err := db.Backup(filepath.Join(os.TempDir(), "data.zip"))
		assert.Nil(t, err)
	}()

	go func() {
		defer wg.Done()
		t.Log("merging")
		err = db.Merge(true)
		assert.Nil(t, err)
	}()

	wg.Wait()

	assert.Equal(t, 20*100+200, db.index.Size())

	records = append(records, samples[100:]...)
	for _, sample := range records {
		value, err := db.Get(sample.k)
		assert.Nil(t, err)
		assert.EqualValues(t, value, sample.v)
	}
}
