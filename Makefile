SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

GO ?= go

all: build

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: tidy
tidy:
	go fmt -mod=readonly ./...
	go mod tidy -v

.PHONY: build
build: bin/gogo-gb

.PHONY: clean
clean:
	$(GO) clean
	rm -f bin/gogo-gb

.PHONY: run
run:
	$(GO) run -mod=readonly .

.PHONY: test
test:
	$(GO) test -mod=readonly -v ./...

.PHONY: bin/gogo-gb # This does exist, but we're not tracking its dependencies. Go is
bin/gogo-gb:
	$(GO) build -mod=readonly -o bin/gogo-gb .

.PHONY: cpu_instrs
cpu_instrs: bin/gogo-gb vendor/gameboy-doctor/gameboy-doctor vendor/gb-test-roms/cpu_instrs/individual/*.gb
#  These are broken upstream:
#    02-interrupts.gb
	@CPU_TESTS=( \
    "01-special.gb" \
    "03-op sp,hl.gb" \
    "04-op r,imm.gb" \
    "05-op rp.gb" \
    "06-ld r,r.gb" \
    "07-jr,jp,call,ret,rst.gb" \
    "08-misc instrs.gb" \
    "09-op r,r.gb" \
    "10-bit ops.gb" \
    "11-op a,(hl).gb" \
  ); \
  for file in "$${CPU_TESTS[@]}"; do \
    test_name=$${file%*.gb}; \
    test_num=$$((10#$${test_name%-*})); \
    echo "=== Starting cpu_instrs test $$file ==="; \
    bin/gogo-gb --cart "vendor/gb-test-roms/cpu_instrs/individual/$$file" \
                --debugger=gameboy-doctor \
                --log=stderr | \
      ./vendor/gameboy-doctor/gameboy-doctor - cpu_instrs "$$test_num" || \
      { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }; \
    echo "=== Finished cpu_instrs test $$file ===" ; \
  done

.PHONY: dmg_acid2
dmg_acid2: bin/gogo-gb vendor/dmg-acid2/dmg-acid2.gb
	bin/gogo-gb --cart "vendor/dmg-acid2/dmg-acid2.gb" \
              --log=stderr \
              --serial-port=stdout \
              --ui

vendor/dmg-acid2/dmg-acid2.gb:
	mkdir -p vendor/dmg-acid2
	curl -fSsL https://github.com/mattcurrie/dmg-acid2/releases/download/v1.0/dmg-acid2.gb > vendor/dmg-acid2/dmg-acid2.gb

vendor/gameboy-doctor/gameboy-doctor:
	git submodule init
	git submodule update

vendor/gb-test-roms/cpu_instrs/individual/*.gb:
	git submodule init
	git submodule update
