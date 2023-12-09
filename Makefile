SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

SRCS := main.go cpu/cpu.go cpu/registers.go cpu/isa/instruction.go cpu/isa/opcodes.go go.mod

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
	rm -f bin/gogo-gb

.PHONY: run
run:
	go run .

bin/gogo-gb: $(SRCS)
	go build -o bin/gogo-gb .

all: build
