package wal

import (
	"encoding/binary"
	"github.com/246859/river/file"
	"github.com/246859/river/pkg/crc"
	"github.com/pkg/errors"
	"github.com/valyala/bytebufferpool"
	"io"
	"os"
	"sync"
)

var (
	ErrAlreadyClosed    = errors.New("file already closed")
	ErrBlockExceeded    = errors.New("block size exceeded")
	ErrWriteToImmutable = errors.New("write to immutable")
)

type ChunkType = byte

const (
	ChunkFullType ChunkType = 1 + iota
	ChunkFirstType
	ChunkMiddleType
	ChunkLastType
)

// ChunkHeaderSize
// +-----------+--------------+------------+------------+
// | check sum | chunk length | chunk type | chunk data |
// +-----------+--------------+------------+------------+
// | 4 Bytes   | 2 Bytes      | 1 Bytes    | <= 32KB-7B |
// +-----------+--------------+------------+------------+
const ChunkHeaderSize = 7

const MaxBlockSize = 32 * file.KB

type ChunkHeader struct {
	crc       uint32
	Length    uint16
	ChunkType ChunkType
}

type ChunkPos struct {
	// file id
	Fid uint32
	// chunk in which block
	Block uint32
	// chunk offset in block
	Offset int64
	// chunk size
	Size int64
}

// File represents a wal file
type File interface {
	// Fid return file id
	Fid() uint32
	// Read bytes with specified block and offset
	// chunkOffset is offset by the start of the specified block
	Read(block uint32, chunkOffset int64) ([]byte, error)
	// Write bytes and returns ChunkPos
	Write(data []byte) (ChunkPos, error)
	// WriteAll write bytes in batch and returns ChunkPos
	WriteAll(data [][]byte) ([]ChunkPos, error)
	// Iterator iterate per chunk in file
	Iterator() (ChunkIterator, error)
	// Size return size of file
	Size() int64
	// Remove remove file
	Remove() error
	// Sync buffer to disk
	Sync() error
	// MarkImmutable mark this file as immutable
	MarkImmutable()
	// Close file and release resource
	Close() error
}

// ChunkIterator iterate chunk data for a log file
type ChunkIterator interface {
	// File returns file that iterator belongs to
	File() File
	// Index returns current chunk position
	Index() ChunkPos
	// Next returns the next chunk data bytes, chunk info, and error
	Next() ([]byte, ChunkPos, error)
}

func newBlock() any {
	return make([]byte, MaxBlockSize)
}

func newChunkHeader() any {
	return make([]byte, ChunkHeaderSize)
}

