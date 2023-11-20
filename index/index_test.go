package index

import (
	"bytes"
	entry2 "github.com/246859/river/entry"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestIndexer(t *testing.T) {
	indexers := []func() Index{
		func() Index {
			return BtreeIndex(32, DefaultCompare)
		},
	}

	for _, indexer := range indexers {
		testIndexer_Get(t, indexer())
		testIndexer_Put(t, indexer())
		testIndexer_Del(t, indexer())
		testIndexer_Iterator(t, indexer())
	}
}

func testIndexer_Get(t *testing.T, indexer Index) {
	bar := Hint{Key: []byte("bar"), Hint: entry2.Hint{Fid: 1, Offset: 2}}
	foo := Hint{Key: []byte("foo"), Hint: entry2.Hint{Fid: 4, Offset: 5}}
	bob := Hint{Key: []byte("bob"), Hint: entry2.Hint{Fid: 7, Offset: 8}}
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
	bar := Hint{Key: []byte("bar"), Hint: entry2.Hint{Fid: 1, Offset: 2}}
	foo := Hint{Key: []byte("foo"), Hint: entry2.Hint{Fid: 4, Offset: 5}}
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
	bar := Hint{Key: []byte("bar"), Hint: entry2.Hint{Fid: 1, Offset: 2}}
	foo := Hint{Key: []byte("foo"), Hint: entry2.Hint{Fid: 4, Offset: 5}}
	bob := Hint{Key: []byte("bob"), Hint: entry2.Hint{Fid: 7, Offset: 8}}

	err := indexer.Put(bar)
	assert.Nil(t, err)
	err = indexer.Put(foo)
	assert.Nil(t, err)

	e, err := indexer.Del(bar.Key)
	assert.True(t, e)
	assert.Nil(t, err)

	e, err = indexer.Del(foo.Key)
	assert.True(t, e)
	assert.Nil(t, err)

	e, err = indexer.Del(bob.Key)
	assert.False(t, e)
	assert.Nil(t, err)
}

func testIndexer_Iterator(t *testing.T, indexer Index) {
	hints := []Hint{
		{Key: []byte("bob"), Hint: entry2.Hint{Fid: 1, Offset: 2}},
		{Key: []byte("jack"), Hint: entry2.Hint{Fid: 4, Offset: 5}},
		{Key: []byte("aaa"), Hint: entry2.Hint{Fid: 7, Offset: 8}},
		{Key: []byte("adas"), Hint: entry2.Hint{Fid: 10, Offset: 11}},
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
	for {
		next, out := iterator.Next()
		if !out {
			break
		}
		assert.EqualValues(t, hints[i], next)
		i++
	}
}
