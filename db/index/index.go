package index

import (
	"bytes"
	"github.com/246859/river/db/data"
)

type Key = []byte

// HintEntry represent an index entry in indexes struct
type HintEntry struct {
	Key Key
	Pos data.RecordPos
}

// Compare
// -1-less, 0-equal, 1-larger
func (i HintEntry) Compare(idx HintEntry) int {
	return bytes.Compare(i.Key, idx.Key)
}

// Indexer storage a set of index entries and define index operation apis that implemented by concrete data struct.
type Indexer interface {
	// Get returns the entry matching the given key
	// if not exist, returns zero-value, false
	Get(key Key) (HintEntry, bool)
	// Put inserts a new entry into the index
	// replace it if already exists, then returns old entry
	Put(entry HintEntry) HintEntry
	// Del deletes the entry that matching the given key from the index, and returns old entry
	// if not exist, returns zero-value, false
	Del(key Key) (HintEntry, bool)
	// Size return num of all entries in index
	Size() int
	// Iterator returns an iterator of index snapshots at a certain moment
	Iterator(asc bool) Iterator
	// Close releases the resources and close indexer
	Close() error
}

// Iterator record snapshot of index entries at a certain moment, iterate them in the given order by cursor
// depending on the implementation of indexer, the behavior of iterator may be different, like hashmap or btree
type Iterator interface {
	// Rewind set cursor to the head of entries snapshots
	Rewind()
	// Seek set cursor to an as close as possible position that near the given key in entries snapshots
	Seek(key Key)
	// Next move the cursor to next, return false if it has no next entry
	Next() HintEntry
	// Close releases the resources and close iterator
	Close() error
}
