package main

import (
	"context"
	"flag"
	"github.com/246859/river/assets"
	"github.com/246859/river/cmd/grpc/server"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var configfile string

func init() {
	flag.StringVar(&configfile, "f", "", "specified configuration file")
}

func main() {
	// listen the os signals
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	// log banner
	_ = assets.OutFsFile(assets.Fs, "banner.txt", os.Stdout)

	riverServer, err := server.NewServer(ctx, configfile)
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
