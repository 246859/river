package riverdb

import (
	"github.com/246859/river/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func benchmarkDB_Put(b *testing.B, db *DB, testkv *randKV, vsize int) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := db.Put(testkv.testBytes(30), testkv.testBytes(vsize), 0)
		assert.Nil(b, err)
	}
}

func benchmarkDB_Get(b *testing.B, db *DB, testkv *randKV) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		db.Get(testkv.testBytes(30))
	}
}

func benchmarkDB_Del(b *testing.B, db *DB, testkv *randKV) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		db.Del(testkv.testBytes(30))
	}
}

func BenchmarkDB_B(b *testing.B) {
	db, closeDB, err := testDB("river-bench", DefaultOptions)
	if err != nil {
		panic(err)
	}
	testkv := testRandKV()

	b.Cleanup(func() {
		if err := closeDB(); err != nil {
			panic(err)
		}
	})

	b.Run("benchmarkDB_Put_value_64B", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.B*64)
	})

	b.Run("benchmarkDB_Put_value_128B", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.B*128)
	})

	b.Run("benchmarkDB_Put_value_256B", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.B*256)
	})

	b.Run("benchmarkDB_Put_value_512B", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.B*512)
	})

	b.Run("benchmarkDB_Get_B", func(b *testing.B) {
		benchmarkDB_Get(b, db, testkv)
	})

	b.Run("benchmarkDB_Del_B", func(b *testing.B) {
		benchmarkDB_Del(b, db, testkv)
	})
}

func BenchmarkDB_KB(b *testing.B) {
	db, closeDB, err := testDB("river-bench", DefaultOptions)
	if err != nil {
		panic(err)
	}
	testkv := testRandKV()

	b.Cleanup(func() {
		if err := closeDB(); err != nil {
			panic(err)
		}
	})

	b.Run("benchmarkDB_Put_value_1KB", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.KB)
	})

	b.Run("benchmarkDB_Put_value_64KB", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.KB*64)
	})

	b.Run("benchmarkDB_Put_value_256KB", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.KB*256)
	})

	b.Run("benchmarkDB_Put_value_512KB", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.KB*512)
	})

	b.Run("benchmarkDB_Get_KB", func(b *testing.B) {
		benchmarkDB_Get(b, db, testkv)
	})

	b.Run("benchmarkDB_Del_KB", func(b *testing.B) {
		benchmarkDB_Del(b, db, testkv)
	})
}

func BenchmarkDB_MB(b *testing.B) {
	db, closeDB, err := testDB("river-bench", DefaultOptions)
	if err != nil {
		panic(err)
	}
	testkv := testRandKV()

	b.Cleanup(func() {
		if err := closeDB(); err != nil {
			panic(err)
		}
	})

	b.Run("benchmarkDB_Put_value_1MB", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.MB)
	})

	b.Run("benchmarkDB_Put_value_8MB", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.MB*8)
	})

	b.Run("benchmarkDB_Put_value_16MB", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.MB*16)
	})

	b.Run("benchmarkDB_Put_value_32MB", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.MB*32)
	})

	b.Run("benchmarkDB_Put_value_64MB", func(b *testing.B) {
		benchmarkDB_Put(b, db, testkv, types.MB*64)
	})

	b.Run("benchmarkDB_Get_MB", func(b *testing.B) {
		benchmarkDB_Get(b, db, testkv)
	})

	b.Run("benchmarkDB_Del_MB", func(b *testing.B) {
		benchmarkDB_Del(b, db, testkv)
	})
}
