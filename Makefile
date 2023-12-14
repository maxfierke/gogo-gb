SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

all: build

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

.PHONY: build
build: bin/gogo-gb

.PHONY: clean
clean:
	go clean
	rm -f bin/gogo-gb

.PHONY: run
run:
	go run .

.PHONY: bin/gogo-gb # This does exist, but we're not tracking its dependencies. Go is
bin/gogo-gb:
	go build -o bin/gogo-gb .
