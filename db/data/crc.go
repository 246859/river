package data

import (
	"hash/crc32"
	"hash/crc64"
)

// Crc32 generate 32-bit CRC sum
type Crc32 func(data []byte) uint32

func IEEECrcSum(data []byte) uint32 {
	return crc32.Checksum(data, crc32.MakeTable(crc32.IEEE))
}

func CastagnoliCrcSum(data []byte) uint32 {
	return crc32.Checksum(data, crc32.MakeTable(crc32.Castagnoli))
}

func KoopmanCrcSum(data []byte) uint32 {
	return crc32.Checksum(data, crc32.MakeTable(crc32.Koopman))
}

// Crc64 generate 64-bit CRC sum
type Crc64 func(data []byte) uint64

func ISOrcSum(data []byte) uint64 {
	return crc64.Checksum(data, crc64.MakeTable(crc64.ISO))
}

func ECMACrcSum(data []byte) uint64 {
	return crc64.Checksum(data, crc64.MakeTable(crc64.ECMA))
}
