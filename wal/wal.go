package wal

import (
	"fmt"
	"github.com/246859/river/types"
	"github.com/google/btree"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pkg/errors"
	"github.com/valyala/bytebufferpool"
	"io"
	"os"
	"path"
	"sort"
	"sync"
)

var (
	ErrFileNotFound   = errors.New("wal file not found")
	ErrDataExceedFile = errors.New("data size has exceeded the max file limit")
)

// DefaultWalSuffix is the extension name of wal file, default is wal.
// eg: 001.wal
const DefaultWalSuffix = "wal"

type Option struct {
	// data dir
	DataDir string
	// max log file size
	MaxFileSize int64
	// log file extension
	Ext string
	// how many block will be cached, off if it is 0
	BlockCache uint32
	// sync buffer to disk per write, if not enabled data maybe not save when machine crashed
	FsyncPerWrite bool
	// specify the written threshold for triggering sync
	FsyncThreshold int64
}

func reviseOpt(opt Option) (Option, error) {
	if opt.Ext == "" {
		opt.Ext = DefaultWalSuffix
	}

	if opt.MaxFileSize < int64(opt.BlockCache*MaxBlockSize) {
		return opt, fmt.Errorf("max file size must be greater than block cache")
	}

	return opt, nil
}

func defaultOption() Option {
	return Option{
		DataDir:        path.Join(os.TempDir(), DefaultWalSuffix),
		MaxFileSize:    types.MB * 768,
		Ext:            DefaultWalSuffix,
		BlockCache:     20,
		FsyncPerWrite:  true,
		FsyncThreshold: MaxBlockSize * 20,
	}
}

// lru 2q cache
type lruCache = lru.TwoQueueCache[uint64, []byte]

type F struct {
	Fid  uint32
	File File
}

func (f F) Less(nf F) bool {
	return f.Fid < nf.Fid
}

func newFBtree(degree int) *btree.BTreeG[F] {
	return btree.NewG[F](degree, func(a, b F) bool {
		return a.Less(b)
	})
}

// load wal file from the given directory
// returns active wal file and immutable wal files
func loadDataDir(dir string, ext string, cache *lruCache, bufferPool *bytebufferpool.Pool) (File, []File, error) {
	// try to mkdir
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}

	// collect fids
	fids := make([]uint32, 0, 20)
	for _, entry := range entries {
		var fid uint32
		_, err := fmt.Sscanf(entry.Name(), "%d"+"."+ext, &fid)
		if err != nil {
			continue
		}
		fids = append(fids, fid)
	}

	sort.Slice(fids, func(i, j int) bool {
		return fids[i] < fids[j]
	})

	var activeFid uint32
	if len(fids) == 0 {
		activeFid = 1
	} else {
		// active file usually is the last one
		activeFid = fids[len(fids)-1]
		fids = fids[:len(fids)-1]
	}

	// open active file
	active, err := openLogFile(dir, ext, activeFid, cache, bufferPool)
	if err != nil {
		return nil, nil, err
	}

	// open immutable files
	immutables := make([]File, 0, len(fids))
	for _, fid := range fids {
		immutable, err := openLogFile(dir, ext, fid, cache, bufferPool)
		if err != nil {
			return nil, nil, err
		}
		immutable.MarkImmutable()
		immutables = append(immutables, immutable)
	}

	return active, immutables, nil
}

// Open returns a new wal instance with the give Option
func Open(option Option) (*Wal, error) {
	option, err := reviseOpt(option)
	if err != nil {
		return nil, err
	}

	wal := &Wal{
		option:         option,
		immutables:     newFBtree(32),
		pendingBytes:   make([][]byte, 0, 20),
		byteBufferPool: new(bytebufferpool.Pool),
	}

	// block cache
	if wal.option.BlockCache > 0 {
		lru2QCache, err := lru.New2Q[uint64, []byte](int(wal.option.BlockCache))
		if err != nil {
			return wal, err
		}
		wal.blockCache = lru2QCache
	}

	// load data dir
	active, immutables, err := loadDataDir(wal.option.DataDir, wal.option.Ext, wal.blockCache, wal.byteBufferPool)
	if err != nil {
		return wal, err
	}

	wal.active = active
	for _, immutable := range immutables {
		wal.immutables.ReplaceOrInsert(F{immutable.Fid(), immutable})
	}

	return wal, nil
}

