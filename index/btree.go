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
	_ Index    = &BTreeN{}
	_ Iterator = &BTreeIterator{}
)

func BtreeNIndex(degree int, compare Compare) *BTreeN {
	bi := new(BTreeN)
	bi.compare = compare
	bi.tree = btree.NewG[Hint](degree, func(a, b Hint) bool {
		return compare(a.Key, b.Key) == Less
	})
	return bi
}

// BTreeN btree with no lock
type BTreeN struct {
	tree    *btree.BTreeG[Hint]
	compare Compare
}

func (b *BTreeN) Compare(k1, k2 Key) int {
	return b.compare(k1, k2)
}

func (b *BTreeN) Clear() {
	b.tree.Clear(false)
}

func (b *BTreeN) Iterator(opt RangeOption) (Iterator, error) {
	return newBTreeIterator(b, opt)
}

func (b *BTreeN) Get(key Key) (Hint, bool) {
	if key == nil {
		return Hint{}, false
	}

	if b.tree == nil {
		return Hint{}, false
	}

	idx := Hint{Key: key}
	oldIdx, exist := b.tree.Get(idx)
	return oldIdx, exist
}

func (b *BTreeN) Put(hs ...Hint) error {
	if len(hs) == 0 {
		return nil
	}

	if b.tree == nil {
		return ErrClosed
	}

	for _, h := range hs {
		if h.Key == nil {
			return entry.ErrNilKey
		}
		b.tree.ReplaceOrInsert(h)
	}
	return nil
}

func (b *BTreeN) Del(ks ...Key) error {
	if len(ks) == 0 {
		return nil
	}

	if b.tree == nil {
		return ErrClosed
	}

	for _, k := range ks {
		if k == nil {
			continue
		}
		hint := Hint{Key: k}
		b.tree.Delete(hint)
	}

	return nil
}

func (b *BTreeN) Size() int {
	if b.tree == nil {
		return 0
	}
	return b.tree.Len()
}

func (b *BTreeN) Close() error {
	b.tree.Clear(false)
	b.tree = nil
	return nil
}

func newBTreeIterator(btr *BTreeN, opt RangeOption) (Iterator, error) {
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

func BtreeIndex(degree int, compare Compare) *BTree {
	bi := new(BTree)
	bi.mutex = sync.RWMutex{}
	bi.tree = BtreeNIndex(degree, compare)
	return bi
}

// BTree is btree implementation of Index that
// allows the actions of finding data, sequential access, inserting data, and deleting to be done in O(log n) time
type BTree struct {
	mutex sync.RWMutex
	tree  *BTreeN
}

func (b *BTree) Compare(k1, k2 Key) int {
	return b.tree.compare(k1, k2)
}

func (b *BTree) Clear() {
	b.mutex.Lock()
	b.tree.Clear()
	b.mutex.Unlock()
}

func (b *BTree) Iterator(opt RangeOption) (Iterator, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return newBTreeIterator(b.tree, opt)
}

func (b *BTree) Get(key Key) (Hint, bool) {
	if key == nil {
		return Hint{}, false
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.tree.Get(key)
}

func (b *BTree) Put(hs ...Hint) error {
	if len(hs) == 0 {
		return nil
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if err := b.tree.Put(hs...); err != nil {
		return err
	}
	return nil
}

func (b *BTree) Del(ks ...Key) error {
	if len(ks) == 0 {
		return nil
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if err := b.tree.Del(ks...); err != nil {
		return err
	}

	return nil
}

func (b *BTree) Size() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.tree.Size()
}

func (b *BTree) Close() error {
	b.mutex.Lock()
	b.tree.Close()
	b.mutex.Unlock()
	return nil
}
