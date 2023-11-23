package entry

import (
	"github.com/pkg/errors"
	"time"
)

var (
	ErrNilKey             = errors.New("key is nil")
	ErrNilRawData         = errors.New("raw data bytes is nil")
	ErrInvalidEntryType   = errors.New("entry type is invalid")
	ErrRedundantEntryType = errors.New("redundant entry type")
	ErrExpired            = errors.New("entry is expired")
)

type EType = byte

const (
	DataEntryType EType = 1 + iota

	DeletedEntryType

	TxnCommitEntryType

	TxnFinishedEntryType
)

func CheckEntryType(t EType) error {
	if t < DataEntryType || t > TxnFinishedEntryType {
		return errors.Wrapf(ErrInvalidEntryType, "%d", t)
	}
	return nil
}

// Entry represents a data entry in data file
type Entry struct {
	Type  EType
	Key   []byte
	Value []byte

	TTL  int64
	TxId int64
}

// Header represents a header of data entry
type Header struct {
	Type  EType
	TTL   int64
	Batch int64
	Ksz   uint32
	Vsz   uint32
}
type Marshaler interface {
	MarshalEntry(entry Entry) ([]byte, error)
}

type UnMarshaler interface {
	UnMarshalEntry(raws []byte) (Entry, error)
}

type HeaderMarshaler interface {
	MarshalHeader(entry Header) ([]byte, int, error)
}

type HeaderUnmarshaler interface {
	UnMarshalHeader(raws []byte) (Header, int, error)
}

type Serializer interface {
	Marshaler
	UnMarshaler
	HeaderMarshaler
	HeaderUnmarshaler
}

var defaultMarshaler = BinaryEntry{}

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

// Validate validates the given entry if is a valid data entry
func Validate(entry Entry) error {
	// must be a data type
	if entry.Type != DataEntryType {
		return ErrRedundantEntryType
	}

	// key must not be nil
	if entry.Key == nil {
		return ErrNilKey
	}

	// must be not expired
	if IsExpired(entry.TTL) {
		return ErrExpired
	}

	return nil
}

func UnixMill() int64 {
	return time.Now().UnixMilli()
}

func NewTTL(ttl time.Duration) int64 {
	return UnixMill() + ttl.Milliseconds()
}

func LeftTTl(ttl int64) time.Duration {
	return time.Duration(ttl - UnixMill())
}

// IsExpired
// if ttl is 0, represents of persistent
// only if ttl > 0, entry has live time
func IsExpired(ttl int64) bool {
	return ttl > 0 && ttl <= UnixMill()
}