// open a new log file
func openLogFile(dir, ext string, fid uint32, cache *lruCache, bufferPool *bytebufferpool.Pool) (*LogFile, error) {
	fd, err := file.OpenStdFile(WalFileName(dir, fid, ext), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	offset, err := fd.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	return &LogFile{
		fid:         fid,
		fd:          fd,
		block:       uint32(offset / MaxBlockSize),
		blockOffset: offset % MaxBlockSize,
		crc:         crc.Koopman32,
		blockPool:   sync.Pool{New: newBlock},
		headerPool:  sync.Pool{New: newChunkHeader},
		blockCache:  cache,
		bufferPool:  bufferPool,
	}, nil
}

type LogFile struct {
	fid uint32
	fd  file.IO

	crc crc.Crc32

	// which block is currently writing
	block uint32
	// size of current writing block
	blockOffset int64

	blockPool  sync.Pool
	headerPool sync.Pool

	blockCache *lruCache
	bufferPool *bytebufferpool.Pool

	closed    bool
	immutable bool
}

func (log *LogFile) Fid() uint32 {
	return log.fid
}

func (log *LogFile) Read(block uint32, chunkOffset int64) ([]byte, error) {
	readBytes, _, err := log.read(block, chunkOffset)
	return readBytes, err
}

func (log *LogFile) read(block uint32, chunkOffset int64) ([]byte, ChunkPos, error) {
	if log.closed {
		return nil, ChunkPos{}, ErrAlreadyClosed
	}

	var (
		fileSize     = log.Size()
		nextChunkPos = ChunkPos{Fid: log.fid}

		dataBytes []byte
		tmpBlock  = log.blockPool.Get().([]byte)
		tmpHeader = log.headerPool.Get().([]byte)
	)

	defer func() {
		log.blockPool.Put(tmpBlock)
		log.headerPool.Put(tmpHeader)
	}()

	// read a block size by per-circulate
	for {

		readSize := int64(MaxBlockSize)
		blockOffset := int64(block * MaxBlockSize)

		if blockOffset+readSize > fileSize {
			readSize = fileSize - blockOffset
		}

		if chunkOffset >= readSize || readSize < ChunkHeaderSize {
			return dataBytes, nextChunkPos, io.EOF
		}

		var (
			cacheKey    = blockCacheKey(log.fid, block)
			hitCache    bool
			cachedBlock []byte
		)

		// check if is cached
		if log.blockCache != nil {
			cachedBlock, hitCache = log.blockCache.Get(cacheKey)
		}

		if hitCache {
			copy(tmpBlock, cachedBlock)
		} else {
			_, err := log.fd.ReadAt(tmpBlock[:readSize], blockOffset)
			if err != nil {
				return dataBytes, nextChunkPos, err
			}

			// cached the block
			if log.blockCache != nil && readSize == MaxBlockSize {
				cachedBlock = make([]byte, MaxBlockSize)
				copy(cachedBlock, tmpBlock)
				log.blockCache.Add(cacheKey, cachedBlock)
			}
		}

		// read header info
		copy(tmpHeader, tmpBlock[chunkOffset:chunkOffset+ChunkHeaderSize])
		headerCrc := binary.LittleEndian.Uint32(tmpHeader[:4])
		payloadLen := binary.LittleEndian.Uint16(tmpHeader[4:6])
		chunkType := tmpHeader[6]
		payloadOffset := chunkOffset + ChunkHeaderSize
		payloadTail := payloadOffset + int64(payloadLen)

		// read chunk payload and write to data bytes
		dataBytes = append(dataBytes, tmpBlock[payloadOffset:payloadTail]...)

		// check crc
		check := log.crc.Sum(tmpBlock[chunkOffset+4 : payloadTail])
		if check != headerCrc {
			return dataBytes, nextChunkPos, crc.ErrCrcCheckFailed
		}

		// chunk read finished
		if chunkType == ChunkFullType || chunkType == ChunkLastType {
			nextChunkPos.Block = block
			nextChunkPos.Offset = payloadTail

			// if current chunk is the last chunk of block that no left space to hold another one
			// then turn to the next block
			if payloadTail+ChunkHeaderSize > MaxBlockSize {
				nextChunkPos.Block++
				nextChunkPos.Offset = 0
			}
			break
		}

		// update condition
		block++
		chunkOffset = 0
	}

	return dataBytes, nextChunkPos, nil
}

func (log *LogFile) Write(data []byte) (ChunkPos, error) {
	var (
		pos ChunkPos
	)
	if log.closed {
		return pos, ErrAlreadyClosed
	}

	batch, err := log.WriteAll([][]byte{data})
	if err != nil {
		return pos, err
	}

	if len(batch) == 1 {
		pos = batch[0]
	}
	return pos, nil
}

func (log *LogFile) WriteAll(data [][]byte) ([]ChunkPos, error) {
	var (
		oldBlock     = log.block
		oldBlockSize = log.blockOffset
		poss         = make([]ChunkPos, len(data))
		err          error
	)

	// write to immutable
	if log.immutable {
		return poss, ErrWriteToImmutable
	}

	// load bytes buffer
	buffer := log.bufferPool.Get()
	buffer.Reset()

	defer func() {
		if err != nil {
			log.block = oldBlock
			log.blockOffset = oldBlockSize
		}
		log.bufferPool.Put(buffer)
	}()

	var pos ChunkPos
	for i := 0; i < len(data); i++ {
		pos, err = log.writeToBuffer(data[i], buffer)
		if err != nil {
			return poss, err
		}
		poss[i] = pos
	}

	err = log.writeToFile(buffer)
	if err != nil {
		return poss, err
	}

	return poss, nil
}

// write chunk data to byte buffer
func (log *LogFile) writeToBuffer(data []byte, buffer *bytebufferpool.ByteBuffer) (ChunkPos, error) {
	var preparedPos ChunkPos

	if log.closed {
		return preparedPos, ErrAlreadyClosed
	}

	oldBufferOffset := buffer.Len()
	padding := int64(0)

	// if current block has no space left to accommodate more than bytes size of ChunkHeaderSize
	// then turn to a new chunk, and fill the rest bytes of the old chunk
	if log.blockOffset < MaxBlockSize && log.blockOffset+ChunkHeaderSize >= MaxBlockSize {
		padding += MaxBlockSize - log.blockOffset
		p := make([]byte, padding)
		buffer.Write(p)

		log.block++
		log.blockOffset = 0
	}

	// position info will be returned
	preparedPos = ChunkPos{
		Fid:    log.fid,
		Block:  log.block,
		Offset: log.blockOffset,
	}

	dataSize := int64(len(data))
	chunkSize := ChunkHeaderSize + dataSize

	// if current block can hold entire chunk which situation is ChunkFull
	// else data should be stored in multiple blocks
	if log.blockOffset+chunkSize <= MaxBlockSize {
		log.appendChunkToBuffer(buffer, data, ChunkFullType)
		preparedPos.Size += chunkSize
	} else {
		var (
			leftDataSize = dataSize
			chunkCount   = int64(0)
			curBlockSz   = log.blockOffset
		)

		// Divide a data into multiple chunks and write them to the buffer in batches.
		// The first chunk written is ChunkFirstType, the middle is ChunkMiddleType, and the last is ChunkLastType.
		// Considering that the maximum limit for a block is 32KB,
		// there may be situations where a data is divided into multiple chunks and stored in different blocks.
		for leftDataSize > 0 {
			var (
				// left space of current block can hold payload
				leftBlockSz = MaxBlockSize - curBlockSz - ChunkHeaderSize
				// how many bytes already written to buffer
				bufferOffset = dataSize - leftDataSize
			)

			// next chunk size
			nextChunkSize := leftBlockSz
			if nextChunkSize > leftDataSize {
				nextChunkSize = leftDataSize
			}

			// next buffer offset
			nextBufferOffset := bufferOffset + nextChunkSize
			if nextBufferOffset > dataSize {
				nextBufferOffset = dataSize
			}

			var chunkType ChunkType
			switch leftDataSize {
			case dataSize:
				chunkType = ChunkFirstType
			case nextChunkSize:
				chunkType = ChunkLastType
			default:
				chunkType = ChunkMiddleType
			}

			// write a chunk to buffer
			log.appendChunkToBuffer(buffer, data[bufferOffset:nextBufferOffset], chunkType)

			// update range info
			leftDataSize -= nextChunkSize
			chunkCount += 1
			curBlockSz = (curBlockSz + ChunkHeaderSize + nextChunkSize) % MaxBlockSize
		}
		preparedPos.Size = chunkCount*ChunkHeaderSize + dataSize
	}

	// check if chunk size is valid
	newBufferOffset := buffer.Len()
	if padding+preparedPos.Size != int64(newBufferOffset-oldBufferOffset) {
		return preparedPos, errors.Errorf("write invalid chunk size, expected %d, actual %d",
			newBufferOffset-oldBufferOffset, padding+preparedPos.Size)
	}

	// update logfile status
	log.blockOffset += preparedPos.Size
	if log.blockOffset >= MaxBlockSize {
		log.block += uint32(log.blockOffset / MaxBlockSize)
		log.blockOffset %= MaxBlockSize
	}

	return preparedPos, nil
}

func (log *LogFile) appendChunkToBuffer(buffer *bytebufferpool.ByteBuffer, data []byte, chunkType ChunkType) {
	header := log.headerPool.Get().([]byte)
	defer log.headerPool.Put(header)

	crcOffset := 4

	// size of chunk payload
	binary.LittleEndian.PutUint16(header[crcOffset:6], uint16(len(data)))
	// chunk type
	header[6] = chunkType

	// check head and data crc sum
	sum := log.crc.Sum(header[crcOffset:])
	sum = log.crc.Update(sum, data)
	binary.LittleEndian.PutUint32(header[:crcOffset], sum)

	// write to buffer
	buffer.Write(header)
	buffer.Write(data)
}

func (log *LogFile) writeToFile(buffer *bytebufferpool.ByteBuffer) error {
	if log.blockOffset > MaxBlockSize {
		return errors.WithMessagef(ErrBlockExceeded, "%d log file %d block szize exceeded: %d, max: %d",
			log.fid, log.block, log.blockOffset, MaxBlockSize)
	}
	_, err := log.fd.Write(buffer.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (log *LogFile) Iterator() (ChunkIterator, error) {
	if log.closed {
		return nil, ErrAlreadyClosed
	}
	return &LogFileChunkIterator{f: log}, nil
}

func (log *LogFile) Remove() error {
	err := log.Close()
	if err != nil && !errors.Is(err, ErrAlreadyClosed) {
		return err
	}
	return os.Remove(log.fd.Name())
}

func (log *LogFile) Size() int64 {
	return int64(log.block*MaxBlockSize) + log.blockOffset
}

func (log *LogFile) Sync() error {
	if log.closed {
		return ErrAlreadyClosed
	}
	return log.fd.Sync()
}

func (log *LogFile) MarkImmutable() {
	log.immutable = true
}

func (log *LogFile) Close() error {
	if log.closed {
		return ErrAlreadyClosed
	}
	log.closed = true
	return log.fd.Close()
}

type LogFileChunkIterator struct {
	f           *LogFile
	block       uint32
	chunkOffset int64
}

func (l *LogFileChunkIterator) Index() ChunkPos {
	return ChunkPos{
		Fid:    l.f.Fid(),
		Block:  l.block,
		Offset: l.chunkOffset,
	}
}

func (l *LogFileChunkIterator) File() File {
	return l.f
}

func (l *LogFileChunkIterator) Next() ([]byte, ChunkPos, error) {
	if l.f.closed {
		return nil, ChunkPos{}, ErrAlreadyClosed
	}

	chunkPos := ChunkPos{
		Fid:    l.f.fid,
		Block:  l.block,
		Offset: l.chunkOffset,
	}

	dataBytes, nextChunkPos, err := l.f.read(chunkPos.Block, chunkPos.Offset)
	if err != nil {
		return nil, ChunkPos{}, err
	}

	// compute the chunk size
	chunkAt := int64(nextChunkPos.Block)*MaxBlockSize + nextChunkPos.Offset
	nextChunkAt := int64(chunkPos.Block)*MaxBlockSize + chunkPos.Offset
	chunkPos.Size = nextChunkAt - chunkAt

	// update chunk index
	l.block = nextChunkPos.Block
	l.chunkOffset = nextChunkPos.Offset

	return dataBytes, chunkPos, nil
}
