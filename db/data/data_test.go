package data

import (
	"github.com/stretchr/testify/assert"
	"math"
	"strings"
	"testing"
	"time"
)

func TestEntry_Marshal_UnMarshal(t *testing.T) {
	// #1
	{
		datas := []Entry{
			{Key: []byte("hello world!"), Value: []byte("world!"), Type: EntryDataType},
			{Key: []byte(""), Value: []byte(""), Type: DeleteEntryType},
			{Key: []byte("hello world!ddddddddddddddddddddddddddddddddddddddddd"), Value: []byte("zcxzxcz"), Type: EntryDataType},
			{Key: []byte("bbc"), Value: []byte(strings.Repeat("a", math.MaxInt32>>5)), Type: EntryDataType},
		}

		for _, entry := range datas {
			entryData, err := MarshalEntry(Entry{
				Key:   entry.Key,
				Value: entry.Value,
				Type:  entry.Type,
			})
			assert.Nil(t, err)
			assert.NotNil(t, entryData)

			uentry, err := UnMarshalEntry(entryData)
			assert.Nil(t, err)
			assert.Equal(t, entry.Key, uentry.Key)
			assert.Equal(t, entry.Value, uentry.Value)
			assert.Equal(t, entry.Type, uentry.Type)
		}
	}

	// #2
	{
		entry, err := MarshalEntry(Entry{Key: nil})
		assert.Equal(t, ErrNilKey, err)
		assert.Len(t, entry, 0)
	}

	// #3
	{
		entry, err := UnMarshalEntry(nil)
		assert.Equal(t, ErrNilRawData, err)
		assert.Nil(t, entry.Key)
		assert.Nil(t, entry.Value)
	}

	// #4
	{
		entry, err := MarshalEntry(Entry{Key: []byte(""), Type: 0})
		assert.ErrorIs(t, err, ErrInvalidEntryType)
		assert.Len(t, entry, 0)
	}
}

func TestHeader_Marshal_UnMarshal(t *testing.T) {
	// #1
	{
		datas := []EntryHeader{
			{crc: 0, Et: EntryDataType, Ksz: 100, Vsz: 200},
			{crc: 10203123, Et: DeleteEntryType, Ksz: 100, Vsz: 200},
			{crc: 1234567, Et: EntryDataType, Ksz: 100, Vsz: 200},
			{crc: 9812312, Et: EntryDataType, Ksz: 100, Vsz: 200},
			{crc: 0, Et: DeleteEntryType, Ksz: 0, Vsz: 0},
		}

		for _, header := range datas {
			marshalHeader, offset, err := MarshalHeader(header)
			assert.Nil(t, err)
			assert.True(t, len(marshalHeader) > 0)
			assert.True(t, offset > 0)
			assert.True(t, len(marshalHeader) <= maxHeaderSize)

			unMarshalHeader, uoffset, err := UnMarshalHeader(marshalHeader)
			assert.Nil(t, err)
			assert.Equal(t, header.Ksz, unMarshalHeader.Ksz)
			assert.Equal(t, header.Vsz, unMarshalHeader.Vsz)
			assert.Equal(t, header.Et, unMarshalHeader.Et)
			assert.True(t, uoffset > 0)
		}
	}

	// #2
	{
		header, offset, err := MarshalHeader(EntryHeader{Et: 0})
		assert.ErrorIs(t, err, ErrInvalidEntryType)
		assert.Len(t, header, 0)
		assert.Equal(t, 0, offset)
	}
}

func TestHint_Marshal_UnMarshal(t *testing.T) {
	// #1
	{
		datas := []EntryHint{
			{Fid: 0, Offset: 0, Size: 0, Tstamp: 0},
			{Fid: 123, Offset: 777, Size: 666, Tstamp: time.Now().UnixMilli()},
		}

		for _, hint := range datas {
			marshalHint, err := MarshalHint(hint)
			assert.Nil(t, err)
			assert.True(t, len(marshalHint) > 0)

			unMarshalHint, err := UnMarshalHint(marshalHint)
			assert.Nil(t, err)
			assert.Equal(t, hint.Fid, unMarshalHint.Fid)
			assert.Equal(t, hint.Size, unMarshalHint.Size)
			assert.Equal(t, hint.Offset, unMarshalHint.Offset)
			assert.Equal(t, hint.Tstamp, unMarshalHint.Tstamp)
		}
	}
}
