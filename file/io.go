package file

import (
	"os"
)

const (
	B = 1 << (iota * 10)
	KB
	MB
	GB
	TB
	PB
)

const (
	RwAppendMode = os.O_CREATE | os.O_APPEND | os.O_RDWR
)

// IO is the abstract of data file, define a set of interface methods to interact with data file
// how to read and write files depending on the IO underlying implementation
type IO interface {
	// Name returns the name of the file
	Name() string

	// ReadAt read bytes from io start at offset
	ReadAt(p []byte, off int64) (n int, err error)

	// Write bytes to io
	Write(p []byte) (n int, err error)

	Seek(offset int64, whence int) (ret int64, err error)
	// Sync buffer to disk
	Sync() error

	// Close io and release resources
	Close() error
}
