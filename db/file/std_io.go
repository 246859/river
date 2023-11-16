package file

import (
	"os"
)

var _ IO = StdFile{}

func OpenStdFile(filename string) (*StdFile, error) {
	fio := &StdFile{}
	fd, err := Open(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return fio, err
	}
	fio.fd = fd
	return fio, nil
}

type StdFile struct {
	fd *os.File
}

func (f StdFile) Name() string {
	return f.fd.Name()
}

func (f StdFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.fd.ReadAt(p, off)
}

func (f StdFile) Write(p []byte) (n int, err error) {
	return f.fd.Write(p)
}

func (f StdFile) Sync() error {
	return f.fd.Sync()
}

func (f StdFile) Close() error {
	return f.fd.Close()
}