package riverdb

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestDB_Stats_Empty(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	stats := db.Stats()
	assert.EqualValues(t, 0, stats.KeyNums)
	assert.EqualValues(t, 0, stats.RecordNums)
	assert.EqualValues(t, 0, stats.DataSize)
	assert.EqualValues(t, 0, stats.HintSize)
}

func TestDB_Stats_Mixed(t *testing.T) {
	db, closeDB, err := testDB(t.Name(), DefaultOptions)
	assert.Nil(t, err)
	defer func() {
		err := closeDB()
		assert.Nil(t, err)
	}()

	testkv := testRandKV()

	for i := 0; i < 10000; i++ {
		x := rand.Int() % 2
		if x == 0 {
			err := db.Put(testkv.testBytes(100), testkv.testBytes(100), 0)
			assert.Nil(t, err)
		} else {
			err := db.Del(testkv.testBytes(100))
			assert.Nil(t, err)
		}
	}

	stats := db.Stats()
	t.Log(fmt.Sprintf("%+v", stats))
	assert.Greater(t, stats.KeyNums, int64(0))
	assert.Greater(t, stats.RecordNums, int64(0))
	assert.GreaterOrEqual(t, stats.RecordNums, stats.KeyNums)
	assert.Greater(t, stats.DataSize, int64(0))
	assert.GreaterOrEqual(t, stats.HintSize, int64(0))
}
