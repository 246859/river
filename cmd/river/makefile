build_version=$(shell git describe --tags --abbrev=0 --always)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

build_win_amd64:
	go env -w GOOS="windows"
	go env -w GOARCH="amd64"
	go build -a -trimpath -o ./bin/river-server.exe -ldflags="-X main.Version=$(build_version)"
	go env -w GOOS="$(GOOS)"
	go env -w GOARCH="$(GOARCH)"

build_linux_amd64:
	go env -w GOOS="linux"
	go env -w GOARCH="amd64"
	go build -a -trimpath -o ./bin/river-server -ldflags="-X main.Version=$(build_version)"
	go env -w GOOS="$(GOOS)"
	go env -w GOARCH="$(GOARCH)"

build_all:
	make build_win_amd64
	make build_linux_amd64

genpb:
	buf generate

.PHONY: build, genpb, build_all