package entry

import "encoding/binary"

// BinaryEntry Binary data format
type BinaryEntry struct{}

// MaxHeaderSize
// |                     31 B                   |
// +------+------+----------+--------+----------+---------+---------+
// | type | ttl  | batch    | key_sz | value_sz | key     | value   |
// +------+------+----------+--------+----------+---------+---------+
// | 1 B  | 10 B | 10 B     | 5 B    | 5 B      | unfixed | unfixed |
// +------+------+----------+--------+----------+---------+---------+
const MaxHeaderSize = 1 + binary.MaxVarintLen64*2 + binary.MaxVarintLen32*2

func (d BinaryEntry) MarshalEntry(entry Entry) ([]byte, error) {
	// check key length
	if entry.Key == nil {
		return []byte{}, ErrNilKey
	}

	headerBytes, offset, err := d.MarshalHeader(Header{
		Type: entry.Type,
		Ksz:  uint32(len(entry.Key)),
		Vsz:  uint32(len(entry.Value)),
		TTL:  entry.TTL,
		TxId: entry.TxId,
	})

	if err != nil {
		return headerBytes, err
	}

	if entry.Value == nil {
		entry.Value = []byte{}
	}

	var (
		ksz        = len(entry.Key)
		vsz        = len(entry.Value)
		esz        = offset + ksz + vsz
		entryBytes = make([]byte, esz)
	)

	// copy header
	copy(entryBytes[:offset], headerBytes)

	// copy key and value
	copy(entryBytes[offset:], entry.Key)
	copy(entryBytes[offset+ksz:], entry.Value)

	return entryBytes, nil
}

func (d BinaryEntry) UnMarshalEntry(bytes []byte) (Entry, error) {
	var entry Entry
	header, offset, err := d.UnMarshalHeader(bytes)
	if err != nil {
		return entry, err
	}

	entry.Type = header.Type
	entry.TTL = header.TTL
	entry.TxId = header.TxId
	entry.Key = make([]byte, header.Ksz)
	entry.Value = make([]byte, header.Vsz)

	copy(entry.Key, bytes[offset:])
	copy(entry.Value, bytes[offset+len(entry.Key):])

	return entry, nil
}

func (d BinaryEntry) MarshalHeader(header Header) ([]byte, int, error) {
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
	// txn id
	offset += binary.PutVarint(headerBytes[offset:], header.TxId)
	// key size
	offset += binary.PutVarint(headerBytes[offset:], int64(header.Ksz))
	// value size
	offset += binary.PutVarint(headerBytes[offset:], int64(header.Vsz))

	return headerBytes, offset, nil
}

func (d BinaryEntry) UnMarshalHeader(raws []byte) (Header, int, error) {
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

	// batch
	batchId, bacthOff := binary.Varint(raws[offset:])
	offset += bacthOff

	// key size
	ksz, koff := binary.Varint(raws[offset:])
	offset += koff

	// value size
	vsz, voff := binary.Varint(raws[offset:])
	offset += voff

	header = Header{
		TTL:  ttl,
		TxId: batchId,
		Type: et,
		Ksz:  uint32(ksz),
		Vsz:  uint32(vsz),
	}

	return header, offset, nil
}
