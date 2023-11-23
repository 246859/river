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
