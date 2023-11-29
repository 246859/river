package riverdb

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Backup use tar gzip to compress data wal files to dest path
func (db *DB) Backup(destpath string) error {
	if db.mask.CheckAny(closed) {
		return ErrDBClosed
	}

	if strings.HasPrefix(processPath(destpath), processPath(db.option.Dir)) {
		return fmt.Errorf("archive destination should not in db directory")
	}

	db.opmu.Lock()
	defer db.opmu.Unlock()

	db.mask.Store(backing)
	defer db.mask.Remove(backing)

	// check dir status
	datadir := db.option.dataDir
	_, err := os.Stat(datadir)
	if err != nil {
		return err
	}

	// rotate data wal file
	db.mu.Lock()
	if err := db.data.Rotate(); err != nil {
		db.mu.Unlock()
		return err
	}
	db.mu.Unlock()

	// tar gzip archive
	err = tarCompress(datadir, destpath)
	if err != nil {
		return err
	}

	// notify watcher
	if db.watcher != nil {
		db.watcher.push(&Event{Type: BackupEvent, Value: destpath})
	}
	return nil
}

// Recover recovers wal files from specified targz archive.
// it will purge current data, and overwrite by the backup.
func (db *DB) Recover(srcpath string) error {
	if db.mask.CheckAny(closed) {
		return ErrDBClosed
	}

	db.opmu.Lock()
	defer db.opmu.Unlock()

	db.mask.Store(recovering)
	defer db.mask.Remove(recovering)

	// check src path
	if stat, err := os.Stat(srcpath); err != nil {
		return err
	} else if stat.IsDir() {
		return errors.New("recovering src path must be regular file path")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.closeWal(); err != nil {
		return err
	}

	// remove all original data
	if err := os.RemoveAll(db.option.dataDir); err != nil {
		return err
	}

	// re mkdir
	if err := os.MkdirAll(db.option.dataDir, os.ModePerm); err != nil {
		return err
	}

	// un tar gzip
	if err := unTarCompress(srcpath, db.option.dataDir); err != nil {
		return err
	}

	// reload data
	if err := db.load(); err != nil {
		return err
	}

	// notify watcher
	if db.watcher != nil {
		db.watcher.push(&Event{Type: RecoverEvent, Value: srcpath})
	}
	return nil
}

func tarCompress(src, dest string) error {
	_, err := os.Stat(src)
	if err != nil {
		return err
	}

	dir := filepath.Dir(dest)
	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipwriter := gzip.NewWriter(file)
	defer gzipwriter.Close()

	tarwriter := tar.NewWriter(gzipwriter)
	defer tarwriter.Close()

	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip un-regular file
		if !info.Mode().IsRegular() || info.Mode().IsDir() {
			return nil
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Size = info.Size()
		header.Name = strings.TrimPrefix(processPath(path), processPath(src))

		if err := tarwriter.WriteHeader(header); err != nil {
			return err
		}

		srcfile, err := os.Open(path)
		if err != nil {
			return err
		}

		defer srcfile.Close()

		if _, err := io.Copy(tarwriter, srcfile); err != nil {
			return err
		}

		return nil
	})
}

func unTarCompress(src, dst string) error {
	srctar, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srctar.Close()

	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return err
	}

	gzipreader, err := gzip.NewReader(srctar)
	if err != nil {
		return err
	}
	defer gzipreader.Close()

	tarreader := tar.NewReader(gzipreader)

	for {
		tarheader, err := tarreader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return err
		}

		destpath := filepath.Join(dst, tarheader.Name)

		switch tarheader.Typeflag {
		case tar.TypeReg:
			destdir := filepath.Dir(destpath)
			if err := os.MkdirAll(destdir, 0755); err != nil {
				return err
			}
			destfile, err := os.OpenFile(destpath, os.O_RDWR|os.O_CREATE, os.FileMode(tarheader.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(destfile, tarreader); err != nil {
				destfile.Close()
				return err
			}
			destfile.Close()
		case tar.TypeDir:
			if err := os.MkdirAll(destpath, 0755); err != nil {
				return err
			}
		}
	}
}

func processPath(path string) string {
	if s := strings.Split(path, ":"); len(s) > 1 {
		path = s[1]
	}
	return strings.ReplaceAll(path, "\\", "/")
}
