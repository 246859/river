package wal

import (
	"github.com/246859/river/pkg/file"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/valyala/bytebufferpool"
	"os"
	"sync"
)

// DefaultWalSuffix is the extension name of wal file, default is wal.
// eg: 001.wal
const DefaultWalSuffix = "wal"

type Option struct {
	// data dir
	Dir string
	// max log file size
	FileSize int64
	// log file extension
	Ext string
	// block lru cache size, off if it is 0
	BlockCache uint64
	// sync buffer to disk per write, if not enabled data will not save when machine crashed
	Fsync bool
	// specify how many bytes have been written before sync to the disk
	BytesSync uint64
}

func defaultOption() Option {
	return Option{
		Dir:        os.TempDir(),
		FileSize:   file.MB * 768,
		Ext:        DefaultWalSuffix,
		BlockCache: maxBlockSize * 20,
		Fsync:      true,
		BytesSync:  maxBlockSize * 20,
	}
}

// lru 2q cache
type lruCache = lru.TwoQueueCache[uint64, []byte]

type Wal struct {
	// files
	active     File
	immutables []File

	// cache
	blockCache     *lruCache
	byteBufferPool *bytebufferpool.Pool

	// status
	bytesWritten int64
	pendingBytes [][]byte
	pendingSize  int64

	mutex        sync.Mutex
	pendingMutex sync.Mutex

	option Option
}
