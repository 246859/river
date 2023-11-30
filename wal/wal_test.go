package wal

import (
	"github.com/246859/river/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path"
	"strings"
	"testing"
)

var test_option = Option{
	DataDir: path.Join(os.TempDir(), DefaultWalSuffix),
	// max file 5 MB
	MaxFileSize:    types.MB * 5,
	Ext:            DefaultWalSuffix,
	BlockCache:     5,
	FsyncPerWrite:  true,
	FsyncThreshold: MaxBlockSize * 20,
}

func tempWal(opt Option) *Wal {
	wal, err := Open(opt)
	if err != nil {
		panic(err)
	}
	return wal
}

func clean(wal *Wal) {
	if err := wal.Purge(); err != nil {
		panic(err)
	}
	if err := wal.Close(); err != nil {
		panic(err)
	}
	if err := os.RemoveAll(wal.option.DataDir); err != nil {
		panic(err)
	}
}

func TestWal_Load(t *testing.T) {
	t.TempDir()
	wal := tempWal(test_option)
	assert.True(t, wal.IsEmpty())
	defer clean(wal)
}

func TestWal_Purge(t *testing.T) {
	wal := tempWal(test_option)
	defer clean(wal)

	// write
	data := []byte("hello world!")
	pos, err := wal.Write(data)
	assert.Nil(t, err)
	assert.EqualValues(t, pos.Block, 0)

	err = wal.Purge()
	assert.Nil(t, err)

	data1 := []byte("hello world1")
	pos1, err := wal.Write(data1)
	assert.Nil(t, err)
	assert.EqualValues(t, pos1.Block, 0)
}

func TestWal_Write_SingleFile(t *testing.T) {
	wal := tempWal(test_option)
	defer clean(wal)

	// write
	data := []byte("hello world!")
	pos, err := wal.Write(data)
	assert.Nil(t, err)
	assert.Greater(t, pos.Size, int64(0))

	data1 := []byte("hello world1")
	pos1, err := wal.Write(data1)
	assert.Nil(t, err)
	assert.Greater(t, pos1.Size, int64(0))

	data2 := []byte("hello world!2")
	pos2, err := wal.Write(data2)
	assert.Nil(t, err)
	assert.Greater(t, pos1.Size, int64(0))

	data3 := []byte("hello world!3")
	pos3, err := wal.Write(data3)
	assert.Nil(t, err)
	assert.Greater(t, pos1.Size, int64(0))

	data4 := []byte("hello world!4")
	pos4, err := wal.Write(data4)
	assert.Nil(t, err)
	assert.Greater(t, pos1.Size, int64(0))

	// read
	read, err := wal.Read(pos)
	assert.Nil(t, err)
	assert.EqualValues(t, data, read)

	read1, err := wal.Read(pos1)
	assert.Nil(t, err)
	assert.EqualValues(t, data1, read1)

	read2, err := wal.Read(pos2)
	assert.Nil(t, err)
	assert.EqualValues(t, data2, read2)

	read3, err := wal.Read(pos3)
	assert.Nil(t, err)
	assert.EqualValues(t, data3, read3)

	read4, err := wal.Read(pos4)
	assert.Nil(t, err)
	assert.EqualValues(t, data4, read4)
}

func TestWal_Write_ManyFiles(t *testing.T) {
	// max file size 5 MB
	wal := tempWal(test_option)
	defer clean(wal)

	data1 := []byte(strings.Repeat("a", types.MB*2))
	pos1, err := wal.Write(data1)
	assert.Nil(t, err)
	assert.Greater(t, pos1.Size, int64(0))

	data2 := []byte(strings.Repeat("a", types.MB*4))
	pos2, err := wal.Write(data2)
	assert.Nil(t, err)
	assert.NotEqualValues(t, pos1.Fid, pos2.Fid)

	read1, err := wal.Read(pos1)
	assert.Nil(t, err)
	assert.EqualValues(t, data1, read1)

	read2, err := wal.Read(pos2)
	assert.Nil(t, err)
	assert.EqualValues(t, data2, read2)

	// write exceed data
	// 6 Mb > max file size
	data3 := []byte(strings.Repeat("a", types.MB*6))
	pos, err := wal.Write(data3)
	assert.ErrorIs(t, err, ErrDataExceedFile)
	assert.EqualValues(t, 0, pos.Fid)
}

func TestWal_Iterator(t *testing.T) {
	wal := tempWal(test_option)
	defer clean(wal)

	datas := [][]byte{
		[]byte(strings.Repeat("a", types.KB)),
		[]byte(strings.Repeat("a", types.KB*200)),
		[]byte(strings.Repeat("a", types.KB*300)),
		[]byte(strings.Repeat("a", types.KB*400)),
		[]byte(strings.Repeat("a", types.KB*500)),
		[]byte(strings.Repeat("a", types.KB*600)),
		[]byte(strings.Repeat("a", types.KB*700)),
		[]byte(strings.Repeat("a", types.KB*800)),
		[]byte(strings.Repeat("a", types.KB*900)),
		[]byte(strings.Repeat("a", types.KB*1000)),
		[]byte(strings.Repeat("a", types.KB*1100)),
		[]byte(strings.Repeat("a", types.KB*1200)),
		[]byte(strings.Repeat("a", types.KB*1300)),
		[]byte(strings.Repeat("a", types.KB*1400)),
	}

	var chunkPos []ChunkPos
	for _, data := range datas {
		pos, err := wal.Write(data)
		assert.Nil(t, err)
		chunkPos = append(chunkPos, pos)
	}

	// normal use
	{
		it, err := wal.Iterator(0, wal.ActiveFid(), ChunkPos{})
		assert.Nil(t, err)

		var i int
		for {
			data, pos, err := it.NextData()
			if errors.Is(err, io.EOF) {
				break
			} else {
				assert.Nil(t, err)
			}

			assert.EqualValues(t, datas[i], data)
			assert.EqualValues(t, chunkPos[i].Block, pos.Block)
			assert.EqualValues(t, chunkPos[i].Offset, pos.Offset)
			i++
		}
	}

	// specified start position
	{
		assert.Greater(t, len(chunkPos), 5)

		it, err := wal.Iterator(0, wal.ActiveFid(), chunkPos[4])
		assert.NotEqualValues(t, 0, it.IndexPos().Block)
		assert.Nil(t, err)

		var i = 4
		for {
			data, pos, err := it.NextData()
			if errors.Is(err, io.EOF) {
				break
			} else {
				assert.Nil(t, err)
			}

			t.Log(it.IndexFid())
			t.Log(it.IndexPos())
			assert.EqualValues(t, datas[i], data)
			assert.EqualValues(t, chunkPos[i].Block, pos.Block)
			assert.EqualValues(t, chunkPos[i].Offset, pos.Offset)
			i++
		}
	}
}
