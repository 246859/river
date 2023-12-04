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

go to view the [benchmark](benchmark.md)