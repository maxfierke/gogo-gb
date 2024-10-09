SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

GO ?= go
MOONEYE_TEST_SUITE_VERION ?= mts-20240127-1204-74ae166

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
	$(GO) clean
	rm -f bin/gogo-gb

.PHONY: run
run:
	$(GO) run .

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: bin/gogo-gb # This does exist, but we're not tracking its dependencies. Go is
bin/gogo-gb:
	$(GO) build -o bin/gogo-gb .

.PHONY: cpu_instrs
cpu_instrs: bin/gogo-gb tests/gameboy-doctor/gameboy-doctor tests/gb-test-roms/cpu_instrs/individual/*.gb
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
    bin/gogo-gb --cart "tests/gb-test-roms/cpu_instrs/individual/$$file" \
                --skip-bootrom \
                --debugger=gameboy-doctor \
                --log=stderr | \
      ./tests/gameboy-doctor/gameboy-doctor - cpu_instrs "$$test_num" || \
      { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }; \
    echo "=== Finished cpu_instrs test $$file ===" ; \
  done

.PHONY: mem_timing
mem_timing: bin/gogo-gb tests/gb-test-roms/mem_timing/individual/*.gb
	@MEM_TESTS=( \
    "01-read_timing.gb" \
    "02-write_timing.gb" \
    "03-modify_timing.gb" \
  ); \
  for file in "$${MEM_TESTS[@]}"; do \
    test_name=$${file%*.gb}; \
    test_num=$$((10#$${test_name%-*})); \
    echo "=== WARNING: WIP, these will hang ==="; \
    echo "=== Starting mem_timing test $$file ==="; \
    bin/gogo-gb --cart "tests/gb-test-roms/mem_timing/individual/$$file" \
                --log=stderr --serial-port=stdout; \
    echo "=== Finished mem_timing test $$file ===" ; \
  done

.PHONY: dmg_acid2
dmg_acid2: bin/gogo-gb tests/dmg-acid2/dmg-acid2.gb
	bin/gogo-gb --cart "tests/dmg-acid2/dmg-acid2.gb" \
              --log=stderr \
              --serial-port=stdout \
              --ui

.PHONY: mealybug_tests
mealybug_tests: tests/mealybug-tearoom-tests/build/ppu/*.gb
	bin/gogo-gb --cart "tests/mealybug-tearoom-tests/build/ppu/m3_scy_change.gb" \
              --log=stderr \
              --serial-port=stdout \
              --ui

.PHONY: mooneye_gb_tests
mooneye_gb_tests: tests/mooneye-gb-test-suite/acceptance/*.gb
	bin/gogo-gb --cart "tests/mooneye-gb-test-suite/acceptance/ppu/hblank_ly_scx_timing-GS.gb" \
              --log=stderr \
              --serial-port=stdout \
              --ui

tests/dmg-acid2/dmg-acid2.gb:
	mkdir -p tests/dmg-acid2
	curl -fSsL https://github.com/mattcurrie/dmg-acid2/releases/download/v1.0/dmg-acid2.gb > tests/dmg-acid2/dmg-acid2.gb

tests/gameboy-doctor/gameboy-doctor:
	git submodule init
	git submodule update

tests/gb-test-roms/cpu_instrs/individual/*.gb:
	git submodule init
	git submodule update

tests/gb-test-roms/mem_timing/individual/*.gb:
	git submodule init
	git submodule update

tests/mooneye-gb-test-suite/acceptance/*.gb:
	mkdir -p tests/mooneye-gb-test-suite
	curl -fSsL https://gekkio.fi/files/mooneye-test-suite/$(MOONEYE_TEST_SUITE_VERION)/$(MOONEYE_TEST_SUITE_VERION).tar.xz > tests/mooneye-gb-test-suite/$(MOONEYE_TEST_SUITE_VERION).tar.xz
	tar -xf tests/mooneye-gb-test-suite/$(MOONEYE_TEST_SUITE_VERION).tar.xz -C tests/mooneye-gb-test-suite --strip-components=1

tests/mealybug-tearoom-tests/build/ppu/*.gb:
	git submodule init
	git submodule update --recursive
	mkdir -p tests/mealybug-tearoom-tests/build/ppu
	cd tests/mealybug-tearoom-tests && unzip mealybug-tearoom-tests.zip -d build/ppu
