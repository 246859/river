package assets

import (
	"embed"
	"io"
	"io/fs"
)

//go:embed *
var Fs embed.FS

func OutFsFile(fs fs.ReadFileFS, file string, out io.Writer) error {
	readFile, err := fs.ReadFile(file)
	if err != nil {
		return err
	}
	_, err = out.Write(readFile)
	if err != nil {
		return err
	}
	return nil
}
