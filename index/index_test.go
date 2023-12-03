package index

import (
	"bytes"
	"fmt"
	entry2 "github.com/246859/river/entry"
	"github.com/246859/river/wal"
	"github.com/stretchr/testify/assert"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestIndexer(t *testing.T) {

	type Sample struct {
		newIndex func() Index
		name     string
	}

	samples := []Sample{
		{
			newIndex: func() Index {
				return BtreeIndex(32, DefaultCompare)
			}, name: "btree",
		},
		{
			newIndex: func() Index {
				return BtreeNIndex(32, DefaultCompare)
			}, name: "btreeN",
		},
	}

	for _, sample := range samples {
		t.Run(fmt.Sprintf("%s_index_get", sample.name), func(t *testing.T) {
			testIndexer_Get(t, sample.newIndex())
		})

		t.Run(fmt.Sprintf("%s_index_put", sample.name), func(t *testing.T) {
			testIndexer_Put(t, sample.newIndex())
		})

		t.Run(fmt.Sprintf("%s_index_del", sample.name), func(t *testing.T) {
			testIndexer_Del(t, sample.newIndex())
		})

		t.Run(fmt.Sprintf("%s_index_iterate", sample.name), func(t *testing.T) {
			testIndexer_Iterator(t, sample.newIndex())
		})
	}
}

func testIndexer_Get(t *testing.T, indexer Index) {
	bar := Hint{Key: []byte("bar"), ChunkPos: wal.ChunkPos{Fid: 1, Offset: 2}}
	foo := Hint{Key: []byte("foo"), ChunkPos: wal.ChunkPos{Fid: 4, Offset: 5}}
	bob := Hint{Key: []byte("bob"), ChunkPos: wal.ChunkPos{Fid: 7, Offset: 8}}
	nilE := Hint{Key: nil}

	indexer.Put(bar)
	indexer.Put(foo)

	getBar, eBar := indexer.Get(bar.Key)
	assert.True(t, eBar)
	assert.Equal(t, bar, getBar)

	getFoo, eFoo := indexer.Get(foo.Key)
	assert.True(t, eFoo)
	assert.Equal(t, foo, getFoo)

	getBob, eBob := indexer.Get(bob.Key)
	assert.False(t, eBob)
	assert.Nil(t, getBob.Key)

	get, b := indexer.Get(nilE.Key)
	assert.Nil(t, get.Key)
	assert.False(t, b)
}

func testIndexer_Put(t *testing.T, indexer Index) {
	bar := Hint{Key: []byte("bar"), ChunkPos: wal.ChunkPos{Fid: 1, Offset: 2}}
	foo := Hint{Key: []byte("foo"), ChunkPos: wal.ChunkPos{Fid: 4, Offset: 5}}
	nilE := Hint{Key: nil}

	err := indexer.Put(bar)
	assert.Nil(t, err)

	err = indexer.Put(foo)
	assert.Nil(t, err)

	err = indexer.Put(foo)
	assert.Nil(t, err)

	err = indexer.Put(nilE)
	assert.ErrorIs(t, err, entry2.ErrNilKey)
}

func testIndexer_Del(t *testing.T, indexer Index) {
	bar := Hint{Key: []byte("bar"), ChunkPos: wal.ChunkPos{Fid: 1, Offset: 2}}
	foo := Hint{Key: []byte("foo"), ChunkPos: wal.ChunkPos{Fid: 4, Offset: 5}}
	bob := Hint{Key: []byte("bob"), ChunkPos: wal.ChunkPos{Fid: 7, Offset: 8}}

	err := indexer.Put(bar)
	assert.Nil(t, err)
	err = indexer.Put(foo)
	assert.Nil(t, err)

	err = indexer.Del(bar.Key)
	assert.Nil(t, err)

	err = indexer.Del(foo.Key)
	assert.Nil(t, err)

	err = indexer.Del(bob.Key)
	assert.Nil(t, err)
}

func testIndexer_Iterator(t *testing.T, indexer Index) {
	hints := []Hint{
		{Key: []byte("bob"), ChunkPos: wal.ChunkPos{Fid: 1, Offset: 2}},
		{Key: []byte("jack"), ChunkPos: wal.ChunkPos{Fid: 4, Offset: 5}},
		{Key: []byte("aaa"), ChunkPos: wal.ChunkPos{Fid: 7, Offset: 8}},
		{Key: []byte("adas"), ChunkPos: wal.ChunkPos{Fid: 10, Offset: 11}},
	}

	for _, h := range hints {
		err := indexer.Put(h)
		assert.Nil(t, err)
	}

	sort.Slice(hints, func(i, j int) bool {
		return bytes.Compare(hints[i].Key, hints[j].Key) < 0
	})

	iterator, err := indexer.Iterator(RangeOption{})
	assert.Nil(t, err)
	assert.NotNil(t, iterator)

	var i int
	for ; iterator.HasNext(); iterator.Next() {
		assert.EqualValues(t, hints[i].Key, iterator.Hint().Key)
		assert.EqualValues(t, hints[i].Block, iterator.Hint().Block)
		assert.EqualValues(t, hints[i].Fid, iterator.Hint().Fid)
		assert.EqualValues(t, hints[i].Offset, iterator.Hint().Offset)
		assert.EqualValues(t, hints[i].TTL, iterator.Hint().TTL)
		i++
	}
}

func TestHintMarshal_UnMarshal(t *testing.T) {
	hints := []Hint{
		{Key: []byte("foo"), TTL: 1, ChunkPos: wal.ChunkPos{Fid: 1, Block: 1, Offset: 1}},
		{Key: []byte(""), TTL: time.Now().UnixMilli(), ChunkPos: wal.ChunkPos{Fid: 1, Block: 1, Offset: 1}},
		{Key: []byte("abc"), TTL: -1, ChunkPos: wal.ChunkPos{Fid: 1, Block: 1, Offset: 1}},
		{Key: []byte("kkk"), TTL: 1, ChunkPos: wal.ChunkPos{Fid: 1, Block: 1, Offset: 1}},
		{Key: []byte(strings.Repeat("a", 1000)), TTL: 1, ChunkPos: wal.ChunkPos{Fid: 1, Block: 1, Offset: 1}},
		{Key: []byte("bob"), TTL: 1, ChunkPos: wal.ChunkPos{Fid: 3, Block: 6, Offset: 200}},
	}

	for _, hint := range hints {
		rawdata := MarshalHint(hint)
		assert.NotNil(t, rawdata)

		marshalHint := UnMarshalHint(rawdata)
		assert.Equal(t, hint, marshalHint)
	}
}
