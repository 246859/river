package entry

import (
	"encoding/binary"
)

type HeaderMarshaler interface {
	MarshalHeader(entry Header) ([]byte, int, error)
}

type HeaderUnmarshaler interface {
	UnMarshalHeader(raws []byte) (Header, int, error)
}

// MaxHeaderSize
// |                     31 B                   |
// +------+------+----------+--------+----------+---------+---------+
// | type | ttl  | batch    | key_sz | value_sz | key     | value   |
// +------+------+----------+--------+----------+---------+---------+
// | 1 B  | 10 B | 10 B     | 5 B    | 5 B      | unfixed | unfixed |
// +------+------+----------+--------+----------+---------+---------+
const MaxHeaderSize = 1 + binary.MaxVarintLen64*2 + binary.MaxVarintLen32*2

// Header represents a header of data entry
type Header struct {
	Type  EType
	TTL   int64
	Batch int64
	Ksz   uint32
	Vsz   uint32
}

func (d defaultEntryMarshaler) MarshalHeader(header Header) ([]byte, int, error) {
	headerBytes := make([]byte, MaxHeaderSize)
	offset := 0

	if err := CheckEntryType(header.Type); err != nil {
		return []byte{}, 0, err
	}

	// header type
	headerBytes[offset] = header.Type
	offset += 1

	// ttl
	offset += binary.PutVarint(headerBytes[offset:], header.TTL)
	// batch id
	offset += binary.PutVarint(headerBytes[offset:], header.Batch)
	// key size
	offset += binary.PutVarint(headerBytes[offset:], int64(header.Ksz))
	// value size
	offset += binary.PutVarint(headerBytes[offset:], int64(header.Vsz))

	return headerBytes, offset, nil
}

func (d defaultEntryMarshaler) UnMarshalHeader(raws []byte) (Header, int, error) {
	var (
		offset = 0
		header Header
	)

	if raws == nil {
		return header, 0, ErrNilRawData
	}

	// Type
	et := raws[offset]
	offset += 1

	if err := CheckEntryType(et); err != nil {
		return header, 0, err
	}

	// ttl
	ttl, ttlOff := binary.Varint(raws[offset:])
	offset += ttlOff

	batchId, bacthOff := binary.Varint(raws[offset:])
	offset += bacthOff

	ksz, koff := binary.Varint(raws[offset:])
	offset += koff

	vsz, voff := binary.Varint(raws[offset:])
	offset += voff

	header = Header{
		TTL:   ttl,
		Batch: batchId,
		Type:  et,
		Ksz:   uint32(ksz),
		Vsz:   uint32(vsz),
	}

	return header, offset, nil
}
