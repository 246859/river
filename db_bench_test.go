package riverdb

import (
	"github.com/246859/river/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testDB_Bench_Get(b *testing.B, datanum int) {

	db, closeDB, err := testDB(b.Name(), DefaultOptions)
	if err != nil {
		panic(err)
	}
	testkv := testRandKV()

	b.Cleanup(func() {
		if err := closeDB(); err != nil {
			panic(err)
		}
	})

	var samples []Record
	for i := 0; i < datanum; i++ {
		samples = append(samples, Record{
			K:   testkv.testUniqueBytes(100),
			V:   testkv.testBytes(types.KB),
			TTL: 0,
		})
	}

	batch, err := db.Batch(BatchOption{
		Size:        int64(datanum / 20),
		SyncOnFlush: true,
	})
	assert.Nil(b, err)

	err = batch.WriteAll(samples)
	assert.Nil(b, err)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := samples[i%datanum]
		_, err := db.Get(s.K)
		assert.Nil(b, err)
	}
}

func testDB_Bench_Put(b *testing.B, datanum int, datasize int) {
	db, closeDB, err := testDB(b.Name(), DefaultOptions)
	if err != nil {
		panic(err)
	}
	testkv := testRandKV()

	b.Cleanup(func() {
		if err := closeDB(); err != nil {
			panic(err)
		}
	})

	var samples []Record
	for i := 0; i < datanum; i++ {
		samples = append(samples, Record{
			K:   testkv.testUniqueBytes(100),
			V:   testkv.testBytes(datasize),
			TTL: 0,
		})
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i := 0; i < datanum; i++ {
			err := db.Put(samples[i].K, samples[i].V, samples[i].TTL)
			assert.Nil(b, err)
		}
	}
}

func BenchmarkDB_Get_1k(b *testing.B) {
	testDB_Bench_Get(b, 1000)
}

func BenchmarkDB_Get_1w(b *testing.B) {
	testDB_Bench_Get(b, 1_0000)
}

func BenchmarkDB_Get_10w(b *testing.B) {
	testDB_Bench_Get(b, 10_0000)
}

func BenchmarkDB_Get_100w(b *testing.B) {
	testDB_Bench_Get(b, 100_0000)
}

// data num test

func BenchmarkDb_Put_1(b *testing.B) {
	testDB_Bench_Put(b, 1, types.KB)
}

func BenchmarkDb_Put_100(b *testing.B) {
	testDB_Bench_Put(b, 100, types.KB)
}

func BenchmarkDb_Put_1K(b *testing.B) {
	testDB_Bench_Put(b, 1000, types.KB)
}

func BenchmarkDb_Put_1W(b *testing.B) {
	testDB_Bench_Put(b, 1_0000, types.KB)
}

func BenchmarkDb_Put_10W(b *testing.B) {
	testDB_Bench_Put(b, 10_0000, types.KB)
}

// data size test

func BenchmarkDb_Put_256B(b *testing.B) {
	testDB_Bench_Put(b, 1, types.B*256)
}

func BenchmarkDb_Put_64KB(b *testing.B) {
	testDB_Bench_Put(b, 1, types.KB*64)
}

func BenchmarkDb_Put_256KB(b *testing.B) {
	testDB_Bench_Put(b, 1, types.KB*256)
}

func BenchmarkDb_Put_1MB(b *testing.B) {
	testDB_Bench_Put(b, 1, types.MB)
}

func BenchmarkDb_Put_4MB(b *testing.B) {
	testDB_Bench_Put(b, 1, types.MB)
}

func BenchmarkDb_Put_8MB(b *testing.B) {
	testDB_Bench_Put(b, 1, types.MB)
}
