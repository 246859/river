package index

import (
	"bytes"
	"github.com/pkg/errors"
)

var (
	ErrNilIterator = errors.New("nil iterator")
	ErrClosed      = errors.New("index closed")
)

type Key = []byte

// RangeOption iterate over all keys in [min, max] with give order
// if min and max are nil, it will return all the keys in index,
// if min is nil and max is not nil, it will return the keys compare or equal than max,
// if max is nil and min is not nil, it will return the keys greater or equal than min,
// both of them are not nil, it will return the keys in range [min, max],
// then filter the keys with the give pattern if it is not empty.
// finally return these keys with the given order.
type RangeOption struct {
	// min key
	Min Key
	// max key
	Max Key
	// pattern matching
	Pattern Key
	// return order of elements, default is ascending
	Descend bool
}

const (
	Less = -1 + iota
	Equal
	Greater
)

// Compare returns a function that decide how to compare keys
type Compare func(a, b Key) int

func DefaultCompare(a, b Key) int {
	return bytes.Compare(a, b)
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
	// Compare returns -1-less, 0-equal, 1-greater
	Compare(k1, k2 Key) int
}

// Iterator record snapshot of index hints at a certain moment, iterate them in the given order by cursor
// depending on the implementation of indexer, the behavior of iterator may be different, like hashmap or btree
type Iterator interface {
	// Rewind set cursor to the head of hints snapshots
	Rewind()
	// Next returns the Hint on the cursor, return true if it is out of range
	Next() (Hint, bool)
	// Close releases the resources and close iterator
	Close() error
}
