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
	@CPU_TESTS=( \
    "01-special.gb" \
    "02-interrupts.gb" \
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
    bin/gogo-gb run "tests/gb-test-roms/cpu_instrs/individual/$$file" \
                --skip-bootrom \
                --debugger=gameboy-doctor \
                --log=stderr \
                --model=dmg \
                --headless | \
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
    bin/gogo-gb run "tests/gb-test-roms/mem_timing/individual/$$file" \
                --log=stderr --serial-port=stdout --headless; \
    echo "=== Finished mem_timing test $$file ===" ; \
  done

.PHONY: acid2
acid2: dmg_acid2 dmg_acid2_cgb cgb_acid2

.PHONY: dmg_acid2
dmg_acid2: bin/gogo-gb tests/dmg-acid2/dmg-acid2.gb
	@rm -f /tmp/gogo-gb-dmg_acid2-dmg.png
	@bin/gogo-gb run "tests/dmg-acid2/dmg-acid2.gb" \
              --log=stderr \
              --debugger=interactive \
              --debugger-attach=false \
              --debugger-soft-break=screenshot \
              --screenshot-path=/tmp/gogo-gb-dmg_acid2-dmg.png \
              --headless
	@compare -metric AE "/tmp/gogo-gb-dmg_acid2-dmg.png" "tests/dmg-acid2/reference-dmg.png" null: 2>&1 && \
    echo && echo 'OK: Matched reference' || \
    (echo && echo 'FAIL: Bad match to reference' && exit 1)

.PHONY: dmg_acid2_cgb
dmg_acid2_cgb: bin/gogo-gb tests/dmg-acid2/dmg-acid2.gb
	@rm -f /tmp/gogo-gb-dmg_acid2-cgb.png
	@bin/gogo-gb run "tests/dmg-acid2/dmg-acid2.gb" \
              --log=stderr \
              --model=cgb \
              --debugger=interactive \
              --debugger-attach=false \
              --debugger-soft-break=screenshot \
              --screenshot-path=/tmp/gogo-gb-dmg_acid2-cgb.png \
              --headless
	@compare -metric AE "/tmp/gogo-gb-dmg_acid2-cgb.png" "tests/dmg-acid2/reference-cgb.png" null: 2>&1 && \
    echo && echo 'OK: Matched reference' || \
    (echo && echo 'FAIL: Bad match to reference' && exit 1)

.PHONY: cgb_acid2
cgb_acid2: bin/gogo-gb tests/cgb-acid2/cgb-acid2.gbc
	@rm -f /tmp/gogo-gb-cgb_acid2.png
	@bin/gogo-gb run "tests/cgb-acid2/cgb-acid2.gbc" \
              --log=stderr \
              --debugger=interactive \
              --debugger-attach=false \
              --debugger-soft-break=screenshot \
              --screenshot-path=/tmp/gogo-gb-cgb_acid2.png \
              --headless
	@compare -metric AE "/tmp/gogo-gb-cgb_acid2.png" "tests/cgb-acid2/reference.png" null: 2>&1 && \
    echo && echo 'OK: Matched reference' || \
    (echo && echo 'FAIL: Bad match to reference' && exit 1)

.PHONY: mealybug_tests
mealybug_tests: bin/gogo-gb tests/mealybug-tearoom-tests/build/ppu/*.gb
	@MEALYBUG_TESTS=( \
    "m2_win_en_toggle.gb" \
    "m3_bgp_change_sprites.gb" \
    "m3_bgp_change.gb" \
    "m3_lcdc_bg_en_change.gb" \
    "m3_lcdc_bg_map_change.gb" \
    "m3_lcdc_obj_en_change_variant.gb" \
    "m3_lcdc_obj_en_change.gb" \
    "m3_lcdc_obj_size_change_scx.gb" \
    "m3_lcdc_obj_size_change.gb" \
    "m3_lcdc_tile_sel_change.gb" \
    "m3_lcdc_tile_sel_win_change.gb" \
    "m3_lcdc_win_en_change_multiple_wx.gb" \
    "m3_lcdc_win_en_change_multiple.gb" \
    "m3_lcdc_win_map_change.gb" \
    "m3_obp0_change.gb" \
    "m3_scx_high_5_bits.gb" \
    "m3_scx_low_3_bits.gb" \
    "m3_scy_change.gb" \
    "m3_window_timing_wx_0.gb" \
    "m3_window_timing.gb" \
    "m3_wx_4_change_sprites.gb" \
    "m3_wx_4_change.gb" \
    "m3_wx_5_change.gb" \
    "m3_wx_6_change.gb" \
  ); \
  for file in "$${MEALYBUG_TESTS[@]}"; do \
    test_name=$${file%*.gb}; \
    echo "=== Starting mealybug test $$test_name ==="; \
    bin/gogo-gb run "tests/mealybug-tearoom-tests/build/ppu/$$file" \
              --log=stderr \
              --serial-port=stdout \
              --debugger=interactive \
              --debugger-attach=false \
              --debugger-soft-break=screenshot \
              --screenshot-path="/tmp/gogo-gb-$$test_name.png" \
              --headless; \
    compare -metric AE -fuzz 5% "/tmp/gogo-gb-$$test_name.png" "tests/mealybug-tearoom-tests/expected/DMG-blob/$$test_name.png" NULL: 2>&1 && \
      echo && echo 'OK: Matched reference' || \
      (echo && echo 'FAIL: Bad match to reference' && exit 1); \
    echo "=== Finished mealybug test $$test_name ===" ; \
  done

.PHONY: mooneye_gb_tests
mooneye_gb_tests: bin/gogo-gb tests/mooneye-gb-test-suite/**/*.gb
	bin/gogo-gb run "tests/mooneye-gb-test-suite/$(MOONEYE_TEST).gb" \
              --log=stderr \
              --serial-port=stdout

tests/dmg-acid2/dmg-acid2.gb:
	mkdir -p tests/dmg-acid2
	curl -fSsL https://github.com/mattcurrie/dmg-acid2/releases/download/v1.0/dmg-acid2.gb > tests/dmg-acid2/dmg-acid2.gb

tests/cgb-acid2/cgb-acid2.gbc:
	mkdir -p tests/cgb-acid2
	curl -fSsL https://github.com/mattcurrie/cgb-acid2/releases/download/v1.1/cgb-acid2.gbc > tests/cgb-acid2/cgb-acid2.gbc

tests/gameboy-doctor/gameboy-doctor:
	git submodule init
	git submodule update

tests/gb-test-roms/cpu_instrs/individual/*.gb:
	git submodule init
	git submodule update

tests/gb-test-roms/mem_timing/individual/*.gb:
	git submodule init
	git submodule update

tests/mooneye-gb-test-suite/**/*.gb:
	mkdir -p tests/mooneye-gb-test-suite
	curl -fSsL https://gekkio.fi/files/mooneye-test-suite/$(MOONEYE_TEST_SUITE_VERION)/$(MOONEYE_TEST_SUITE_VERION).tar.xz > tests/mooneye-gb-test-suite/$(MOONEYE_TEST_SUITE_VERION).tar.xz
	tar -xf tests/mooneye-gb-test-suite/$(MOONEYE_TEST_SUITE_VERION).tar.xz -C tests/mooneye-gb-test-suite --strip-components=1

tests/mealybug-tearoom-tests/build/ppu/*.gb:
	git submodule init
	git submodule update --recursive
	mkdir -p tests/mealybug-tearoom-tests/build/ppu
	cd tests/mealybug-tearoom-tests && unzip mealybug-tearoom-tests.zip -d build/ppu
