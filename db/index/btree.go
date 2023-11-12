package index

import (
	"github.com/246859/river/db/data"
	"github.com/google/btree"
	"sort"
	"sync"
)

var (
	_ Indexer  = &BTree{}
	_ Iterator = &BTreeIterator{}
)

func BtreeIdx(degree int) *BTree {
	bi := new(BTree)
	bi.tree = btree.NewG[HintEntry](degree, func(a, b HintEntry) bool {
		return a.Compare(b) < 0
	})
	return bi
}

// BTree is btree implementation of Indexer that
// allows the actions of finding data, sequential access, inserting data, and deleting to be done in O(log n) time
type BTree struct {
	tree  *btree.BTreeG[HintEntry]
	mutex sync.RWMutex
}

func (b *BTree) Iterator(asc bool) Iterator {
	b.mutex.RLock()
	iterator := NewBTreeIterator(b, asc)
	b.mutex.RUnlock()
	return iterator
}

func (b *BTree) Get(key Key) (HintEntry, bool) {
	if key == nil {
		return HintEntry{}, false
	}

	idx := HintEntry{Key: key}
	oldIdx, exist := b.tree.Get(idx)
	return oldIdx, exist
}

func (b *BTree) Put(entry HintEntry) (HintEntry, error) {
	if entry.Key == nil {
		return HintEntry{}, data.ErrNilKey
	}
	b.mutex.Lock()
	oldIdxe, _ := b.tree.ReplaceOrInsert(entry)
	b.mutex.Unlock()
	return oldIdxe, nil
}

func (b *BTree) Del(key Key) (HintEntry, bool) {
	if key == nil {
		return HintEntry{}, false
	}

	idxe := HintEntry{Key: key}
	b.mutex.Lock()
	oldIdxe, exist := b.tree.Delete(idxe)
	b.mutex.Unlock()
	return oldIdxe, exist
}

func (b *BTree) Size() int {
	return b.tree.Len()
}

func (b *BTree) Close() error {
	return nil
}

func NewBTreeIterator(btr *BTree, asc bool) Iterator {

	var idx int
	entries := make([]HintEntry, btr.Size())

	itFn := func(item HintEntry) bool {
		entries[idx] = item
		idx++
		return true
	}

	if asc {
		btr.tree.Ascend(itFn)
	} else {
		btr.tree.Descend(itFn)
	}

	return &BTreeIterator{
		entries: entries,
		asc:     asc,
		cursor:  -1,
	}
}

type BTreeIterator struct {
	entries []HintEntry
	asc     bool
	cursor  int
}

func (bi *BTreeIterator) Next() HintEntry {
	if bi.cursor < len(bi.entries) {
		bi.cursor++
	}
	bi.cursor = len(bi.entries)
	return bi.Entry()
}

func (bi *BTreeIterator) Rewind() {
	bi.cursor = 0
}

func (bi *BTreeIterator) Seek(key Key) {
	n := len(bi.entries)
	searchIdx := HintEntry{Key: key}
	bi.cursor = sort.Search(n, func(i int) bool {
		return (bi.asc && bi.entries[i].Compare(searchIdx) <= 0) || (!bi.asc && bi.entries[i].Compare(searchIdx) >= 0)
	})
}

func (bi *BTreeIterator) Entry() HintEntry {
	if 0 <= bi.cursor && bi.cursor < len(bi.entries) {
		return bi.entries[bi.cursor]
	}
	return HintEntry{}
}

func (bi *BTreeIterator) Close() error {
	return nil
}