// Wal instance represents a data directory and its sub-files that store write-ahead-log format data
type Wal struct {
	// active file, data will be written to this file
	// if size of active file has no left space to hold data, open a new active one and make old one immutable
	active File

	// immutable files, read only
	immutables *btree.BTreeG[F]

	// block cache
	blockCache *lruCache

	// byte buffer pool
	byteBufferPool *bytebufferpool.Pool

	// bytes have written after the last Sync
	bytesWritten int64

	// bytes have been written to pendingBytes
	pendingSize int64

	// pendingBytes holds the bytes pending to be written in batches
	pendingBytes [][]byte

	mutex        sync.Mutex
	pendingMutex sync.Mutex

	option Option
}

// Read reads data at the specified position
func (w *Wal) Read(pos ChunkPos) ([]byte, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	var f File
	if pos.Fid == w.active.Fid() {
		f = w.active
	} else {
		ff, e := w.immutables.Get(F{Fid: pos.Fid})
		if e {
			f = ff.File
		}
	}

	if f == nil {
		return nil, errors.WithMessagef(ErrFileNotFound, "%09d not found", pos.Fid)
	}

	return f.Read(pos.Block, pos.Offset)
}

// Write writes data bytes to the active file
func (w *Wal) Write(data []byte) (ChunkPos, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	estimateSize := estimateBlockSize(int64(len(data)))

	// data too large to hold in single wal file
	if estimateSize > w.option.MaxFileSize {
		return ChunkPos{}, ErrDataExceedFile
	}

	// active wal file has no left space to hold data, allocate a new file
	if w.active.Size()+estimateSize > w.option.MaxFileSize {
		if err := w.rotate(); err != nil {
			return ChunkPos{}, err
		}
	}

	pos, err := w.active.Write(data)
	if err != nil {
		return pos, err
	}

	// decide if sync data to disk
	if err := w.sync(pos.Size, false); err != nil {
		return pos, err
	}

	return pos, nil
}

// PendingWrite writes data to pendingBytes
func (w *Wal) PendingWrite(data []byte) {
	w.pendingMutex.Lock()
	defer w.pendingMutex.Unlock()

	size := estimateBlockSize(int64(len(data)))

	w.pendingSize += size
	w.pendingBytes = append(w.pendingBytes, data)
}

func (w *Wal) PendingClean() {
	w.pendingMutex.Lock()
	defer w.pendingMutex.Unlock()

	w.pendingSize = 0
	w.pendingBytes = w.pendingBytes[:0]
}

// PendingPush writes pendingBytes to active file
func (w *Wal) PendingPush() ([]ChunkPos, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if len(w.pendingBytes) == 0 {
		return []ChunkPos{}, nil
	}

	// data exceed
	if w.pendingSize > w.option.MaxFileSize {
		return nil, ErrDataExceedFile
	}

	// no left space enough
	if w.pendingSize+w.active.Size() > w.option.MaxFileSize {
		if err := w.rotate(); err != nil {
			return nil, err
		}
	}

	pos, err := w.active.WriteAll(w.pendingBytes)
	if err != nil {
		return pos, err
	}

	defer w.PendingClean()

	// decide if sync
	if err := w.sync(w.pendingSize, false); err != nil {
		return pos, err
	}

	return pos, nil
}

func (w *Wal) Rotate() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.rotate()
}

// rotate open a new file to replace current active file, old active file will be mark as immutable
func (w *Wal) rotate() error {
	// force sync
	if err := w.sync(0, true); err != nil {
		return err
	}

	// open new file
	rotateFile, err := openLogFile(w.option.DataDir, w.option.Ext, w.active.Fid()+1, w.blockCache, w.byteBufferPool)
	if err != nil {
		return err
	}

	w.active.MarkImmutable()
	w.immutables.ReplaceOrInsert(F{w.active.Fid(), w.active})
	w.active = rotateFile

	return nil
}

