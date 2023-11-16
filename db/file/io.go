package file

// IO is the abstract of data file, define a set of interface methods to interact with data file
// how to read and write files depending on the IO underlying implementation
type IO interface {
	// Name returns the name of the file
	Name() string

	// ReadAt read bytes from io start at offset
	ReadAt(p []byte, off int64) (n int, err error)

	// Write bytes to io
	Write(p []byte) (n int, err error)

	// Sync buffer to disk
	Sync() error

	// Close io and release resources
	Close() error
}
