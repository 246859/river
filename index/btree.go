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

func BtreeIndex(degree int, less LessKey) *BTree {
	bi := new(BTree)
	bi.less = less
	bi.tree = btree.NewG[Hint](degree, func(a, b Hint) bool {
		return less(a.Key, b.Key)
	})
	return bi
}

// BTree is btree implementation of Index that
// allows the actions of finding data, sequential access, inserting data, and deleting to be done in O(log n) time
type BTree struct {
	tree  *btree.BTreeG[Hint]
	mutex sync.RWMutex

	less LessKey
}

func (b *BTree) Iterator(opt RangeOption) (Iterator, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return newBTreeIterator(b, opt)
}

func (b *BTree) Get(key Key) (Hint, bool) {
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

	hint := Hint{Key: key}
	_, exist := b.tree.Delete(hint)

	return exist, nil
}

func (b *BTree) Size() int {
	return b.tree.Len()
}

func (b *BTree) Close() error {
	return nil
}

func newBTreeIterator(btr *BTree, opt RangeOption) (Iterator, error) {

	var (
		minHint = Hint{Key: opt.Min}
		maxHint = Hint{Key: opt.Max}
		// temporary string, never to modify it
		pattern  = str.BytesToString(opt.Pattern)
		iterator Iterator
	)

	if minHint.Key != nil && maxHint.Key != nil && !btr.less(minHint.Key, maxHint.Key) {
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

	hints := make([]Hint, 0, 200)

	searchFn := func(h Hint) bool {
		candidate := true
		if reg != nil {
			candidate = reg.Match(h.Key)
		}
		if candidate {
			hints = append(hints, h)
		}
		return true
	}

	// all keys
	if minHint.Key == nil && maxHint.Key == nil {
		btr.tree.Ascend(searchFn)
		// less than or equal max
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
		hints:  hints,
		cursor: 0,
	}

	return iterator, nil
}

type BTreeIterator struct {
	hints  []Hint
	cursor int
}

func (bi *BTreeIterator) Next() (Hint, bool) {
	var (
		hint Hint
		out  bool
	)

	if bi.cursor < len(bi.hints) {
		hint = bi.hints[bi.cursor]
		bi.cursor++
	}

	if bi.cursor >= len(bi.hints) {
		out = true
	}

	return hint, out
}

func (bi *BTreeIterator) Rewind() {
	bi.cursor = 0
}

func (bi *BTreeIterator) Close() error {
	return nil
}