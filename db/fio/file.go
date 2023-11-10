package fio

import (
	"github.com/246859/river/pkg/file"
	"os"
)

func NewFileIO(filename string) (*File, error) {
	fio := &File{}
	fd, err := file.Open(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return fio, err
	}
	fio.fd = fd
	return fio, nil
}

type File struct {
	fd *os.File
}

func (f File) Read(p []byte, off int64) (n int, err error) {
	return f.fd.ReadAt(p, off)
}

func (f File) Write(p []byte) (n int, err error) {
	return f.fd.Write(p)
}

func (f File) Sync() error {
	return f.fd.Sync()
}

func (f File) Close() error {
	return f.fd.Close()
}
