package riverdb

import (
	"fmt"
	"github.com/246859/river/file"
	"github.com/246859/river/index"
	"github.com/246859/river/wal"
)

const (
	dataName = "data"
	hintName = "hint"
	lockName = "river"
)

const defaultMaxFileSize = file.MB * 256

const blockSize = wal.MaxBlockSize

var DefaultOptions = Options{
	MaxSize:        defaultMaxFileSize,
	BlockCache:     defaultMaxFileSize / file.MB,
	Fsync:          false,
	FsyncThreshold: blockSize * (defaultMaxFileSize / file.MB),
	Compare:        index.DefaultCompare,
}

// Option applying changes to the given option
type Option func(option *Options)

// Options represent db configuration
type Options struct {
	// data dir that stores data files
	Dir string
	// max bytes size of the single data file can hold
	MaxSize int64
	// wal block cache size
	BlockCache uint32
	// call sync per write
	Fsync bool
	// call sync when reach the threshold
	FsyncThreshold int64
	// decide how to sort keys
	Compare index.Compare
	// manually gc after closed db
	ClosedGc bool
}

func (opt Options) check() error {
	if opt.Dir == "" {
		return fmt.Errorf("db data dir must be specified")
	}

	if opt.Compare == nil {
		return fmt.Errorf("key comparator must be specified")
	}

	if opt.MaxSize <= 0 {
		return fmt.Errorf("invalid max file size: %d", opt.MaxSize)
	}

	if (int64(opt.BlockCache) * wal.MaxBlockSize) >= opt.MaxSize {
		return fmt.Errorf("block cache size should be less then max file size")
	}

	if opt.FsyncThreshold >= opt.MaxSize {
		return fmt.Errorf("sync threshold should be less than max file size")
	}

	return nil
}

func WithDir(dir string) Option {
	return func(option *Options) {
		option.Dir = dir
	}
}

func WithMaxSize(size int64) Option {
	return func(option *Options) {
		option.MaxSize = size
	}
}

func WithBlockCache(block uint32) Option {
	return func(option *Options) {
		option.BlockCache = block
	}
}

func WithFsync(sync bool) Option {
	return func(option *Options) {
		option.Fsync = sync
	}
}

func WithFsyncThreshold(threshold int64) Option {
	return func(option *Options) {
		option.FsyncThreshold = threshold
	}
}

func WithCompare(compare index.Compare) Option {
	return func(option *Options) {
		option.Compare = compare
	}
}

func WithClosedGc(gc bool) Option {
	return func(option *Options) {
		option.ClosedGc = gc
	}
}
