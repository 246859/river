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