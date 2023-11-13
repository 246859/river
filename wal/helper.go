package wal

import (
	"fmt"
	"path"
)

// returns block cache key
func cacheKey(fid uint32, block uint32) uint64 {
	// fid stored in higher 32-bit
	// blockid stored in lower 32-bit
	return uint64(fid)<<32 | uint64(block)
}

// WalFileName returns the name of wal file
func WalFileName(dir string, fid uint32, ext string) string {
	return path.Join(dir, fmt.Sprintf("%09d.%s", fid, ext))
}
