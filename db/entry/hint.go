package entry

import "encoding/binary"

type HintMarshaler interface {
	MarshalHint(hint Hint) ([]byte, error)
}

type HintUnMarshaler interface {
	UnMarshalHint(rawdata []byte) (Hint, error)
}

// +-----+-------+--------+------+
// | fid | block | offset | ttl  |
// +-----+-------+--------+------+
// | 5 B | 5 B   | 10 B   | 10 B |
// +-----+-------+--------+------+
const maxHintSize = binary.MaxVarintLen32*2 + binary.MaxVarintLen64*2

// Hint represents the position of record in which file specified by fid
type Hint struct {
	// represents entry storage in which file
	Fid uint32
	// block
	Block uint32
	// chunk offset
	Offset int64
	// ttl
	TTL int64
}

func (d defaultEntryMarshaler) MarshalHint(hint Hint) ([]byte, error) {
	var (
		hintBytes = make([]byte, maxHintSize)
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

	return hintBytes, nil
}

func (d defaultEntryMarshaler) UnMarshalHint(rawdata []byte) (Hint, error) {
	var (
		hint   Hint
		offset = 0
	)

	if rawdata == nil {
		return hint, ErrNilRawData
	}

	fid, idOff := binary.Varint(rawdata[offset:])
	offset += idOff

	block, bOff := binary.Varint(rawdata[offset:])
	offset += bOff

	chunkOffset, cOff := binary.Varint(rawdata[offset:])
	offset += cOff

	ttl, ttlOff := binary.Varint(rawdata[offset:])
	offset += ttlOff

	hint = Hint{
		Fid:    uint32(fid),
		Block:  uint32(block),
		Offset: chunkOffset,
		TTL:    ttl,
	}

	return hint, nil
}
