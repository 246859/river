package main

import (
	"context"
	"flag"
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
		option = options
	}
	option.Version = Version

	riverServer, err := server.NewServer(ctx, option)
	if err != nil {
		slog.Error("river grpc server boot failed", err.Error())
	}
	defer riverServer.Close()

	go func() {
		riverServer.Run()
	}()

	select {
	case <-ctx.Done():
		if riverServer != nil {
			riverServer.Close()
		}
	}
}
