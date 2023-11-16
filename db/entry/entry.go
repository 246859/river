package entry

import (
	"github.com/pkg/errors"
)

var (
	ErrNilKey           = errors.New("key is nil")
	ErrNilRawData       = errors.New("raw data bytes is nil")
	ErrInvalidEntryType = errors.New("entry type is invalid")
)

type EType = byte

const (
	DataEntryType EType = 1 + iota

	DeletedEntryType

	BatchFinishedEntryType
)

func CheckEntryType(t EType) error {
	if t < DataEntryType || t > BatchFinishedEntryType {
		return errors.Wrapf(ErrInvalidEntryType, "%d", t)
	}
	return nil
}

// Entry represents a data entry in data file
type Entry struct {
	Type  EType
	Key   []byte
	Value []byte

	TTL   int64
	Batch int64
}

type Marshaler interface {
	MarshalEntry(entry Entry) ([]byte, error)
}

type UnMarshaler interface {
	UnMarshalEntry(raws []byte) (Entry, error)
}

type defaultEntryMarshaler struct {
}

func (d defaultEntryMarshaler) MarshalEntry(entry Entry) ([]byte, error) {
	// check key length
	if entry.Key == nil {
		return []byte{}, ErrNilKey
	}

	headerBytes, offset, err := d.MarshalHeader(Header{
		Type:  entry.Type,
		Ksz:   uint32(len(entry.Key)),
		Vsz:   uint32(len(entry.Value)),
		TTL:   entry.TTL,
		Batch: entry.Batch,
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

func (d defaultEntryMarshaler) UnMarshalEntry(bytes []byte) (Entry, error) {
	var entry Entry
	header, offset, err := d.UnMarshalHeader(bytes)
	if err != nil {
		return entry, err
	}

	entry.Type = header.Type
	entry.TTL = header.TTL
	entry.Key = make([]byte, header.Ksz)
	entry.Value = make([]byte, header.Vsz)

	copy(entry.Key, bytes[offset:])
	copy(entry.Value, bytes[offset+len(entry.Key):])

	return entry, nil
}

var defaultMarshaler = defaultEntryMarshaler{}

func MarshalEntry(entry Entry) ([]byte, error) {
	return defaultMarshaler.MarshalEntry(entry)
}

func UnMarshalEntry(rawdata []byte) (Entry, error) {
	return defaultMarshaler.UnMarshalEntry(rawdata)
}

func MarshalHeader(header Header) ([]byte, int, error) {
	return defaultMarshaler.MarshalHeader(header)
}

func UnMarshalHeader(rawdata []byte) (Header, int, error) {
	return defaultMarshaler.UnMarshalHeader(rawdata)
}

func MarshalHint(hint Hint) ([]byte, error) {
	return defaultMarshaler.MarshalHint(hint)
}

func UnMarshalHint(rawdata []byte) (Hint, error) {
	return defaultMarshaler.UnMarshalHint(rawdata)
}
