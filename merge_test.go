package riverdb

import (
	"fmt"
	"github.com/246859/river/types"
	"github.com/stretchr/testify/assert"
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
			k:   []byte(fmt.Sprintf("%d", i)),
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
		if err != nil {
			t.Log("key==>>>>", sample.k)
		}
		assert.Nil(t, err)
		assert.EqualValues(t, value, sample.v)
	}
}

func TestDB_Merge_1(t *testing.T) {
	for i := 0; i < 50; i++ {
		t.Log(i)
		testDB_Merge(t)
		time.Sleep(100 * time.Millisecond)
	}
}

func TestDB_Merge_Txn_Truncate(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	testkv := testRandKV()

	var rs []Record
	var rsmu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(11)

	for i := 0; i < 10; i++ {
		go func(gid int) {
			defer wg.Done()
			time.Sleep(time.Duration(gid) * 500 * time.Microsecond)
			err := db.Begin(func(txn *Txn) error {
				t.Logf("go-%d writing", gid)
				assert.Nil(t, err)
				for i := 0; i < 10; i++ {
					r := Record{
						K:   []byte(fmt.Sprintf("%d-%d-%s", gid, i, testkv.testUniqueBytes(20))),
						V:   testkv.testBytes(1024),
						TTL: 0,
					}
					err := txn.Put(r.K, r.V, r.TTL)
					assert.Nil(t, err)
					t.Logf("go-%d-%d --> fid %d", gid, i, db.data.ActiveFid())

					rsmu.Lock()
					rs = append(rs, r)
					rsmu.Unlock()
				}
				t.Logf("go-%d finished", gid)
				assert.Nil(t, err)

				return nil
			})
			assert.Nil(t, err)
		}(i)
	}

	go func() {
		defer wg.Done()
		time.Sleep(300 * time.Microsecond)
		t.Log("<<< merge begin >>>")
		err := db.Merge(true)
		assert.Nil(t, err)
		t.Log("<<< merge finished >>>")
	}()

	wg.Wait()

	for _, r := range rs {
		value, err := db.Get(r.K)
		assert.Nil(t, err)
		if err != nil {
			t.Log(r.K)
		}
		assert.EqualValues(t, r.V, value)
	}
}

func TestMerge_CheckPoint(t *testing.T) {
	// watch merge event
	opt := DefaultOptions
	opt.WatchSize = 30000
	opt.MergeCheckpoint = 3.5
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

	watcher, err := db.Watcher("put_get", MergeEvent)
	assert.Nil(t, err)

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

	for _, r := range rs {
		err := db.Put(r.K, r.V, r.TTL)
		assert.Nil(t, err)
	}

	stats := db.Stats()

	t.Log(stats.DataSize)
	t.Log(stats.KeyNums)
	t.Log(stats.RecordNums)
	time.Sleep(time.Second * 2)
	err = watcher.Close()
	assert.Nil(t, err)
	wg.Wait()
}
