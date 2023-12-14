# river-server
```
 ███████   ██                         ███████   ██████
░██░░░░██ ░░             To the moon ░██░░░░██ ░█░░░░██
░██   ░██  ██ ██    ██  █████  ██████░██    ░██░█   ░██
░███████  ░██░██   ░██ ██░░░██░░██░░█░██    ░██░██████
░██░░░██  ░██░░██ ░██ ░███████ ░██ ░ ░██    ░██░█░░░░ ██
░██  ░░██ ░██ ░░████  ░██░░░░  ░██   ░██    ██ ░█    ░██
░██   ░░██░██  ░░██   ░░██████░███   ░███████  ░███████
░░     ░░ ░░    ░░     ░░░░░░ ░░░    ░░░░░░░   ░░░░░░░

2023/12/14 16:36:33 INFO no config file specified
2023/12/14 16:36:33 INFO version: v0.3.1
2023/12/14 16:36:33 INFO db server initializing...
2023/12/14 16:36:33 INFO db loaded √
2023/12/14 16:36:33 INFO grpc server loaded √
2023/12/14 16:36:33 INFO river server listening at :6868
2023/12/14 16:36:34 INFO river server closed
```
river-server is a simple grpc server written in go.



## install

```sh
go install github.com/246859/river/cmd/river@latest
```

if you don't have go environment, you can download binary-package in release page.



## config

config file formt must be TOML, default config as follows.

```toml
# server address
address = ":6868"

# server password
password = ""

# max data files size, default 64MB
max_file_size = 67108864

# how many blocks will be cached, default 64
block_cache = 64

# call fsync on per write operation
fsync_per_write = false

# default write threshold to trigger fsync, default 32kb
fsync_threshold = 32768

# transaction level
# 1 -- serializable
# 2 -- readcommitted
txn_level = 2

# checkpoint to trigger merge operation
merge_checkpoint = 4.5

# log output file
log_file = "/var/lib/river/db.log"

# log level
log_level = "INFO"
```



## example

you can interact with river-server by using river-client.A simple example as follows.

```go
import (
	"context"
	"fmt"
	"github.com/246859/river/cmd/river/client"
	"time"
)

func main() {
	timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	riverClient, err := client.NewClient(timeout, client.Options{
		Target:   ":6868",
		Password: "abcdefhijklmnopqrstuvwxyz",
	})

	if err != nil {
		panic(err)
	}

	value, err := riverClient.Get(context.Background(), []byte("key"))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(value))
}
```

