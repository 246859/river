package index

import (
	"bytes"
	"github.com/246859/river/db/entry"
)

type Key = []byte

// RangeOption range [min, max] keys option
type RangeOption struct {
	// min key
	Min Key
	// max key
	max Key

	// pattern matching
	pattern Key

	// return order of elements
	Descend bool
}

// Hint represents a kev hint data
type Hint struct {
	Key  Key
	Hint entry.Hint
}

// Compare
// -1-less, 0-equal, 1-larger
func (i Hint) Compare(idx Hint) int {
	return bytes.Compare(i.Key, idx.Key)
}

// Index storage a set of index hints and define index operation apis that implemented by concrete data struct.
type Index interface {
	// Get returns the entry matching the given key
	// if not exist, returns zero-value, false
	Get(key Key) (Hint, bool)
	// Put inserts a new entry into the index, replace it if already exists
	Put(entry Hint) error
	// Del deletes the entry that matching the given key from the index, and returns true
	// if not exist, returns false
	Del(key Key) (bool, error)
	// Size return num of all hints in index
	Size() int
	// Iterator returns an iterator of index snapshots at a certain moment
	Iterator(opt RangeOption) (Iterator, error)
	// Close releases the resources and close indexer
	Close() error
}

// Iterator record snapshot of index hints at a certain moment, iterate them in the given order by cursor
// depending on the implementation of indexer, the behavior of iterator may be different, like hashmap or btree
type Iterator interface {
	// Rewind set cursor to the head of hints snapshots
	Rewind()
	// Next returns the Hint on the cursor, return false if it has no next hint
	Next() (Hint, bool)
	// Close releases the resources and close iterator
	Close() error
}
