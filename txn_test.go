package riverdb

import (
	"github.com/246859/river/types"
	"github.com/stretchr/testify/assert"
	"io"
	"sync"
	"testing"
	"time"
)

func TestTxn_Begin_Commit(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	err = db.Begin(func(txn *Txn) error {
		return nil
	})
	assert.Nil(t, err)
}

func TestTxn_Put_Get(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	testkv := testRandKV()
	// put-commit
	{
		k, v := testkv.testUniqueBytes(10), testkv.testBytes(types.KB)
		err := db.Begin(func(txn *Txn) error {

			err = txn.Put(k, v, 0)
			assert.Nil(t, err)

			value, err := txn.Get(k)
			assert.Nil(t, err)
			assert.Equal(t, v, value)

			get, err := db.Get(k)
			assert.ErrorIs(t, err, ErrKeyNotFound)
			assert.Nil(t, get)

			return nil
		})

		get, err := db.Get(k)
		assert.Nil(t, err)
		assert.Equal(t, v, get)
	}

	// put-rollback
	{
		k, v := testkv.testUniqueBytes(10), testkv.testBytes(types.KB)
		err := db.Begin(func(txn *Txn) error {
			err = txn.Put(k, v, 0)
			assert.Nil(t, err)

			value, err := txn.Get(k)
			assert.Nil(t, err)
			assert.Equal(t, v, value)

			get, err := db.Get(k)
			assert.ErrorIs(t, err, ErrKeyNotFound)
			assert.Nil(t, get)

			return io.EOF
		})
		assert.NotNil(t, err)

		get, err := db.Get(k)
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Nil(t, get)
	}
}
func TestTxn_Readonly_1(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	err = db.View(func(txn *Txn) error {
		err := txn.Put([]byte("1"), []byte("1"), 0)
		assert.ErrorIs(t, err, ErrTxnReadonly)
		return err
	})
	assert.ErrorIs(t, err, ErrTxnReadonly)
}

func TestTxn_Readonly_2(t *testing.T) {
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

	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			err := db.View(func(txn *Txn) error {
				for _, s := range samples {
					get, err := txn.Get(s.k)
					assert.Nil(t, err)
					assert.Equal(t, s.v, get)
				}
				return nil
			})
			assert.Nil(t, err)
		}()
	}

	wg.Wait()
}

func TestTxn_UpdateOnly(t *testing.T) {
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

	var wg sync.WaitGroup
	var mu sync.Mutex
	var records []record
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			err := db.Begin(func(txn *Txn) error {
				var rs []record
				for i := 0; i < 10; i++ {
					rs = append(rs, record{
						k:   testkv.testUniqueBytes(20),
						v:   testkv.testBytes(types.KB * 10),
						ttl: 0,
					})
				}
				for _, r := range rs {
					err := txn.Put(r.k, r.v, r.ttl)
					assert.Nil(t, err)
				}

				mu.Lock()
				records = append(records, rs...)
				mu.Unlock()
				return nil
			})
			assert.Nil(t, err)
		}()
	}

	wg.Wait()

	for _, r := range records {
		get, err := db.Get(r.k)
		assert.Nil(t, err)
		assert.Equal(t, r.v, get)
	}
}

func TestTxn_Mixed(t *testing.T) {
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

	var wg sync.WaitGroup
	wg.Add(2)

	s := samples[100]
	// txn1
	go func() {
		defer wg.Done()
		err := db.Begin(func(txn *Txn) error {
			value, err := txn.Get(s.k)
			assert.Nil(t, err)

			err = txn.Put(s.k, append(value, 1), 0)
			assert.Nil(t, err)

			time.Sleep(80 * time.Millisecond)
			return nil
		})
		assert.Nil(t, err)
	}()

	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		err := db.Begin(func(txn *Txn) error {
			time.Sleep(200 * time.Millisecond)
			assert.Nil(t, err)
			value, err := txn.Get(s.k)
			assert.Nil(t, err)

			err = txn.Put(s.k, append(value, 1), 0)
			assert.Nil(t, err)
			return nil
		})
		assert.ErrorIs(t, err, ErrTxnConflict)
	}()

	wg.Wait()
}

func TestTxn_Serializable(t *testing.T) {
	opt := DefaultOptions
	opt.Level = Serializable
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

	var wg sync.WaitGroup
	wg.Add(2)

	s := samples[100]
	// txn1
	go func() {
		defer wg.Done()
		err := db.Begin(func(txn *Txn) error {
			value, err := txn.Get(s.k)
			assert.Nil(t, err)

			err = txn.Put(s.k, append(value, 1), 0)
			assert.Nil(t, err)

			time.Sleep(80 * time.Millisecond)
			return nil
		})
		assert.Nil(t, err)
	}()

	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		err := db.Begin(func(txn *Txn) error {
			time.Sleep(200 * time.Millisecond)
			assert.Nil(t, err)
			value, err := txn.Get(s.k)
			assert.Nil(t, err)

			err = txn.Put(s.k, append(value, 1), 0)
			assert.Nil(t, err)
			return nil
		})
		assert.Nil(t, err)
	}()

	wg.Wait()
}
