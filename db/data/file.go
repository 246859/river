package data

import (
	"github.com/246859/river/db/fio"
)

type File struct {
	Fid    uint32
	offset uint64
	io     fio.IO
}
