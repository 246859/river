package crc

import (
	"github.com/pkg/errors"
	"hash/crc32"
	"hash/crc64"
)

var (
	ErrCrcInvalid = errors.New("invalid crc")
)

var (
	_IEEE32       = crc32.MakeTable(crc32.IEEE)
	_Castagnoli32 = crc32.MakeTable(crc32.Castagnoli)
	_Koopman32    = crc32.MakeTable(crc32.Koopman)

	_ISO64  = crc64.MakeTable(crc64.ISO)
	_ECMA64 = crc64.MakeTable(crc64.ECMA)
)

var (
	IEEE32       Crc32 = crc32I{_IEEE32}
	Castagnoli32 Crc32 = crc32I{_Castagnoli32}
	Koopman32    Crc32 = crc32I{_Koopman32}

	ISO64  Crc64 = crc64I{_ISO64}
	ECMA64 Crc64 = crc64I{_ECMA64}
)

// Crc32 generate 32-bit CRC sum
type Crc32 interface {
	Sum(data []byte) uint32
	Update(crc uint32, data []byte) uint32
}

type Crc64 interface {
	Sum(data []byte) uint64
	Update(crc uint64, data []byte) uint64
}

type crc32I struct {
	t *crc32.Table
}

func (c crc32I) Sum(data []byte) uint32 {
	return crc32.Checksum(data, c.t)
}

func (c crc32I) Update(crc uint32, data []byte) uint32 {
	return crc32.Update(crc, c.t, data)
}

type crc64I struct {
	t *crc64.Table
}

func (c crc64I) Sum(data []byte) uint64 {
	return crc64.Checksum(data, c.t)
}

func (c crc64I) Update(crc uint64, data []byte) uint64 {
	return crc64.Update(crc, c.t, data)
}
