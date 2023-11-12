package data

import (
	"encoding/binary"
	"github.com/246859/river/pkg/crc"
	"github.com/pkg/errors"
)

var (
	ErrNilKey           = errors.New("key is nil")
	ErrCrcCheckFailed   = errors.New("crc check failed")
	ErrNilRawData       = errors.New("raw data bytes is nil")
	ErrInvalidEntryType = errors.New("entry type is invalid")
)

type EntryType = byte

const (
	EntryDataType EntryType = 1 + iota

	DeleteEntryType
)

func CheckEntryType(t EntryType) error {
	if t < 1 || t > 2 {
		return errors.Wrapf(ErrInvalidEntryType, "%d", t)
	}
	return nil
}

// Entry represents a data entry in data file
type Entry struct {
	Key   []byte
	Value []byte
	Type  EntryType

	Timestamp int64
}

type Marshaler interface {
	MarshalEntry(entry Entry) ([]byte, error)
}

type UnMarshaler interface {
	UnMarshalEntry(raws []byte) (Entry, error)
}

type defaultEntryMarshaler struct {
	crc crc.Crc32
}

func (d defaultEntryMarshaler) MarshalEntry(entry Entry) ([]byte, error) {
	// check key length
	if entry.Key == nil {
		return []byte{}, ErrNilKey
	}

	headerBytes, offset, err := d.MarshalHeader(EntryHeader{
		Et:  entry.Type,
		Ksz: uint32(len(entry.Key)),
		Vsz: uint32(len(entry.Value)),
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

	// compute crc check sum
	crcSum := d.crc(entryBytes[4:])
	binary.LittleEndian.PutUint32(entryBytes[:4], crcSum)

	return entryBytes, nil
}

func (d defaultEntryMarshaler) UnMarshalEntry(bytes []byte) (Entry, error) {
	var entry Entry
	header, offset, err := d.UnMarshalHeader(bytes)
	if err != nil {
		return entry, err
	}

	entry.Type = header.Et
	entry.Timestamp = header.tstmap
	entry.Key = make([]byte, header.Ksz)
	entry.Value = make([]byte, header.Vsz)

	copy(entry.Key, bytes[offset:])
	copy(entry.Value, bytes[offset+len(entry.Key):])

	return entry, nil
}

var defaultMarshaler = defaultEntryMarshaler{crc: crc.KoopmanCrcSum}

func MarshalEntry(entry Entry) ([]byte, error) {
	return defaultMarshaler.MarshalEntry(entry)
}

func UnMarshalEntry(rawdata []byte) (Entry, error) {
	return defaultMarshaler.UnMarshalEntry(rawdata)
}

func MarshalHeader(header EntryHeader) ([]byte, int, error) {
	return defaultMarshaler.MarshalHeader(header)
}

func UnMarshalHeader(rawdata []byte) (EntryHeader, int, error) {
	return defaultMarshaler.UnMarshalHeader(rawdata)
}

func MarshalHint(hint EntryHint) ([]byte, error) {
	return defaultMarshaler.MarshalHint(hint)
}

func UnMarshalHint(rawdata []byte) (EntryHint, error) {
	return defaultMarshaler.UnMarshalHint(rawdata)
}
