package wal

import (
	"fmt"
	"path"
)

// returns block cache key
func blockCacheKey(fid uint32, block uint32) uint64 {
	// fid stored in higher 32-bit
	// blockid stored in lower 32-bit
	return uint64(fid)<<32 | uint64(block)
}

// WalFileName returns the name of wal file
func WalFileName(dir string, fid uint32, ext string) string {
	return path.Join(dir, fmt.Sprintf("%09d.%s", fid, ext))
}

// estimate the possible max size of data will be written in wal file
// if the first written block has already contains chunks, maybe occurs the first data chunk could be hold by the block
// but no left space to hold the second chunk header, in that case, the block need padding to fill the left space.
// padding + block * header + data size
func estimateBlockSize(dataSize int64) int64 {
	block := dataSize/MaxBlockSize + 1
	return block*ChunkHeaderSize + dataSize + ChunkHeaderSize
}
