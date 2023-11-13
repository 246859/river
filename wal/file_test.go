package wal

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/bytebufferpool"
	"io"
	"os"
	"strings"
	"testing"
)

func newLogFile(fid uint32) *LogFile {
	q, err := lru.New2Q[uint64, []byte](32)
	if err != nil {
		panic(err)
	}
	file, err := openLogFile(os.TempDir(), "wal", fid, q, new(bytebufferpool.Pool))
	if err != nil {
		panic(err)
	}
	return file
}

func TestLogFile_Write_Full(t *testing.T) {
	{
		// write chunk full data whose size far less than block size
		file := newLogFile(1)
		defer file.Remove()

		// #1
		{
			data := []byte(strings.Repeat("r", 100))
			pos, err := file.Write(data)
			assert.Nil(t, err)
			assert.Equal(t, file.fid, pos.Fid)
			assert.True(t, pos.Block == 0)
			assert.True(t, pos.Offset == 0)
			assert.True(t, pos.Size == int64(len(data)+chunkHeaderSize))
		}

		// #2
		{
			data := []byte(strings.Repeat("x", 200))
			pos, err := file.Write(data)
			assert.Nil(t, err)
			assert.Equal(t, file.fid, pos.Fid)
			assert.True(t, pos.Block == 0)
			assert.True(t, pos.Offset > 0)
			assert.True(t, pos.Size == int64(len(data)+chunkHeaderSize))
		}
	}
}

func TestLogFile_Write_FullBlock(t *testing.T) {
	// write chunk data full of block
	file := newLogFile(1)
	defer file.Remove()
	data := []byte(strings.Repeat("a", maxBlockSize-chunkHeaderSize))
	pos, err := file.Write(data)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, pos.Block)
	assert.EqualValues(t, 0, pos.Offset)
	assert.EqualValues(t, maxBlockSize, pos.Size)

	data2 := []byte(strings.Repeat("a", 10))
	pos2, err := file.Write(data2)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, pos2.Block)
	assert.EqualValues(t, 0, pos2.Offset)
	assert.EqualValues(t, len(data2)+chunkHeaderSize, pos2.Size)
}

func TestLogFile_Write_Padding(t *testing.T) {
	file := newLogFile(1)
	defer file.Remove()
	// left block space = 6 bytes, could not hold a chunkheader which is 7 bytes
	data := []byte(strings.Repeat("a", maxBlockSize-chunkHeaderSize-6))
	pos, err := file.Write(data)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, pos.Block)
	assert.EqualValues(t, maxBlockSize-6, pos.Size)

	data1 := []byte(strings.Repeat("b", 12))
	pos1, err := file.Write(data1)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, pos1.Block)
}

func TestLogFile_Write_First_Middle_Last(t *testing.T) {
	file := newLogFile(1)
	defer file.Remove()

	// fist-last
	data := []byte(strings.Repeat("a", maxBlockSize+chunkHeaderSize))
	pos, err := file.Write(data)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, pos.Block)

	// fist-middle-last
	data1 := []byte(strings.Repeat("a", maxBlockSize*4))
	pos1, err := file.Write(data1)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, pos1.Block)
}

func TestLogFile_Read_Full(t *testing.T) {
	file := newLogFile(1)
	defer file.Remove()

	data := []byte(strings.Repeat("r", 100))
	pos, err := file.Write(data)
	assert.Nil(t, err)

	radata, err := file.Read(pos.Block, pos.Offset)
	assert.Nil(t, err)
	assert.Equal(t, data, radata)

	data1 := []byte(strings.Repeat("r", 100))
	pos1, err := file.Write(data)
	assert.Nil(t, err)

	radata1, err := file.Read(pos1.Block, pos1.Offset)
	assert.Nil(t, err)
	assert.Equal(t, data1, radata1)
}

func TestLogFile_Read_Padding(t *testing.T) {
	file := newLogFile(1)
	defer file.Remove()
	// left block space = 6 bytes, could not hold a chunkheader which is 7 bytes
	data := []byte(strings.Repeat("a", maxBlockSize-chunkHeaderSize-6))
	pos, err := file.Write(data)
	assert.Nil(t, err)

	data1 := []byte(strings.Repeat("b", 12))
	pos1, err := file.Write(data1)
	assert.Nil(t, err)

	read, err := file.Read(pos.Block, pos.Offset)
	assert.Nil(t, err)
	assert.Equal(t, data, read)

	read1, err := file.Read(pos1.Block, pos1.Offset)
	assert.Nil(t, err)
	assert.Equal(t, data1, read1)
}

