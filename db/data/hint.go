package data

import "encoding/binary"

type HintMarshaler interface {
	MarshalHint(hint EntryHint) ([]byte, error)
}

type HintUnMarshaler interface {
	UnMarshalHint(rawdata []byte) (EntryHint, error)
}

// |                        35 bytes                          |
// +-------------+--------------+--------------+--------------+
// | fid         | offset       | size         | timestamp    |
// +-------------+--------------+--------------+--------------+
// | max 5 bytes | max 10 bytes | max 10 bytes | max 10 bytes |
// +-------------+--------------+--------------+--------------+
const maxHintSize = binary.MaxVarintLen32 + binary.MaxVarintLen64*3

// EntryHint represents the position of record in which file specified by fid
type EntryHint struct {
	// represents entry storage in which file
	Fid uint32
	// entry offset of file
	Offset uint64
	// entry size in file
	Size uint64
	// entry operation timestamp
	Tstamp int64
}

func (d defaultEntryMarshaler) MarshalHint(hint EntryHint) ([]byte, error) {
	var (
		hintBytes = make([]byte, maxHintSize)
		offset    = 0
	)

	// data file id
	offset += binary.PutUvarint(hintBytes[offset:], uint64(hint.Fid))

	// entry offset at file
	offset += binary.PutUvarint(hintBytes[offset:], hint.Offset)

	// bytes size of entry data
	offset += binary.PutUvarint(hintBytes[offset:], hint.Size)

	// update timestamp
	offset += binary.PutVarint(hintBytes[offset:], hint.Tstamp)

	return hintBytes, nil
}

func (d defaultEntryMarshaler) UnMarshalHint(rawdata []byte) (EntryHint, error) {
	var (
		hint   EntryHint
		offset = 0
	)

	if rawdata == nil {
		return hint, ErrNilRawData
	}

	fid, idOff := binary.Uvarint(rawdata[offset:])
	offset += idOff

	eoffset, oof := binary.Uvarint(rawdata[offset:])
	offset += oof

	size, soff := binary.Uvarint(rawdata[offset:])
	offset += soff

	tstamp, tof := binary.Varint(rawdata[offset:])
	offset += tof

	hint = EntryHint{
		Fid:    uint32(fid),
		Offset: eoffset,
		Size:   size,
		Tstamp: tstamp,
	}

	return hint, nil
}
