package data

import (
	"encoding/binary"
	"time"
)

type HeaderMarshaler interface {
	MarshalHeader(entry EntryHeader) ([]byte, int, error)
}

type HeaderUnmarshaler interface {
	UnMarshalHeader(raws []byte) (EntryHeader, int, error)
}

// |                              25 bytes                           |
// +---------+------------+--------------+-------------+-------------+---------+---------+
// | crc     | entry_type | timestamp    | key_size    | value_size  | key     | value   |
// +---------+------------+--------------+-------------+-------------+---------+---------+
// | 4 bytes | 1 bytes    | max 10 bytes | max 5 bytes | max 5 bytes | unfixed | unfixed |
// +---------+------------+--------------+-------------+-------------+---------+---------+
//
// In theory, the maximum available space for value can reach 2^40 bytes = (2^12)*GB = 4096GB
// but not recommended put too large values that will reduce throughput
const maxHeaderSize = 4 + binary.MaxVarintLen64 + 1 + binary.MaxVarintLen32*2

// EntryHeader represents a header of data entry
type EntryHeader struct {
	crc uint32
	// unix milliseconds
	tstmap int64
	Et     EntryType
	Ksz    uint32
	Vsz    uint32
}

func (d defaultEntryMarshaler) MarshalHeader(header EntryHeader) ([]byte, int, error) {

	var (
		now = time.Now()
		// 32-bit crc sum
		offset = 4
		// key size per bytes
		ksz = header.Ksz
		// value size per bytes
		vsz = header.Vsz

		headerBytes = make([]byte, maxHeaderSize)
	)

	if err := CheckEntryType(header.Et); err != nil {
		return []byte{}, 0, err
	}

	// header type
	headerBytes[offset] = header.Et
	offset += 1

	// timestamp
	offset += binary.PutVarint(headerBytes[offset:], now.UnixMilli())
	// key size
	offset += binary.PutVarint(headerBytes[offset:], int64(ksz))
	// value size
	offset += binary.PutVarint(headerBytes[offset:], int64(vsz))

	return headerBytes, offset, nil
}

func (d defaultEntryMarshaler) UnMarshalHeader(raws []byte) (EntryHeader, int, error) {
	var (
		offset = 4
		header EntryHeader
	)

	if raws == nil {
		return header, 0, ErrNilRawData
	}

	// crc
	crc := binary.LittleEndian.Uint32(raws[:offset])

	// Et
	et := raws[offset]
	offset += 1

	if err := CheckEntryType(et); err != nil {
		return header, 0, err
	}

	timestamp, tstOff := binary.Varint(raws[offset:])
	offset += tstOff

	ksz, koff := binary.Varint(raws[offset:])
	offset += koff

	vsz, voff := binary.Varint(raws[offset:])
	offset += voff

	header = EntryHeader{
		crc:    crc,
		tstmap: timestamp,
		Et:     et,
		Ksz:    uint32(ksz),
		Vsz:    uint32(vsz),
	}

	return header, offset, nil
}
