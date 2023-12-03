package riverdb

import (
	"github.com/246859/river/types"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestDB_Merge_0(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
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
	opt.MaxSize = types.MB
	opt.BlockCache = 0
	opt.FsyncThreshold = 100 * types.KB

	db, closeDB, err := testDB(t.Name(), opt)
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
			v:   testkv.testBytes(10 * types.KB),
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
			v:   testkv.testBytes(10 * types.KB),
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
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
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
			v:   testkv.testBytes(10 * types.KB),
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
					v:   testkv.testBytes(10 * types.KB),
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
		err := db.Backup(filepath.Join(os.TempDir(), "test.tar.gz"))
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

func TestMerge_CheckPoint(t *testing.T) {
	// watch merge event
	opt := DefaultOptions
	opt.WatchSize = 30000
	opt.WatchEvents = append(opt.WatchEvents, MergeEvent)
	db, closeDB, err := testDB(t.Name(), opt)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	testkv := testRandKV()

	var rs []Record

	// must trigger merge event
	for i := 0; i < 20_000; i++ {
		rs = append(rs, Record{
			K:   testkv.testBytes(3),
			V:   testkv.testBytes(types.KB * 5),
			TTL: 0,
		})
	}

	watcher, err := db.Watcher(MergeEvent)
	assert.Nil(t, err)

	for _, r := range rs {
		err := db.Put(r.K, r.V, r.TTL)
		assert.Nil(t, err)
	}

	stats := db.Stats()

	t.Log(stats.DataSize)
	t.Log(stats.KeyNums)
	t.Log(stats.RecordNums)

	// listen the merge events
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		listen, err := watcher.Listen()
		assert.Nil(t, err)

		var c int
		for event := range listen {
			t.Log(event)
			assert.Equal(t, MergeEvent, event.Type)
			c++
		}
		assert.Equal(t, 1, c)
	}()

	time.Sleep(time.Second)
	err = watcher.Close()
	assert.Nil(t, err)
	wg.Wait()
}
