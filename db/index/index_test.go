package index

import (
	"github.com/246859/river/db/data"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIndexer(t *testing.T) {
	indexers := []func() Indexer{
		func() Indexer {
			return IdxBtree(32)
		},
	}

	for _, indexer := range indexers {
		testIndexer_Get(t, indexer())
		testIndexer_Put(t, indexer())
		testIndexer_Del(t, indexer())
		testIndexer_Iterator(t, indexer())
	}
}

func testIndexer_Get(t *testing.T, indexer Indexer) {
	bar := HintEntry{Key: []byte("bar"), Pos: data.EntryHint{Fid: 1, Offset: 2, Size: 3}}
	foo := HintEntry{Key: []byte("foo"), Pos: data.EntryHint{Fid: 4, Offset: 5, Size: 6}}
	bob := HintEntry{Key: []byte("bob"), Pos: data.EntryHint{Fid: 7, Offset: 8, Size: 9}}
	nilE := HintEntry{Key: nil}

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

func testIndexer_Put(t *testing.T, indexer Indexer) {
	bar := HintEntry{Key: []byte("bar"), Pos: data.EntryHint{Fid: 1, Offset: 2, Size: 3}}
	foo := HintEntry{Key: []byte("foo"), Pos: data.EntryHint{Fid: 4, Offset: 5, Size: 6}}
	nilE := HintEntry{Key: nil}

	oldBar, errBar := indexer.Put(bar)
	assert.Nil(t, oldBar.Key)
	assert.Nil(t, errBar)

	oldFoo, errFoo := indexer.Put(foo)
	assert.Nil(t, oldFoo.Key)
	assert.Nil(t, errFoo)

	oldFoo, errFoo = indexer.Put(foo)
	assert.NotNil(t, oldFoo.Key)
	assert.Nil(t, errFoo)
	assert.Equal(t, foo.Pos, oldFoo.Pos)

	oldNil, errNil := indexer.Put(nilE)
	assert.Nil(t, oldNil.Key)
	assert.Equal(t, data.ErrNilKey, errNil)
}

func testIndexer_Del(t *testing.T, indexer Indexer) {
	bar := HintEntry{Key: []byte("bar"), Pos: data.EntryHint{Fid: 1, Offset: 2, Size: 3}}
	foo := HintEntry{Key: []byte("foo"), Pos: data.EntryHint{Fid: 4, Offset: 5, Size: 6}}
	bob := HintEntry{Key: []byte("bob"), Pos: data.EntryHint{Fid: 7, Offset: 8, Size: 9}}

	indexer.Put(bar)
	indexer.Put(foo)

	oldBar, eBar := indexer.Del(bar.Key)
	assert.True(t, eBar)
	assert.Equal(t, bar, oldBar)

	oldFoo, eFoo := indexer.Del(foo.Key)
	assert.True(t, eFoo)
	assert.Equal(t, foo, oldFoo)

	oldBob, eBob := indexer.Del(bob.Key)
	assert.False(t, eBob)
	assert.Nil(t, oldBob.Key)
}

func testIndexer_Iterator(t *testing.T, indexer Indexer) {
	entries := []HintEntry{
		{Key: []byte("bar"), Pos: data.EntryHint{Fid: 1, Offset: 2, Size: 3}},
		{Key: []byte("foo"), Pos: data.EntryHint{Fid: 4, Offset: 5, Size: 6}},
		{Key: []byte("bob"), Pos: data.EntryHint{Fid: 7, Offset: 8, Size: 9}},
		{Key: []byte("jack"), Pos: data.EntryHint{Fid: 10, Offset: 11, Size: 12}},
	}

	for _, entry := range entries {
		indexer.Put(entry)
	}

	it := indexer.Iterator(true)

	for entry := it.Next(); entry.Key != nil; {
		assert.NotNil(t, entry.Key)
	}

	dit := indexer.Iterator(false)

	for entry := dit.Next(); entry.Key != nil; {
		assert.NotNil(t, entry.Key)
	}

	assert.Nil(t, it.Next().Key)
}
