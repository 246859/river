package data

// RecordPos represents the position of record in which file specified by fid
type RecordPos struct {
	// represents record storage in which file
	Fid uint64
	// record offset of file
	Offset int64
	// record size in file
	Size uint64
}

type RecordHeader struct {
}
