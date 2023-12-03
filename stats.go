package riverdb

// Stats represents a simple statistics information of db at a moment
type Stats struct {
	// number of key in db
	KeyNums int64
	// number of values in db, due to bitcask model is append-only, it usually greater than KeyNums.
	// normally, result of RecordNums / KeyNums can be used to determine if is needed to merge the wal files
	RecordNums int64
	// real data size
	DataSize int64
	// hint file size
	HintSize int64
}

func (db *DB) Stats() Stats {
	db.mu.RLock()
	defer db.mu.RUnlock()
	stat := db.data.Stat()
	var stats Stats
	stats.KeyNums = int64(db.index.Size())
	stats.RecordNums = db.numOfRecord
	stats.DataSize = stat.SizeOfWal
	stats.HintSize = stat.SizeOfWal

	return stats
}
