build_version=$(shell git describe --tags --abbrev=0 --always)

build:
	go build -a -o ./bin/ -ldflags="-X main.Version=$(build_version)"

genpb:
	buf generate

.PHONY: build, genpb