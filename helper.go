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

var h = FastRand()

// memHash is the hash function used by go map, it utilizes available hardware instructions(behaves
// as aeshash if aes instruction is available).
// NOTE: The hash seed changes for every process. So, this cannot be used as a persistent hash.
func memHash(data []byte) uint64 {
	ptr := unsafe.Pointer(unsafe.SliceData(data))
	return uint64(memhash(ptr, uintptr(h), uintptr(len(data))))
}

func isExpiredOrDeleted(en entry.Entry) bool {
	return entry.IsExpired(en.TTL) || en.Type != entry.DataEntryType
}

// BitFlag 64-bit mask is used to store different status information
type BitFlag uint64

func (bf *BitFlag) Store(flags ...uint64) {
	for _, flag := range flags {
		*bf |= BitFlag(flag)
	}
}

func (bf *BitFlag) Check(flags ...uint64) bool {
	var f uint64
	for _, flag := range flags {
		f |= flag
	}

	return *bf&BitFlag(f) != 0
}

func (bf *BitFlag) Revoke(flags ...uint64) {
	var f uint64
	for _, flag := range flags {
		f |= flag
	}
	*bf ^= BitFlag(f)
}
