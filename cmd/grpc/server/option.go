package server

import (
	"github.com/BurntSushi/toml"
	"os"
)

const (
	DefaultDir     = "/var/lib/river/db"
	DefaultLogFile = "/var/lib/river/db.log"
)

type Options struct {
	// server listen address
	Address string `toml:"address"`
	// server password
	Password string `toml:"password"`
	// tls configuration
	TlsKey string `toml:"tls_key"`
	TlsPem string `toml:"tls_perm"`

	// max bytes size of the single data file can hold
	MaxSize int64 `toml:"max_file_size"`
	// wal block cache size
	BlockCache uint32 `toml:"block_cache"`
	// call sync per write
	Fsync bool `toml:"fsync"`
	// call sync when reach the threshold
	FsyncThreshold int64 `toml:"fsync_threshold"`
	// transaction isolation level
	TxnLevel uint8 `toml:"txn_level"`
	// merge checkpoint
	Checkpoint float64 `toml:"merge_checkpoint"`

	LogFile  string `toml:"log_file"`
	LogLevel string `toml:"log_level"`
}

// readOption Reads option from the specified path
func readOption(cfgpath string) (Options, error) {
	var opt Options

	if _, err := os.Stat(cfgpath); err != nil {
		return Options{}, err
	}

	raw, err := os.ReadFile(cfgpath)
	if err != nil {
		return Options{}, err
	}

	if err := toml.Unmarshal(raw, &opt); err != nil {
		return opt, err
	}
	return opt, nil
}
