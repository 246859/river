author=$(shell git config user.name)
build_version=$(shell git describe --tags --always)

bench:
	go test -v -bench . -run db_bench_test.go

test:
	go test -v ./... -count=1 -skip=Client*


build:
	go build -v --trimpath ./...

.PHONY:	test, build, bench