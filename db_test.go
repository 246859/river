package riverdb

import (
	"github.com/246859/river/entry"
	"github.com/246859/river/file"
	"github.com/stretchr/testify/assert"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDB_Open(t *testing.T) {
	// default open
	{
		_, closeDB, err := testDB(t, DefaultOptions)
		assert.Nil(t, err)
		assert.Nil(t, closeDB())
	}

	// error empty dir
	{
		opt := DefaultOptions
		opt.Dir = ""
		db, err := Open(opt)
		assert.NotNil(t, err)
		assert.Nil(t, db)
	}

	// error nil comparator
	{
		opt := DefaultOptions
		opt.Compare = nil
		db, err := Open(opt)
		assert.NotNil(t, err)
		assert.Nil(t, db)
	}

	// error zero max size
	{
		opt := DefaultOptions
		opt.MaxSize = 0
		db, err := Open(opt)
		assert.NotNil(t, err)
		assert.Nil(t, db)
	}

	// error invalid max size
	{
		opt := DefaultOptions
		opt.MaxSize = int64(opt.BlockCache)
		db, err := Open(opt)
		assert.NotNil(t, err)
		assert.Nil(t, db)
	}

	// error invalid block cache
	{
		opt := DefaultOptions
		opt.BlockCache = uint32(opt.MaxSize)
		db, err := Open(opt)
		assert.NotNil(t, err)
		assert.Nil(t, db)
	}

	// error invalid threshold
	{
		opt := DefaultOptions
		opt.FsyncThreshold = opt.MaxSize + 1
		db, err := Open(opt)
		assert.NotNil(t, err)
		assert.Nil(t, db)
	}
}

func TestDB_Put_Get_1(t *testing.T) {
	db, closeDB, err := testDB(t, DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	samples := []struct {
		k   []byte
		v   []byte
		ttl time.Duration
	}{
		{[]byte("key"), []byte("value"), 0},
		{[]byte("key1"), []byte("value1"), -1},
		{[]byte("key2"), []byte(strings.Repeat("a", file.KB)), -1},
		{[]byte(strings.Repeat("a", 10000)), []byte(strings.Repeat("a", file.KB)), -1},
	}

	for _, sample := range samples {
		err := db.Put(sample.k, sample.v, sample.ttl)
		assert.Nil(t, err)
	}

	for _, sample := range samples {
		value, err := db.Get(sample.k)
		assert.Nil(t, err)
		assert.Equal(t, sample.v, value)
	}
}

func TestDB_Put_Get_2(t *testing.T) {
	db, closeDB, err := testDB(t, DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	// set the expiration, will be expired soon
	{
		key, value := []byte("hello world!"), []byte("value")
		err = db.Put(key, value, 1)
		assert.Nil(t, err)

		getV, err := db.Get(key)
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Equal(t, []byte(nil), getV)
	}

	// set the expiration, will not be expired soon
	{
		key, value := []byte("hello world!"), []byte("value")
		err = db.Put(key, value, time.Hour)
		assert.Nil(t, err)

		getV, err := db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value, getV)
	}

	// update existing key
	{
		key, value := []byte("hello world!"), []byte("value")
		err = db.Put(key, value, 0)
		assert.Nil(t, err)

		getV, err := db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value, getV)

		value2 := []byte("value2")
		err = db.Put(key, value2, 0)
		assert.Nil(t, err)

		getV2, err := db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value2, getV2)
	}

	// update the repeat value, and set the expiration
	{
		key, value := []byte("key3"), []byte("value3")
		err = db.Put(key, value, 0)
		assert.Nil(t, err)

		getV, err := db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value, getV)

		err = db.Put(key, value, 1)
		assert.Nil(t, err)

		getV2, err := db.Get(key)
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Equal(t, []byte(nil), getV2)
	}

	// set the nil key
	{
		key, value := []byte(nil), []byte("value3")
		err = db.Put(key, value, 0)
		assert.ErrorIs(t, err, entry.ErrNilKey)

		getV, err := db.Get(key)
		assert.ErrorIs(t, err, entry.ErrNilKey)
		assert.Equal(t, []byte(nil), getV)
	}

	// nil value
	{
		key, value := []byte("hello world!"), []byte(nil)
		err = db.Put(key, value, 0)
		assert.Nil(t, err)

		getV, err := db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, []byte{}, getV)
	}
}

