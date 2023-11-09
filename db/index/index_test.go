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
	bar := Entry{Key: []byte("bar"), Pos: data.RecordPos{Fid: 1, Offset: 2, Size: 3}}
	foo := Entry{Key: []byte("foo"), Pos: data.RecordPos{Fid: 4, Offset: 5, Size: 6}}
	bob := Entry{Key: []byte("bob"), Pos: data.RecordPos{Fid: 7, Offset: 8, Size: 9}}

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
}

func testIndexer_Put(t *testing.T, indexer Indexer) {
	bar := Entry{Key: []byte("bar"), Pos: data.RecordPos{Fid: 1, Offset: 2, Size: 3}}
	foo := Entry{Key: []byte("foo"), Pos: data.RecordPos{Fid: 4, Offset: 5, Size: 6}}

	oldBar := indexer.Put(bar)
	assert.Nil(t, oldBar.Key)

	oldFoo := indexer.Put(foo)
	assert.Nil(t, oldFoo.Key)

	oldFoo = indexer.Put(foo)
	assert.NotNil(t, oldFoo.Key)
	assert.Equal(t, foo.Pos, oldFoo.Pos)
}

func testIndexer_Del(t *testing.T, indexer Indexer) {
	bar := Entry{Key: []byte("bar"), Pos: data.RecordPos{Fid: 1, Offset: 2, Size: 3}}
	foo := Entry{Key: []byte("foo"), Pos: data.RecordPos{Fid: 4, Offset: 5, Size: 6}}
	bob := Entry{Key: []byte("bob"), Pos: data.RecordPos{Fid: 7, Offset: 8, Size: 9}}

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
	entries := []Entry{
		{Key: []byte("bar"), Pos: data.RecordPos{Fid: 1, Offset: 2, Size: 3}},
		{Key: []byte("foo"), Pos: data.RecordPos{Fid: 4, Offset: 5, Size: 6}},
		{Key: []byte("bob"), Pos: data.RecordPos{Fid: 7, Offset: 8, Size: 9}},
		{Key: []byte("jack"), Pos: data.RecordPos{Fid: 10, Offset: 11, Size: 12}},
	}

	for _, entry := range entries {
		indexer.Put(entry)
	}

	it := indexer.Iterator(true)

	for entry := it.Next(); entry.Key != nil; {
		assert.NotNil(t, entry.Key)
	}

	assert.Nil(t, it.Next().Key)
}