// Iterator returns an FileIterator
// maxFid determines the file range for the iterator
// pos determines the chunk position for the iterator
func (w *Wal) Iterator(minFid, maxFid uint32, pos ChunkPos) (FileIterator, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if maxFid <= 0 {
		maxFid = w.active.Fid()
	}

	var fs []File

	w.immutables.Ascend(func(f F) bool {
		if f.Fid > maxFid {
			return false
		}

		if f.Fid >= minFid && f.Fid <= maxFid {
			fs = append(fs, f.File)
		}
		return true
	})

	if w.active.Fid() >= minFid && w.active.Fid() <= maxFid {
		fs = append(fs, w.active)
	}

	return newFileIteratorWithPos(fs, pos)
}

func (w *Wal) sync(size int64, force bool) error {

	// force sync
	if force || w.option.FsyncPerWrite {
		return w.active.Sync()
	}

	// check if is reach threshold
	w.bytesWritten += size
	needSync := w.option.FsyncThreshold > 0 && w.bytesWritten >= w.option.FsyncThreshold
	if needSync {
		err := w.active.Sync()
		if err == nil {
			w.bytesWritten = 0
		}
		return err
	}
	return nil
}

// Sync call fsync on the active file
func (w *Wal) Sync() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.sync(0, true)
}

// Close will close all the immutable files
// sync the active file and close it
func (w *Wal) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// clean cache
	if w.blockCache != nil {
		w.blockCache.Purge()
	}

	var closeErr error
	// close immutable files
	w.immutables.Ascend(func(f F) bool {
		closeErr = f.File.Close()
		return closeErr == nil
	})
	if closeErr != nil {
		return closeErr
	}

	// sync active file
	if err := w.active.Sync(); err != nil {
		return err
	}

	// finally close active file
	if err := w.active.Close(); err != nil {
		return err
	}

	return nil
}

// Purge close all the file and remove them
func (w *Wal) Purge() error {
	err := w.Close()
	if err != nil {
		return err
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	var removeErr error
	// remove immutable files
	w.immutables.Ascend(func(f F) bool {
		removeErr = f.File.Remove()
		return removeErr == nil
	})
	if removeErr != nil {
		return removeErr
	}

	// remove active file
	err = w.active.Remove()
	if err != nil {
		return err
	}

	return nil
}

// IsEmpty check wal data if is empty
func (w *Wal) IsEmpty() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.active.Size() == 0 && w.immutables.Len() == 0
}

// ActiveFid returns current active file id
func (w *Wal) ActiveFid() uint32 {
	if w.active != nil {
		return w.active.Fid()
	}
	return 0
}

func newFileIterator(files []File) (FileIterator, error) {
	if len(files) == 0 {
		return FileIterator{}, nil
	}

	var its []ChunkIterator

	for _, f := range files {
		it, err := f.Iterator()
		if err != nil {
			return FileIterator{}, err
		}
		its = append(its, it)
	}

	// sort by fid
	sort.Slice(its, func(i, j int) bool {
		return its[i].File().Fid() < its[j].File().Fid()
	})

	return FileIterator{
		its:   its,
		index: 0,
	}, nil
}

func newFileIteratorWithPos(files []File, pos ChunkPos) (FileIterator, error) {
	it, err := newFileIterator(files)
	if err != nil {
		return it, err
	}

	for {
		if it.IndexFid() < pos.Fid {
			it.NextFile()
			continue
		}

		if it.IndexPos().Block >= pos.Block && it.IndexPos().Offset >= pos.Offset {
			break
		}

		_, _, err := it.NextData()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return it, err
		}
	}

	return it, nil
}

// FileIterator iterates over all the files and chunks
type FileIterator struct {
	its   []ChunkIterator
	index int
}

// IndexFid returns the current fid
func (f *FileIterator) IndexFid() uint32 {
	if f.index < len(f.its) {
		ff := f.its[f.index].File()
		if ff != nil {
			return ff.Fid()
		}
	}
	return 0
}

// IndexPos returns curren index position
func (f *FileIterator) IndexPos() ChunkPos {
	if f.index < len(f.its) {
		return f.its[f.index].Index()
	}
	return ChunkPos{}
}

func (f *FileIterator) NextFile() {
	f.index++
}

func (f *FileIterator) NextData() ([]byte, ChunkPos, error) {
	if f.index >= len(f.its) {
		return nil, ChunkPos{}, io.EOF
	}

	data, pos, err := f.its[f.index].Next()
	if errors.Is(err, io.EOF) {
		f.NextFile()
		return f.NextData()
	}

	return data, pos, err
}