func TestDB_Put_Get_Fragment(t *testing.T) {
	opt := DefaultOptions
	opt.MaxSize = file.MB
	opt.BlockCache = 0
	opt.FsyncThreshold = 100 * file.KB

	db, closeDB, err := testDB(t, opt)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		t.Log(err)
		assert.Nil(t, err)
	}()

	type record struct {
		k   []byte
		v   []byte
		ttl time.Duration
	}

	// make sure key is unique
	testkv := testRandKV()

	var samples []record
	for i := 0; i < 10000; i++ {
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

	for _, sample := range samples {
		value, err := db.Get(sample.k)
		assert.Nil(t, err)
		assert.EqualValues(t, sample.v, value)
	}
}

func TestDB_Put_Concurrent(t *testing.T) {
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

	// make sure key is unique
	testkv := testRandKV()

	var mu sync.Mutex
	var wg sync.WaitGroup
	var records []record
	wg.Add(200)

	for i := 0; i < 200; i++ {
		go func() {
			defer wg.Done()
			r := record{
				k:   testkv.testUniqueBytes(20),
				v:   testkv.testBytes(file.KB * 10),
				ttl: 0,
			}
			err := db.Put(r.k, r.v, r.ttl)
			assert.Nil(t, err)

			mu.Lock()
			records = append(records, r)
			mu.Unlock()
		}()
	}

	wg.Wait()

	assert.Equal(t, 200, db.index.Size())

	for _, r := range records {
		value, err := db.Get(r.k)
		assert.Nil(t, err)
		assert.EqualValues(t, value, r.v)
	}
}

func TestDB_Del(t *testing.T) {
	db, closeDB, err := testDB(t, DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	// normal delete
	{
		key, value := []byte("hello world!"), []byte("value")
		err = db.Put(key, value, 0)
		assert.Nil(t, err)

		getV, err := db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value, getV)

		err = db.Del(key)
		assert.Nil(t, err)

		getV, err = db.Get(key)
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Equal(t, []byte(nil), getV)
	}

	// delete twice
	{
		key, value := []byte("hello world!1"), []byte("value1")
		err = db.Put(key, value, 0)
		assert.Nil(t, err)

		getV, err := db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value, getV)

		err = db.Del(key)
		assert.Nil(t, err)

		getV, err = db.Get(key)
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Equal(t, []byte(nil), getV)

		err = db.Del(key)
		assert.Nil(t, err)

		getV, err = db.Get(key)
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Equal(t, []byte(nil), getV)
	}
}

func TestDB_TTL(t *testing.T) {
	db, closeDB, err := testDB(t, DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	// persistent key-value
	{
		key, value := []byte("hello world!"), []byte("value")
		err = db.Put(key, value, 0)
		assert.Nil(t, err)

		getV, err := db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value, getV)

		ttl, err := db.TTL(key)
		assert.Nil(t, err)
		assert.Equal(t, time.Duration(0), ttl)
	}

	// expired in an hour key-value
	{
		key, value := []byte("hello world!1"), []byte("value")
		err = db.Put(key, value, time.Hour)
		assert.Nil(t, err)

		getV, err := db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value, getV)

		time.Sleep(time.Second)

		ttl, err := db.TTL(key)
		assert.Nil(t, err)
		assert.Less(t, ttl, time.Hour)
	}

	// expired right now
	{
		key, value := []byte("hello world!2"), []byte("value")
		err = db.Put(key, value, 1)
		assert.Nil(t, err)

		time.Sleep(100)

		ttl, err := db.TTL(key)
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Equal(t, time.Duration(0), ttl)
	}
}

