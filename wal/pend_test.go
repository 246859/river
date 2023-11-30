package wal

import (
	"github.com/246859/river/types"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestPending_Write(t *testing.T) {
	wal := tempWal(test_option)
	defer clean(wal)

	pending := wal.Pending(20)

	datas := [][]byte{
		[]byte(strings.Repeat("a", types.KB)),
		[]byte(strings.Repeat("a", types.KB*200)),
		[]byte(strings.Repeat("a", types.KB*300)),
		[]byte(strings.Repeat("a", types.KB*400)),
		[]byte(strings.Repeat("a", types.KB*500)),
		[]byte(strings.Repeat("a", types.KB*600)),
		[]byte(strings.Repeat("a", types.KB*700)),
		[]byte(strings.Repeat("a", types.KB*800)),
	}

	for _, data := range datas {
		err := pending.Write(data)
		assert.Nil(t, err)
	}

	chunkPos, err := pending.Flush(true)
	assert.Nil(t, err)
	assert.EqualValues(t, len(datas), len(chunkPos))
}

func TestPending_Out_Of_Max(t *testing.T) {
	wal := tempWal(test_option)
	defer clean(wal)

	pending := wal.Pending(20)

	err := pending.Write([]byte(strings.Repeat("a", types.KB)))
	assert.Nil(t, err)

	err = pending.Write([]byte(strings.Repeat("a", int(wal.option.MaxFileSize+1))))
	assert.NotNil(t, err)
}

func TestPending_Exceed(t *testing.T) {
	wal := tempWal(test_option)
	defer clean(wal)

	pending := wal.Pending(20)

	datas := [][]byte{
		[]byte(strings.Repeat("a", int(wal.option.MaxFileSize/5))),
		[]byte(strings.Repeat("a", int(wal.option.MaxFileSize/5))),
		[]byte(strings.Repeat("a", int(wal.option.MaxFileSize/5))),
		[]byte(strings.Repeat("a", int(wal.option.MaxFileSize/5))),
		[]byte(strings.Repeat("a", int(wal.option.MaxFileSize/5))),
		[]byte(strings.Repeat("a", int(wal.option.MaxFileSize/5))),
		[]byte(strings.Repeat("a", int(wal.option.MaxFileSize/5))),
	}

	for _, data := range datas {
		err := pending.Write(data)
		assert.Nil(t, err)
	}

	chunkPos, err := pending.Flush(true)
	assert.NotNil(t, err)
	assert.Nil(t, chunkPos)
}

func TestPending_Rotate(t *testing.T) {
	wal := tempWal(test_option)
	defer clean(wal)

	pos, err := wal.Write([]byte(strings.Repeat("a", int(wal.option.MaxFileSize/2))))
	assert.Nil(t, err)
	assert.EqualValues(t, 1, pos.Fid)

	pending := wal.Pending(20)
	err = pending.Write([]byte(strings.Repeat("a", int(wal.option.MaxFileSize*2/3))))
	assert.Nil(t, err)
	chunkPos, err := pending.Flush(true)
	assert.Nil(t, err)
	assert.Len(t, chunkPos, 1)
	assert.EqualValues(t, 2, chunkPos[0].Fid)
}
