author=$(shell git config user.name)
build_version=$(shell git describe --tags --always)