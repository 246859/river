package index

import (
	"encoding/binary"
	"github.com/246859/river/wal"
)

// MaxHintHeaderSize
// +-----+-------+--------+------+------+
// | fid | block | offset | ttl  | key  |
// +-----+-------+--------+------+------+
// | 5 B | 5 B   | 10 B   | 10 B | auto |
// +-----+-------+--------+------+------+
const MaxHintHeaderSize = binary.MaxVarintLen32*2 + binary.MaxVarintLen64*2

// Hint represents a kev hint data
type Hint struct {
	wal.ChunkPos
	TTL int64
	Key Key

	// it will be ignored when marshalling and unmarshalling.
	// Meta only used in runtime, it will never be persisted to the database,
	// to save memory, do not stored big data in this field.
	Meta any
}

func MarshalHint(hint Hint) []byte {
	var (
		hintBytes = make([]byte, MaxHintHeaderSize)
		offset    = 0
	)

	// data file id
	offset += binary.PutVarint(hintBytes[offset:], int64(hint.Fid))

	// entry offset at file
	offset += binary.PutVarint(hintBytes[offset:], int64(hint.Block))

	// update timestamp
	offset += binary.PutVarint(hintBytes[offset:], hint.Offset)

	// update timestamp
	offset += binary.PutVarint(hintBytes[offset:], hint.TTL)

	// len of key is unfixed, so put it in the last
	res := make([]byte, offset+len(hint.Key))
	copy(res, hintBytes[:offset])
	copy(res[offset:], hint.Key)

	return res
}

func UnMarshalHint(rawdata []byte) Hint {
	var (
		hint   Hint
		offset = 0
	)

	// fid
	fid, idOff := binary.Varint(rawdata[offset:])
	offset += idOff

	// block
	block, bOff := binary.Varint(rawdata[offset:])
	offset += bOff

	// offset
	chunkOffset, cOff := binary.Varint(rawdata[offset:])
	offset += cOff

	// ttl
	ttl, ttlOff := binary.Varint(rawdata[offset:])
	offset += ttlOff

	keybytes := make([]byte, len(rawdata)-offset)
	copy(keybytes, rawdata[offset:])

	hint = Hint{
		ChunkPos: wal.ChunkPos{
			Fid:    uint32(fid),
			Block:  uint32(block),
			Offset: chunkOffset,
		},
		TTL: ttl,
		Key: keybytes,
	}

	return hint
}
