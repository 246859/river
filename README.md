# RiverDB

```
 ███████   ██                         ███████   ██████
░██░░░░██ ░░             To the moon ░██░░░░██ ░█░░░░██
░██   ░██  ██ ██    ██  █████  ██████░██    ░██░█   ░██
░███████  ░██░██   ░██ ██░░░██░░██░░█░██    ░██░██████
░██░░░██  ░██░░██ ░██ ░███████ ░██ ░ ░██    ░██░█░░░░ ██
░██  ░░██ ░██ ░░████  ░██░░░░  ░██   ░██    ██ ░█    ░██
░██   ░░██░██  ░░██   ░░██████░███   ░███████  ░███████
░░     ░░ ░░    ░░     ░░░░░░ ░░░    ░░░░░░░   ░░░░░░░
```

RiverDB is a light-weight embeddable key-value nosql database, it is base on bitcask model. Features as follows:

- Using memory for data indexing
- Using WAL to store actual data on disk
- Support for ACID transactions
- Support zip backup and recover from zip
- Support data ttl
- Support custom sorting rules
- Support range matching and iteration
- Support event watcher

RiverDB can be used as a standalone database or as an underlying storage engine.

> The project is still under testing and stability cannot be guaranteed

## install

it is a embeddable db, so you can use it in your code without network transportation.

```sh
go get -u github.com/246859/river
```



## use

this is a simple example for use put and get operation.

```go
import (
	"fmt"
	riverdb "github.com/246859/river"
)

func main() {
	// open the river db
	db, err := riverdb.Open(riverdb.DefaultOptions, riverdb.WithDir("riverdb"))
	if err != nil {
		panic(err)
	}
    defer db.Close()
	// put key-value pairs
	err = db.Put([]byte("key"), []byte("value"), 0)
	if err != nil {
		panic(err)
	}

	// get value from key
	value, err := db.Get([]byte("key"))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(value))
}

```

simple use for transaction by `Begin`, `Commit`, `RollBack` APIs.

```go
import (
	"errors"
	"fmt"
	riverdb "github.com/246859/river"
)

func main() {
	// open the river db
	db, err := riverdb.Open(riverdb.DefaultOptions, riverdb.WithDir("riverdb"))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	txn, err := db.Begin(true)
	if err != nil {
		panic(err)
	}
	_, err = txn.Get([]byte("notfund"))
	if errors.Is(err, riverdb.ErrKeyNotFound) {
		fmt.Println("key not found")
	}else {
		txn.RollBack()
	}
	txn.Commit()
}

```

Remember to close db after used up.


## benchmark
```
goos: windows
goarch: amd64
pkg: github.com/246859/river
cpu: 11th Gen Intel(R) Core(TM) i7-11800H @ 2.30GHz
BenchmarkDB_B/benchmarkDB_Put_value_64B-16                214749              5798 ns/op            1650 B/op         36 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_128B-16               160741              6965 ns/op            1717 B/op         36 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_256B-16               136750              7750 ns/op            1866 B/op         37 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_512B-16               120902              8818 ns/op            2223 B/op         37 allocs/op
BenchmarkDB_B/benchmarkDB_Get_B-16                        447379              2560 ns/op            1382 B/op          2 allocs/op
BenchmarkDB_B/benchmarkDB_Del_B-16                        639976              1809 ns/op             632 B/op         11 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_1KB-16               130693             11007 ns/op            2666 B/op         37 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_64KB-16                4690            321066 ns/op          131357 B/op         41 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_256KB-16               1104           1126429 ns/op          675491 B/op         47 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_512KB-16                512           2308804 ns/op         1620989 B/op         53 allocs/op
BenchmarkDB_KB/benchmarkDB_Get_KB-16                       15399             69486 ns/op          133335 B/op          3 allocs/op
BenchmarkDB_KB/benchmarkDB_Del_KB-16                     1007193              1156 ns/op             627 B/op         11 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_1MB-16                  273           4328178 ns/op         1154033 B/op         53 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_8MB-16                   37          39345992 ns/op        10538434 B/op        174 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_16MB-16                  14          73422429 ns/op        28108471 B/op        304 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_32MB-16                   9         161750389 ns/op        72378227 B/op        589 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_64MB-16                   5         344447860 ns/op        214184576 B/op      1214 allocs/op
BenchmarkDB_MB/benchmarkDB_Get_MB-16                         506           7137680 ns/op        15719500 B/op        205 allocs/op
BenchmarkDB_MB/benchmarkDB_Del_MB-16                     1711191               691.2 ns/op           626 B/op         10 allocs/op
PASS
ok      github.com/246859/river 35.047s
```