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

.PHONY: test
test:
	go test -v ./...

.PHONY: bin/gogo-gb # This does exist, but we're not tracking its dependencies. Go is
bin/gogo-gb:
	go build -o bin/gogo-gb .

.PHONY: cpu_instrs
cpu_instrs: bin/gogo-gb vendor/gameboy-doctor/gameboy-doctor
#	bin/gogo-gb --cart ./vendor/gb-test-roms/cpu_instrs/individual/01-special.gb | ./vendor/gameboy-doctor/gameboy-doctor - cpu_instrs 1 || { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }
#	bin/gogo-gb --cart ./vendor/gb-test-roms/cpu_instrs/individual/02-interrupts.gb | ./vendor/gameboy-doctor/gameboy-doctor - cpu_instrs 2 || { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }
	bin/gogo-gb --cart ./vendor/gb-test-roms/cpu_instrs/individual/03-op\ sp,hl.gb | ./vendor/gameboy-doctor/gameboy-doctor - cpu_instrs 3 || { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }
#	bin/gogo-gb --cart ./vendor/gb-test-roms/cpu_instrs/individual/04-op\ r,imm.gb | ./vendor/gameboy-doctor/gameboy-doctor - cpu_instrs 4 || { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }
	bin/gogo-gb --cart ./vendor/gb-test-roms/cpu_instrs/individual/05-op\ rp.gb | ./vendor/gameboy-doctor/gameboy-doctor - cpu_instrs 5 || { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }
	bin/gogo-gb --cart ./vendor/gb-test-roms/cpu_instrs/individual/06-ld\ r,r.gb | ./vendor/gameboy-doctor/gameboy-doctor - cpu_instrs 6 || { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }
	bin/gogo-gb --cart ./vendor/gb-test-roms/cpu_instrs/individual/07-jr,jp,call,ret,rst.gb | ./vendor/gameboy-doctor/gameboy-doctor - cpu_instrs 7 || { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }
	bin/gogo-gb --cart ./vendor/gb-test-roms/cpu_instrs/individual/08-misc\ instrs.gb | ./vendor/gameboy-doctor/gameboy-doctor - cpu_instrs 8 || { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }
#	bin/gogo-gb --cart ./vendor/gb-test-roms/cpu_instrs/individual/09-op\ r,r.gb | ./vendor/gameboy-doctor/gameboy-doctor - cpu_instrs 9 || { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }
#	bin/gogo-gb --cart ./vendor/gb-test-roms/cpu_instrs/individual/10-bit\ ops.gb | ./vendor/gameboy-doctor/gameboy-doctor - cpu_instrs 10 || { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }
#	bin/gogo-gb --cart './vendor/gb-test-roms/cpu_instrs/individual/11-op a,(hl).gb' | ./vendor/gameboy-doctor/gameboy-doctor - cpu_instrs 11 || { ec=$$?; [ $$ec -eq 141 ] && true || (exit $$ec); }
