# RiverDB
![Static Badge](https://img.shields.io/badge/go-%3E%3D1.21-blue)
![GitHub License](https://img.shields.io/github/license/246859/river)

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

RiverDB is a light-weight embeddable key-value nosql database, it is base on bitcask model and wal. Features as follows:

- ACID transactions
- record ttl
- custom key sorting rules
- range matching and iteration
- event watcher
- batch write and delete
- targzip backup and recover from backup

RiverDB can be used as a standalone database or as an underlying storage engine.

> The project is still under testing and stability cannot be guaranteed

## install

it is a embeddable db, so you can use it in your code without network transportation.

```sh
go get -u github.com/246859/river
```



## how to use

### quick start

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

Remember to close db after used up.

### iteration

riverdb iteration is key-only.

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

    db.Range(riverdb.RangeOptions{
       Min:     nil,
       Max:     nil,
       Pattern: nil,
       Descend: false,
    }, func(key riverdb.Key) bool {
       fmt.Println(key)
       return false
    })
}
```

### transaction

simplely use transaction by `Begin`, `Commit`, `RollBack` APIs.

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

### batch operation

batch operation has better performance than call `db.Put` or `db.Del` directly in large amount of data

```go
import (
	"fmt"
	riverdb "github.com/246859/river"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	// open db
	db, err := riverdb.Open(riverdb.DefaultOptions, riverdb.WithDir(filepath.Join(os.TempDir(), "example")))
	if err != nil {
		panic(err)
	}
	// close
	defer db.Close()

	// open batch
	batch, err := db.Batch(riverdb.BatchOption{
		Size:        500,
		SyncOnFlush: true,
	})

	var rs []riverdb.Record
	var ks []riverdb.Key

	for i := 0; i < 1000; i++ {
		rs = append(rs, riverdb.Record{
			K:   []byte(strconv.Itoa(i)),
			V:   []byte(strings.Repeat("a", i)),
			TTL: 0,
		})
		ks = append(ks, rs[i].K)
	}

	// write all
	if err := batch.WriteAll(rs); err != nil {
		panic(err)
	}

	// delete all
	if err := batch.DeleteAll(ks); err != nil {
		panic(err)
	}

	// wait to batch finished
	if err := batch.Flush(); err != nil {
		panic(err)
	}

	// 2000
	fmt.Println(batch.Effected())
}
```



### backup & recover

backup only archive data in datadir

```go
import (
	riverdb "github.com/246859/river"
	"os"
	"path/filepath"
)

func main() {
	// open db
	db, err := riverdb.Open(riverdb.DefaultOptions, riverdb.WithDir(filepath.Join(os.TempDir(), "example")))
	if err != nil {
		panic(err)
	}
	// close
	defer db.Close()

	archive := filepath.Join(os.TempDir(), "example.tar.gz")
	err = db.Backup(archive)
	if err != nil {
		panic(err)
	}

	err = db.Recover(archive)
	if err != nil {
		panic(err)
	}
}
```



### statistic

```go
func main() {
	// open db
	db, err := riverdb.Open(riverdb.DefaultOptions, riverdb.WithDir(filepath.Join(os.TempDir(), "example")))
	if err != nil {
		panic(err)
	}
	// close
	defer db.Close()
	
    // statistic
	stats := db.Stats()
	fmt.Println(stats.DataSize)
	fmt.Println(stats.HintSize)
	fmt.Println(stats.KeyNums)
	fmt.Println(stats.RecordNums)
}
```



### watcher

you can modfiy which event to watch in db option

```go
import (
	"fmt"
	riverdb "github.com/246859/river"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func main() {
	// open db
	db, err := riverdb.Open(riverdb.DefaultOptions, riverdb.WithDir(filepath.Join(os.TempDir(), "example")))
	defer db.Close()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	watcher, err := db.Watcher(riverdb.PutEvent)
	if err != nil {
		panic(err)
	}

	db.Put([]byte("hello world"), []byte("world"), 0)

	go func() {
		defer wg.Done()
		listen, err := watcher.Listen()
		if err != nil {
			panic(err)
		}

		for event := range listen {
			fmt.Println(event)
		}
	}()

	time.Sleep(time.Second)
	watcher.Close()

	wg.Wait()
}
```



### merge

you can use `db.Merge` to proactively clean up redundant data in the database.

```go
import (
	riverdb "github.com/246859/river"
	"os"
	"path/filepath"
)

func main() {
	// open db
	db, err := riverdb.Open(riverdb.DefaultOptions, riverdb.WithDir(filepath.Join(os.TempDir(), "example")))
	defer db.Close()
	if err != nil {
		panic(err)
	}

	db.Merge(true)
}
```

set `Options.MergeCheckup=0` if you want to disable the default merge check up job.



## benchmark

```
goos: windows
goarch: amd64
pkg: github.com/246859/river
cpu: 11th Gen Intel(R) Core(TM) i7-11800H @ 2.30GHz
BenchmarkDB_B
BenchmarkDB_B/benchmarkDB_Put_value_16B
BenchmarkDB_B/benchmarkDB_Put_value_16B-16                108243             11050 ns/op           25658 B/op         27 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_64B
BenchmarkDB_B/benchmarkDB_Put_value_64B-16                 91654             11353 ns/op           25688 B/op         27 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_128B
BenchmarkDB_B/benchmarkDB_Put_value_128B-16                99914             12012 ns/op           25744 B/op         27 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_256B
BenchmarkDB_B/benchmarkDB_Put_value_256B-16                89656             12834 ns/op           25929 B/op         28 allocs/op
BenchmarkDB_B/benchmarkDB_Put_value_512B
BenchmarkDB_B/benchmarkDB_Put_value_512B-16                79682             14194 ns/op           26328 B/op         28 allocs/op
BenchmarkDB_B/benchmarkDB_Get_B
BenchmarkDB_B/benchmarkDB_Get_B-16                        472155              2478 ns/op            1322 B/op          5 allocs/op
BenchmarkDB_B/benchmarkDB_Del_B
BenchmarkDB_B/benchmarkDB_Del_B-16                        642699              1752 ns/op             427 B/op          7 allocs/op
BenchmarkDB_KB
BenchmarkDB_KB/benchmarkDB_Put_value_1KB
BenchmarkDB_KB/benchmarkDB_Put_value_1KB-16                58642             20106 ns/op           26844 B/op         28 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_16KB
BenchmarkDB_KB/benchmarkDB_Put_value_16KB-16               13563             85796 ns/op           45704 B/op         28 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_64KB
BenchmarkDB_KB/benchmarkDB_Put_value_64KB-16                3909            296592 ns/op          151887 B/op         32 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_256KB
BenchmarkDB_KB/benchmarkDB_Put_value_256KB-16               1083           1208055 ns/op          725232 B/op         39 allocs/op
BenchmarkDB_KB/benchmarkDB_Put_value_512KB
BenchmarkDB_KB/benchmarkDB_Put_value_512KB-16                466           2341486 ns/op         1636059 B/op         45 allocs/op
BenchmarkDB_KB/benchmarkDB_Get_KB
BenchmarkDB_KB/benchmarkDB_Get_KB-16                       38292             27245 ns/op           41288 B/op          5 allocs/op
BenchmarkDB_KB/benchmarkDB_Del_KB
BenchmarkDB_KB/benchmarkDB_Del_KB-16                     1091222              1041 ns/op             266 B/op          6 allocs/op
BenchmarkDB_MB
BenchmarkDB_MB/benchmarkDB_Put_value_1MB
BenchmarkDB_MB/benchmarkDB_Put_value_1MB-16                  217           4825615 ns/op         1607356 B/op         50 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_8MB
BenchmarkDB_MB/benchmarkDB_Put_value_8MB-16                   94          35765929 ns/op        16788028 B/op        164 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_16MB
BenchmarkDB_MB/benchmarkDB_Put_value_16MB-16                  27          63377081 ns/op        41576377 B/op        259 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_32MB
BenchmarkDB_MB/benchmarkDB_Put_value_32MB-16                  37         170156170 ns/op        117083367 B/op       651 allocs/op
BenchmarkDB_MB/benchmarkDB_Put_value_64MB
BenchmarkDB_MB/benchmarkDB_Put_value_64MB-16                   4         250742625 ns/op        194236838 B/op       901 allocs/op
BenchmarkDB_MB/benchmarkDB_Get_MB
BenchmarkDB_MB/benchmarkDB_Get_MB-16                        2660            381140 ns/op          805636 B/op         14 allocs/op
BenchmarkDB_MB/benchmarkDB_Del_MB
BenchmarkDB_MB/benchmarkDB_Del_MB-16                     1749214               669.4 ns/op           234 B/op          6 allocs/op
BenchmarkDB_Mixed
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_64B
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_64B-16             96477             11841 ns/op           25712 B/op         27 allocs/op
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_128KB
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_128KB-16            2374            611411 ns/op          314365 B/op         35 allocs/op
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_MB
BenchmarkDB_Mixed/benchmarkDB_Put_Mixed_MB-16                226           4977799 ns/op         3562019 B/op         57 allocs/op
BenchmarkDB_Mixed/benchmarkDB_Get_Mixed
BenchmarkDB_Mixed/benchmarkDB_Get_Mixed-16                 60637             20533 ns/op           34072 B/op          5 allocs/op
BenchmarkDB_Mixed/benchmarkDB_Del_Mixed
BenchmarkDB_Mixed/benchmarkDB_Del_Mixed-16               1073124              1046 ns/op             262 B/op          6 allocs/op
PASS
ok      github.com/246859/river 49.126s
```