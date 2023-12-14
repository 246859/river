build_version=$(shell git describe --tags --abbrev=0 --always)

build_win:
	SET CGO_ENABLED=0
	SET GOOS="win"
	go build -a -o ./bin/river-server.exe -ldflags="-X main.Version=$(build_version)"

build_linux:
	SET CGO_ENABLED=0
	SET GOOS="linux"
	go build -a -o ./bin/river-server -ldflags="-X main.Version=$(build_version)"

genpb:
	buf generate

.PHONY: build, genpb