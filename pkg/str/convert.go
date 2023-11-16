package str

import "unsafe"

// StringToBytes return the underlying []byte of string, returned bytes must not be modified
func StringToBytes(str string) []byte {
	ptr := unsafe.StringData(str)
	return unsafe.Slice(ptr, len(str))
}

// BytesToString return the string representation of the bytes, if the bytes changes, string changes too
func BytesToString(bytes []byte) string {
	return unsafe.String(unsafe.SliceData(bytes), len(bytes))
}
