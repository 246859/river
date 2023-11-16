package entry

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
			{Key: []byte("hello world!"), Value: []byte("world!"), Type: DataEntryType, Batch: 1, TTL: time.Now().UnixMilli()},
			{Key: []byte(""), Value: []byte(""), Type: DeletedEntryType, Batch: 2, TTL: time.Now().UnixMilli()},
			{Key: []byte("hello world!ddddddddddddddddddddddddddddddddddddddddd"), Value: []byte("zcxzxcz"), Type: DataEntryType, Batch: 3, TTL: time.Now().UnixMilli()},
			{Key: []byte("bbc"), Value: []byte(strings.Repeat("a", math.MaxInt32>>5)), Type: DataEntryType, Batch: 4, TTL: time.Now().UnixMilli()},
		}

		for _, entry := range datas {
			entryData, err := MarshalEntry(entry)
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
		datas := []Header{
			{Type: DataEntryType, TTL: time.Now().UnixMilli(), Batch: 1, Ksz: 10, Vsz: 20},
			{Type: DeletedEntryType, TTL: time.Now().UnixMilli(), Batch: 1, Ksz: 10, Vsz: 20},
			{Type: BatchFinishedEntryType, TTL: time.Now().UnixMilli(), Batch: 1, Ksz: 10, Vsz: 20},
			{Type: DataEntryType, TTL: time.Now().UnixMilli(), Batch: 10, Ksz: 10, Vsz: 10000},
			{Type: DataEntryType, TTL: time.Now().UnixMilli(), Batch: 1, Ksz: 100, Vsz: 2000},
		}

		for _, header := range datas {
			marshalHeader, offset, err := MarshalHeader(header)
			assert.Nil(t, err)
			assert.True(t, len(marshalHeader) > 0)
			assert.True(t, offset > 0)
			assert.True(t, len(marshalHeader) <= MaxHeaderSize)

			unMarshalHeader, uoffset, err := UnMarshalHeader(marshalHeader)
			assert.Nil(t, err)
			assert.Equal(t, header.Ksz, unMarshalHeader.Ksz)
			assert.Equal(t, header.Vsz, unMarshalHeader.Vsz)
			assert.Equal(t, header.Type, unMarshalHeader.Type)
			assert.True(t, uoffset > 0)
		}
	}

	// #2
	{
		header, offset, err := MarshalHeader(Header{Type: 0})
		assert.ErrorIs(t, err, ErrInvalidEntryType)
		assert.Len(t, header, 0)
		assert.Equal(t, 0, offset)
	}
}

func TestHint_Marshal_UnMarshal(t *testing.T) {
	// #1
	{
		datas := []Hint{
			{Fid: 0, Block: 0, Offset: 0, TTL: time.Now().UnixMilli()},
			{Fid: 1, Block: 2, Offset: 3, TTL: time.Now().UnixMilli()},
			{Fid: 4, Block: 5, Offset: 6, TTL: time.Now().UnixMilli()},
			{Fid: 7, Block: 8, Offset: 9, TTL: time.Now().UnixMilli()},
		}

		for _, hint := range datas {
			marshalHint, err := MarshalHint(hint)
			assert.Nil(t, err)
			assert.True(t, len(marshalHint) > 0)

			unMarshalHint, err := UnMarshalHint(marshalHint)
			assert.Nil(t, err)
			assert.Equal(t, hint.Fid, unMarshalHint.Fid)
			assert.Equal(t, hint.Block, unMarshalHint.Block)
			assert.Equal(t, hint.Offset, unMarshalHint.Offset)
			assert.Equal(t, hint.TTL, unMarshalHint.TTL)
		}
	}
}