func TestLogFile_WriteAll(t *testing.T) {
	file := newLogFile(1)
	defer file.Remove()

	datas := [][]byte{
		// full
		[]byte(strings.Repeat("a", 100)),
		// first-last
		[]byte(strings.Repeat("a", maxBlockSize)),
		// full
		[]byte(strings.Repeat("a", 100)),
		// first-middle-last
		[]byte(strings.Repeat("a", maxBlockSize*3+200)),
		// first-last
		[]byte(strings.Repeat("a", maxBlockSize)),
		// full
		[]byte(strings.Repeat("a", 100)),
	}

	poses, err := file.WriteAll(datas)
	assert.Nil(t, err)

	for i, pos := range poses {
		readi, err := file.Read(pos.Block, pos.Offset)
		assert.Nil(t, err)
		assert.EqualValues(t, datas[i], readi)
	}
}

func TestLogFile_Read_First_Middle_Last(t *testing.T) {
	file := newLogFile(1)
	defer file.Remove()

	// fist-last
	data := []byte(strings.Repeat("a", maxBlockSize+chunkHeaderSize))
	pos, err := file.Write(data)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, pos.Block)

	read, err := file.Read(pos.Block, pos.Offset)
	assert.Nil(t, err)
	assert.Equal(t, data, read)

	// fist-middle-last
	data1 := []byte(strings.Repeat("a", maxBlockSize*4))
	pos1, err := file.Write(data1)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, pos1.Block)

	read1, err := file.Read(pos1.Block, pos1.Offset)
	assert.Nil(t, err)
	assert.Equal(t, data1, read1)
}

func TestLogFile_Write_Read_ManyChunks(t *testing.T) {
	file := newLogFile(1)
	defer file.Remove()

	// fist-last
	data := []byte(strings.Repeat("a", maxBlockSize+100))
	pos, err := file.Write(data)
	assert.Nil(t, err)

	// full
	data1 := []byte("hello world!")
	pos1, err := file.Write(data1)
	assert.Nil(t, err)

	// fist-mioddle-last
	data2 := []byte(strings.Repeat("a", maxBlockSize*3+100))
	pos2, err := file.Write(data2)
	assert.Nil(t, err)

	// full
	data3 := []byte("hello world!")
	pos3, err := file.Write(data3)
	assert.Nil(t, err)

	read, err := file.Read(pos.Block, pos.Offset)
	assert.Nil(t, err)
	assert.Equal(t, data, read)

	read1, err := file.Read(pos1.Block, pos1.Offset)
	assert.Nil(t, err)
	assert.Equal(t, data1, read1)

	read2, err := file.Read(pos2.Block, pos2.Offset)
	assert.Nil(t, err)
	assert.Equal(t, data2, read2)

	read3, err := file.Read(pos3.Block, pos3.Offset)
	assert.Nil(t, err)
	assert.Equal(t, data3, read3)
}

func TestLogFile_Iterator(t *testing.T) {
	file := newLogFile(1)
	defer file.Remove()

	datas := [][]byte{
		// full
		[]byte(strings.Repeat("a", 100)),
		// first-last
		[]byte(strings.Repeat("a", maxBlockSize)),
		// full
		[]byte(strings.Repeat("a", 100)),
		// first-middle-last
		[]byte(strings.Repeat("a", maxBlockSize*3+200)),
		// first-last
		[]byte(strings.Repeat("a", maxBlockSize)),
		// full
		[]byte(strings.Repeat("a", 100)),
	}

	poses, err := file.WriteAll(datas)
	assert.Nil(t, err)

	it, err := file.Iterator()
	assert.Nil(t, err)
	assert.NotNil(t, it)

	i := 0
	for {
		dataBytes, pos, err := it.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		assert.Nil(t, err)
		assert.EqualValues(t, poses[i].Block, pos.Block)
		assert.EqualValues(t, poses[i].Offset, pos.Offset)
		assert.EqualValues(t, datas[i], dataBytes)
		i++
	}
}

func TestLogFile_BigData(t *testing.T) {

	samples := [][2]int{
		{100, 1000},
		{100, maxBlockSize},
		{100, maxBlockSize * 32},
		{1000, maxBlockSize * 64},
		{10_0000, 10},
	}

	for _, sample := range samples {
		file := newLogFile(1)

		count := sample[0]
		size := sample[1]

		data := []byte(strings.Repeat("a", size))

		var poses []ChunkPos
		for i := 0; i < count; i++ {
			pos, err := file.Write(data)
			assert.Nil(t, err)
			poses = append(poses, pos)
		}

		it, err := file.Iterator()
		assert.Nil(t, err)

		var i int
		for {
			bytes, pos, err := it.Next()
			if errors.Is(err, io.EOF) {
				break
			}
			assert.Nil(t, err)
			assert.EqualValues(t, data, bytes)

			assert.EqualValues(t, poses[i].Block, pos.Block)
			assert.EqualValues(t, poses[i].Offset, pos.Offset)
			i++
		}

		file.Remove()
	}
}
