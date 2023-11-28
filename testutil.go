package riverdb

import (
	"fmt"
	"github.com/bwmarrin/snowflake"
	"math/rand"
	"os"
	"path/filepath"
)

// testDB return a test db for testing
func testDB(option Options) (db *DB, closeDB func() error, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}

	dir := filepath.Join(homeDir, fmt.Sprintf("test_db_%s", string(testRandKV().testUniqueBytes(10))))
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		return nil, nil, err
	}
	db, err = Open(option, WithDir(dir))
	if err != nil {
		return nil, nil, err
	}
	return db, func() error {
		if err := db.Purge(); err != nil {
			return err
		}
		if err := db.Close(); err != nil {
			return err
		}
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
		return nil
	}, nil
}

func testRandKV() *randKV {
	node := testRandNode()
	return &randKV{node}
}

type randKV struct {
	node *snowflake.Node
}

func (r *randKV) testBytes(ln int) []byte {
	ln = rand.Intn(ln)
	key := make([]byte, 0, ln)
	for i := 0; i < ln; i++ {
		key = append(key, byte('a'+rand.Int()%26))
	}
	return key
}

// use snowflake to make sure that bytes is unique
func (r *randKV) testUniqueBytes(ln int) []byte {
	return append(r.node.Generate().Bytes(), r.testBytes(ln)...)
}

func testRandNode() *snowflake.Node {
	node, err := snowflake.NewNode(int64(rand.Int() % 1024))
	if err != nil {
		panic(err)
	}
	return node
}
