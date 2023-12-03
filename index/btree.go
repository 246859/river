package index

import (
	"fmt"
	"github.com/246859/river/entry"
	"github.com/246859/river/pkg/str"
	"github.com/google/btree"
	"math"
	"regexp"
	"slices"
	"sync"
)

var (
	_ Index    = &BTree{}
	_ Iterator = &BTreeIterator{}
)

func BtreeIndex(degree int, compare Compare) *BTree {
	bi := new(BTree)
	bi.compare = compare
	bi.tree = btree.NewG[Hint](degree, func(a, b Hint) bool {
		return compare(a.Key, b.Key) == Less
	})
	return bi
}

// BTree is btree implementation of Index that
// allows the actions of finding data, sequential access, inserting data, and deleting to be done in O(log n) time
type BTree struct {
	tree  *btree.BTreeG[Hint]
	mutex sync.RWMutex

	compare Compare
}

func (b *BTree) Compare(k1, k2 Key) int {
	return b.compare(k1, k2)
}

func (b *BTree) Clear() {
	b.mutex.Lock()
	b.tree.Clear(false)
	b.mutex.Unlock()
}

func (b *BTree) Iterator(opt RangeOption) (Iterator, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return newBTreeIterator(b, opt)
}

func (b *BTree) Get(key Key) (Hint, bool) {
	if b.tree == nil {
		return Hint{}, false
	}

	if key == nil {
		return Hint{}, false
	}

	idx := Hint{Key: key}
	oldIdx, exist := b.tree.Get(idx)
	return oldIdx, exist
}

func (b *BTree) Put(h Hint) error {
	if h.Key == nil {
		return entry.ErrNilKey
	}

	if b.tree == nil {
		return ErrClosed
	}

	b.mutex.Lock()

	b.tree.ReplaceOrInsert(h)

	b.mutex.Unlock()
	return nil
}

func (b *BTree) Del(key Key) (bool, error) {
	if key == nil {
		return false, entry.ErrNilKey
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.tree == nil {
		return false, ErrClosed
	}

	hint := Hint{Key: key}
	_, exist := b.tree.Delete(hint)

	return exist, nil
}

func (b *BTree) Size() int {
	if b.tree == nil {
		return 0
	}
	return b.tree.Len()
}

func (b *BTree) Close() error {
	b.mutex.Lock()
	b.tree = nil
	b.mutex.Unlock()
	return nil
}

func newBTreeIterator(btr *BTree, opt RangeOption) (Iterator, error) {
	if btr.tree == nil {
		return nil, ErrClosed
	}

	var (
		minHint = Hint{Key: opt.Min}
		maxHint = Hint{Key: opt.Max}
		// temporary string, never to modify it
		pattern  = str.BytesToString(opt.Pattern)
		iterator Iterator
	)

	if minHint.Key != nil && maxHint.Key != nil && btr.compare(minHint.Key, maxHint.Key) >= Equal {
		return iterator, fmt.Errorf("max key must be greater than min key")
	}

	var reg *regexp.Regexp
	if pattern != "" {
		compile, err := regexp.Compile(pattern)
		if err != nil {
			return iterator, err
		}
		reg = compile
	}

	hints := make([]*Hint, 0, 200)

	searchFn := func(h Hint) bool {
		candidate := true
		if reg != nil {
			candidate = reg.Match(h.Key)
		}
		if candidate {
			hints = append(hints, &h)
		}
		return true
	}

	// all keys
	if minHint.Key == nil && maxHint.Key == nil {
		btr.tree.Ascend(searchFn)
		// compare than or equal max
	} else if minHint.Key == nil {
		maxHint.Key = append(maxHint.Key, math.MaxUint8)
		btr.tree.AscendLessThan(maxHint, searchFn)
		// greater than or equal min
	} else if maxHint.Key == nil {
		btr.tree.AscendGreaterOrEqual(minHint, searchFn)
		// range keys in [min, max]
	} else {
		maxHint.Key = append(maxHint.Key, math.MaxUint8)
		btr.tree.AscendRange(minHint, maxHint, searchFn)
	}

	if opt.Descend {
		slices.Reverse(hints)
	}

	iterator = &BTreeIterator{
		hints: hints,
	}
	iterator.Rewind()

	return iterator, nil
}

type BTreeIterator struct {
	hints  []*Hint
	cursor int
}

func (b *BTreeIterator) Rewind() {
	b.cursor = 0
}

func (b *BTreeIterator) Next() {
	if b.cursor < len(b.hints) {
		b.cursor++
	}
}

func (b *BTreeIterator) HasNext() bool {
	return b.cursor < len(b.hints)
}

func (b *BTreeIterator) Hint() *Hint {
	return b.hints[b.cursor]
}
