# gogo-gb
a gameboy emulator for funsies

Current status: Games are playable at 60 FPS. CPU cycle accuracy, but some timings not quite right. No audio.

## TODO

- [X] Pass all of Blargg's `cpu_instrs` ROMs (verified via `gameboy-doctor`)
- [X] Implement MBC1
- [X] Implement MBC5
- [X] Implement MBC3 (w/o RTC)
- [X] Implement serial port (w/ option to log to console)
- [X] Implement timer
- [X] Pass Blargg's `instr_timing.gb` ROM (manually verified)
- [X] Implement a basic interactive debugger
- [X] Implement LCD
- [X] Implement PPU, VRAM, OAM, etc.
- [X] Implement Joypad
- [X] Implement SRAM save & restore
- [X] Achieve ~60 FPS
- [X] Implement `watch` in debugger for memory & register changes
- [X] Implement RTC for MBC3
- [ ] Implement PPU registers debugging
- [ ] Implement GBC
- [ ] Pass `dmg-acid2` test ROM
- [ ] Implement Sound/APU

## Later/Maybe Never?

Just being realistic about my likelihood of getting to these:

- [ ] Pass Blargg's `mem_timing` ROMs (manually verified)
- [ ] Pass Blargg's `mem_timing-2` ROMs (manually verified)
- [ ] FIFO-based rendering PPU (currently scanline)
- [ ] Implement emulation for every known DMG bug
- [ ] Implement SGB mode
- [ ] Implement MBC2
- [ ] Implement MBC6
- [ ] Implement MBC7
- [ ] Implement MBC1M, MMM01, other multicarts, or Hudson carts
- [ ] Implement (any) accessories

## Inspiration Material

* [DMG-01](https://rylev.github.io/DMG-01/public/book/introduction.html)
* [Gameboy Doctor](https://github.com/robert/gameboy-doctor)
* [Let's Write a Gameboy Emulator in Python](https://www.inspiredpython.com/course/game-boy-emulator/let-s-write-a-game-boy-emulator-in-python)
* [Pandocs](https://gbdev.io/pandocs/About.html)
* [Writing a Gameboy Emulator in Rust](https://yushiomote.org/posts/gameboy-emu)
