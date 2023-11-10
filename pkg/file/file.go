package file

import (
	"os"
	"path/filepath"
)

func Open(filename string, flag int, mode os.FileMode) (*os.File, error) {
	dir := filepath.Dir(filename)
	if dir != "." {
		err := os.MkdirAll(dir, mode)
		if err != nil {
			return nil, err
		}
	}
	return os.OpenFile(filename, flag, mode)
}

func Exist(filepath string) bool {
	_, err := os.Stat(filepath)
	if err != nil {
		return false
	}
	return true
}
