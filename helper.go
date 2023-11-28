package riverdb

import (
	"github.com/246859/river/entry"
	"unsafe"
)

//go:noescape
//go:linkname memhash runtime.memhash
func memhash(p unsafe.Pointer, h, s uintptr) uintptr

// FastRand is a fast thread local random function.
//
//go:linkname FastRand runtime.fastrand
func FastRand() uint32

// memHash is the hash function used by go map, it utilizes available hardware instructions(behaves
// as aeshash if aes instruction is available).
// NOTE: The hash seed changes for every process. So, this cannot be used as a persistent hash.
func memHash(data []byte) uint64 {
	h := FastRand()
	ptr := unsafe.Pointer(unsafe.SliceData(data))
	return uint64(memhash(ptr, uintptr(h), uintptr(len(data))))
}

func isExpiredOrDeleted(en entry.Entry) bool {
	return entry.IsExpired(en.TTL) || en.Type != entry.DataEntryType
}

// Mask 32-bit mask is used to store different status information
type Mask uint32

func (m *Mask) Store(ms ...uint32) {
	for _, mask := range ms {
		*m |= Mask(mask)
	}
}

func (m *Mask) Remove(ms ...uint32) {
	for _, mask := range ms {
		*m ^= Mask(mask)
	}
}

func (m *Mask) CheckAny(ms ...uint32) bool {
	for _, mask := range ms {
		if *m&Mask(mask) != 0 {
			return true
		}
	}
	return false
}

func (m *Mask) CheckSome(ms ...uint32) bool {
	for _, mask := range ms {
		if *m&Mask(mask) == 0 {
			return false
		}
	}
	return true
}
