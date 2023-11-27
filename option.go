package riverdb

import (
	"fmt"
	"github.com/246859/river/file"
	"github.com/246859/river/index"
	"github.com/246859/river/wal"
	"path/filepath"
)

const (
	dataName     = "data"
	mergeName    = "merge"
	hintName     = "hint"
	finishedName = "finished"
	lockName     = "river"
)

const defaultMaxFileSize = file.MB * 256

const blockSize = wal.MaxBlockSize

var DefaultOptions = Options{
	MaxSize:        defaultMaxFileSize,
	BlockCache:     defaultMaxFileSize / file.MB,
	Fsync:          false,
	FsyncThreshold: blockSize * (defaultMaxFileSize / file.MB),
	Compare:        index.DefaultCompare,
	WatchSize:      100,
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
	// kv put/get events size for watch queue, disabled if is 0
	WatchSize int
	// decide how to sort keys
	Compare index.Compare
	// manually gc after closed db
	ClosedGc bool

	dataDir  string
	mergeDir string
	filelock string
}

func revise(opt Options) (Options, error) {
	if opt.Dir == "" {
		return opt, fmt.Errorf("db data dir must be specified")
	}

	if opt.Compare == nil {
		return opt, fmt.Errorf("key comparator must be specified")
	}

	if opt.MaxSize <= 0 {
		return opt, fmt.Errorf("invalid max file size: %d", opt.MaxSize)
	}

	if (int64(opt.BlockCache) * wal.MaxBlockSize) >= opt.MaxSize {
		return opt, fmt.Errorf("block cache size should be less then max file size")
	}

	if opt.FsyncThreshold >= opt.MaxSize {
		return opt, fmt.Errorf("sync threshold should be less than max file size")
	}

	opt.dataDir = filepath.Join(opt.Dir, dataName)
	opt.mergeDir = filepath.Join(opt.Dir, mergeName)
	opt.filelock = filepath.Join(opt.Dir, lockName)

	return opt, nil
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
