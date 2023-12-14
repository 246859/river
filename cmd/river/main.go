package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/246859/river/assets"
	"github.com/246859/river/cmd/river/server"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var (
	Version    string
	ConfigFile string
)

func init() {
	flag.StringVar(&ConfigFile, "f", "", "specified configuration file")
	flag.Parse()
}

func main() {
	// listen the os signals
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	// log banner
	_ = assets.OutFsFile(assets.Fs, "banner.txt", os.Stdout)

	// red config file
	var option server.Options
	if len(ConfigFile) > 0 {
		options, err := server.ReadOption(ConfigFile)
		if err != nil {
			slog.Error("", "err", err)
			return
		}
		slog.Info("config file loaded √")
		option = options
	} else {
		slog.Info("no config file specified")
	}
	option.Version = Version

	// initialize the server
	riverServer, err := server.NewServer(ctx, option)
	if err != nil {
		slog.Error(fmt.Sprintf("river grpc server boot failed %s", err.Error()))
	}
	defer func() {
		if riverServer != nil {
			riverServer.Close()
		}
	}()

	// run the server in another goroutine
	runCh := make(chan error)
	go func() {
		runCh <- riverServer.Run()
		close(runCh)
	}()

	// close when recv os ignals or run failed
	select {
	case <-ctx.Done():
		if riverServer != nil {
			riverServer.Close()
		}
	case err := <-runCh:
		slog.Error("river server run failed", "error", err)
	}
}
