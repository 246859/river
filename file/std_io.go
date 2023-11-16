package file

import (
	"os"
	"path/filepath"
)

var _ IO = StdFile{}

func OpenStdFile(filename string, flag int, mode os.FileMode) (*StdFile, error) {
	fio := &StdFile{}
	fd, err := openStd(filename, flag, mode)
	if err != nil {
		return fio, err
	}
	fio.fd = fd
	return fio, nil
}

func openStd(filename string, flag int, mode os.FileMode) (*os.File, error) {
	dir := filepath.Dir(filename)
	if dir != "." {
		err := os.MkdirAll(dir, mode)
		if err != nil {
			return nil, err
		}
	}
	return os.OpenFile(filename, flag, mode)
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

func (f StdFile) Seek(offset int64, whence int) (ret int64, err error) {
	return f.fd.Seek(offset, whence)
}

func (f StdFile) Sync() error {
	return f.fd.Sync()
}

func (f StdFile) Close() error {
	return f.fd.Close()
}
