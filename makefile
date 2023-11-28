author=$(shell git config user.name)
build_version=$(shell git describe --tags --always)


test:
	go test -v ./...


build:
	go build -v --trimpath ./...

.PHONY:	test, build