func TestDB_Expire(t *testing.T) {
	db, closeDB, err := testDB(t, DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	// expire soon
	{
		key, value := []byte("hello world!"), []byte("value")
		err = db.Put(key, value, 0)
		assert.Nil(t, err)

		getV, err := db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value, getV)

		err = db.Expire(key, 1)
		assert.Nil(t, err)

		getV, err = db.Get(key)
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Equal(t, []byte(nil), getV)
	}

	// expire persistent
	{
		key, value := []byte("hello world!1"), []byte("value")
		err = db.Put(key, value, time.Second)
		assert.Nil(t, err)

		getV, err := db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value, getV)

		err = db.Expire(key, 0)
		assert.Nil(t, err)

		time.Sleep(time.Second)

		getV, err = db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value, getV)
	}

	// expired later
	{
		key, value := []byte("hello world!2"), []byte("value2")
		err = db.Put(key, value, 0)
		assert.Nil(t, err)

		getV, err := db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value, getV)

		err = db.Expire(key, time.Second)
		assert.Nil(t, err)

		getV, err = db.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, value, getV)

		time.Sleep(time.Second)

		getV, err = db.Get(key)
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Equal(t, []byte(nil), getV)
	}

	// expire expired
	{
		key, value := []byte("hello world!3"), []byte("value2")
		err = db.Put(key, value, 1)
		assert.Nil(t, err)

		err := db.Expire(key, 0)
		assert.ErrorIs(t, err, ErrKeyNotFound)
	}

	// expire -1
	{
		key, value := []byte("hello world!3"), []byte("value2")
		err = db.Put(key, value, time.Second)
		assert.Nil(t, err)

		err := db.Expire(key, -1)
		assert.Nil(t, err)

		time.Sleep(time.Second)

		getV, err := db.Get(key)
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Equal(t, []byte(nil), getV)
	}
}

func TestDB_Range(t *testing.T) {
	db, closeDB, err := testDB(t, DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	samples := []struct {
		k   []byte
		v   []byte
		ttl time.Duration
	}{
		{[]byte("key"), []byte("value"), 0},
		{[]byte("key1"), []byte("value1"), -1},
		{[]byte("key2"), []byte(strings.Repeat("a", file.KB)), -1},
		{[]byte("key3"), []byte(strings.Repeat("a", file.KB)), -1},
		{[]byte("key4"), []byte(strings.Repeat("a", file.KB)), -1},
		{[]byte("key5"), []byte(strings.Repeat("a", file.KB)), -1},
		{[]byte("key6"), []byte(strings.Repeat("a", file.KB)), -1},
		{[]byte("key7"), []byte(strings.Repeat("a", file.KB)), -1},
		{[]byte(strings.Repeat("a", 10000)), []byte(strings.Repeat("a", file.KB)), -1},
	}

	for _, sample := range samples {
		err := db.Put(sample.k, sample.v, sample.ttl)
		assert.Nil(t, err)
	}

	{
		err = db.Range(RangeOptions{
			Min:     nil,
			Max:     nil,
			Pattern: nil,
			Descend: false,
		}, func(key Key) bool {
			value, err := db.Get(key)
			assert.Nil(t, err)

			has := slices.ContainsFunc(samples, func(s struct {
				k   []byte
				v   []byte
				ttl time.Duration
			}) bool {
				return string(s.k) == string(key) && string(s.v) == string(value)
			})
			assert.True(t, has)

			return true
		})

		assert.Nil(t, err)
	}

	// min-max
	{
		count := 0
		err = db.Range(RangeOptions{
			Min:     []byte("key3"),
			Max:     []byte("key7"),
			Pattern: nil,
			Descend: false,
		}, func(key Key) bool {
			value, err := db.Get(key)
			assert.Nil(t, err)

			has := slices.ContainsFunc(samples, func(s struct {
				k   []byte
				v   []byte
				ttl time.Duration
			}) bool {
				return string(s.k) == string(key) && string(s.v) == string(value)
			})
			assert.True(t, has)

			count++
			return true
		})

		assert.Nil(t, err)
		assert.Equal(t, 5, count)
	}

	// min > max
	{
		err = db.Range(RangeOptions{
			Min:     []byte("key3"),
			Max:     []byte("key2"),
			Pattern: nil,
			Descend: false,
		}, func(key Key) bool {
			value, err := db.Get(key)
			assert.Nil(t, err)

			has := slices.ContainsFunc(samples, func(s struct {
				k   []byte
				v   []byte
				ttl time.Duration
			}) bool {
				return string(s.k) == string(key) && string(s.v) == string(value)
			})
			assert.True(t, has)

			return true
		})

		assert.NotNil(t, err)
	}

	// pattern matching
	{
		count := 0
		err = db.Range(RangeOptions{
			Pattern: []byte("[a]"),
			Descend: false,
		}, func(key Key) bool {
			value, err := db.Get(key)
			assert.Nil(t, err)

			has := slices.ContainsFunc(samples, func(s struct {
				k   []byte
				v   []byte
				ttl time.Duration
			}) bool {
				return string(s.k) == string(key) && string(s.v) == string(value)
			})
			assert.True(t, has)

			count++
			return true
		})

		assert.Nil(t, err)
		assert.Equal(t, 1, count)
	}
}
