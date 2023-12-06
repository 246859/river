package riverdb

import (
	"github.com/246859/river/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBatch_Empty(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	batch, err := db.Batch(BatchOption{
		Size:         200,
		SyncPerBatch: false,
		SyncOnFlush:  true,
	})

	assert.Nil(t, err)

	err = batch.WriteAll(nil)
	assert.Nil(t, err)

	err = batch.Flush()
	assert.Nil(t, err)
}

func TestBatch_ErrorOption(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	batch, err := db.Batch(BatchOption{
		Size:         -1,
		SyncPerBatch: false,
		SyncOnFlush:  true,
	})

	assert.NotNil(t, err)
	assert.Nil(t, batch)
}

func TestBatch_WriteAll(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	batch, err := db.Batch(BatchOption{
		Size:        200,
		SyncOnFlush: true,
	})

	rslen := 10000

	testkv := testRandKV()
	var rs []Record
	for i := 0; i < rslen; i++ {
		rs = append(rs, Record{
			K:   testkv.testUniqueBytes(20),
			V:   testkv.testBytes(2000),
			TTL: 0,
		})
	}

	err = batch.WriteAll(rs)
	assert.Nil(t, err)

	err = batch.Flush()
	assert.Nil(t, err)

	for _, r := range rs {
		value, err := db.Get(r.K)
		assert.Nil(t, err)
		assert.EqualValues(t, r.V, value)
	}
}

func TestBatch_DeleteAll(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	testkv := testRandKV()
	var rs []Record
	var ks []Key
	for i := 0; i < 1100; i++ {
		rs = append(rs, Record{
			K:   testkv.testUniqueBytes(20),
			V:   testkv.testBytes(2000),
			TTL: 0,
		})
		ks = append(ks, rs[i].K)
	}

	for _, r := range rs {
		err := db.Put(r.K, r.V, r.TTL)
		assert.Nil(t, err)
	}

	for _, r := range rs {
		value, err := db.Get(r.K)
		assert.Nil(t, err)
		assert.EqualValues(t, r.V, value)
	}

	batch, err := db.Batch(BatchOption{
		Size:         200,
		SyncPerBatch: true,
	})

	err = batch.DeleteAll(ks)
	assert.Nil(t, err)

	err = batch.Flush()
	assert.Nil(t, err)

	for _, r := range rs {
		value, err := db.Get(r.K)
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Nil(t, value)
	}

}

func TestBatch_WriteAll_Fragment(t *testing.T) {
	opt := DefaultOptions
	opt.BlockCache = 0
	opt.FsyncThreshold = 100 * types.KB

	db, closeDB, err := testDB(t.Name(), opt)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		t.Log(err)
		assert.Nil(t, err)
	}()

	// make sure key is unique
	testkv := testRandKV()

	var samples []Record
	for i := 0; i < 10000; i++ {
		samples = append(samples, Record{
			K:   testkv.testUniqueBytes(100),
			V:   testkv.testBytes(10 * types.KB),
			TTL: 0,
		})
	}

	batch, err := db.Batch(BatchOption{
		Size:        500,
		SyncOnFlush: true,
	})

	err = batch.WriteAll(samples)
	assert.Nil(t, err)

	err = batch.Flush()
	assert.Nil(t, err)

	for _, sample := range samples {
		value, err := db.Get(sample.K)
		assert.Nil(t, err)
		assert.EqualValues(t, sample.V, value)
	}
}

func TestBatch_Cut(t *testing.T) {
	opt := DefaultOptions
	opt.BlockCache = 0
	opt.MaxSize = types.MB
	opt.FsyncThreshold = 100 * types.KB

	db, closeDB, err := testDB(t.Name(), opt)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		t.Log(err)
		assert.Nil(t, err)
	}()

	// make sure key is unique
	testkv := testRandKV()

	var samples []Record
	for i := 0; i < 1000; i++ {
		samples = append(samples, Record{
			K:   testkv.testUniqueBytes(100),
			V:   testkv.testBytes(10 * types.KB),
			TTL: 0,
		})
	}

	batch, err := db.Batch(BatchOption{
		Size:        500,
		SyncOnFlush: true,
	})

	err = batch.WriteAll(samples)
	assert.Nil(t, err)

	err = batch.Flush()
	assert.Nil(t, err)

	assert.EqualValues(t, 1000, batch.Effected())

	for _, sample := range samples {
		value, err := db.Get(sample.K)
		assert.Nil(t, err)
		assert.EqualValues(t, sample.V, value)
	}
}
