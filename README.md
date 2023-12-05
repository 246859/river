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
	"fmt"
	riverdb "github.com/246859/river"
	"strings"
)

func main() {
	// open the river db
	db, err := riverdb.Open(riverdb.DefaultOptions, riverdb.WithDir("riverdb"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// read write transactions
	db.Begin(func(txn *riverdb.Txn) error {
		for i := 0; i < 10; i++ {
			db.Put([]byte(strings.Repeat("a", i+1)), []byte(strings.Repeat("a", i+1)), 0)
		}
		return nil
	})

	// read only transactions
	db.View(func(txn *riverdb.Txn) error {
		for i := 0; i < 10; i++ {
			get, err := db.Get([]byte(strings.Repeat("a", i)))
			if err != nil {
				return err
			}
			fmt.Println(string(get))
		}
		return nil
	})
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
BenchmarkDB_Get_1k
BenchmarkDB_Get_1k-16             367923              3188 ns/op            1575 B/op         10 allocs/op
BenchmarkDB_Get_1w
BenchmarkDB_Get_1w-16             364064              3304 ns/op            2140 B/op         10 allocs/op
BenchmarkDB_Get_10w
BenchmarkDB_Get_10w-16            148518              7604 ns/op           12200 B/op         10 allocs/op
BenchmarkDB_Get_100w
BenchmarkDB_Get_100w-16            69440             17838 ns/op           32789 B/op         12 allocs/op
BenchmarkDb_Put_1k
BenchmarkDb_Put_1k-16                 79          13911381 ns/op         5965862 B/op      34059 allocs/op
BenchmarkDb_Put_1w
BenchmarkDb_Put_1w-16                  8         137385188 ns/op        59874926 B/op     340384 allocs/op
BenchmarkDb_Put_10w
BenchmarkDb_Put_10w-16                 1        1328774200 ns/op        618986752 B/op   3407017 allocs/op
BenchmarkDb_Put_100w
BenchmarkDb_Put_100w-16                1        13500072100 ns/op       6189542936 B/op 34068784 allocs/op
BenchmarkDb_Put_256B
BenchmarkDb_Put_256B-16           142414              9996 ns/op            5465 B/op         34 allocs/op
BenchmarkDb_Put_64KB
BenchmarkDb_Put_64KB-16             7975            540736 ns/op           72214 B/op         36 allocs/op
BenchmarkDb_Put_256KB
BenchmarkDb_Put_256KB-16            1705           1671027 ns/op          213913 B/op         40 allocs/op
BenchmarkDb_Put_1MB
BenchmarkDb_Put_1MB-16              1255           2947766 ns/op          535097 B/op         50 allocs/op
BenchmarkDb_Put_4MB
BenchmarkDb_Put_4MB-16               427           5633411 ns/op         1028939 B/op         63 allocs/op
BenchmarkDb_Put_8MB
BenchmarkDb_Put_8MB-16               243           5230816 ns/op          777539 B/op         56 allocs/op
PASS
ok      github.com/246859/river 79.966s
```