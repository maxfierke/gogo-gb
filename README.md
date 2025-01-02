# gogo-gb
a gameboy emulator for funsies

Current status: Games are playable, but slow. Graphics are buggy. No audio.

## TODO

- [X] Pass all of Blargg's `cpu_instrs` ROMs via `gameboy-doctor` (expect `02-interrupts.gb`, which isn't verifyable via `gameboy-doctor`)
- [X] Implement serial port (w/ option to log to console)
- [X] Implement timer
- [X] Pass Blargg's `cpu_instrs`/`02-interrupts.gb` ROM (manually verified)
- [X] Pass Blargg's `instr_timing.gb` ROM (manually verified)
- [X] Implement a basic interactive debugger
- [ ] Pass Blargg's `mem_timing.gb` ROM (manually verified)
- [X] Implement LCD
- [X] Implement PPU, VRAM, OAM, etc.
- [ ] Pass all of Blargg's `mem_timing-2` ROMs (manually verified)
- [X] Implement Joypad
- [ ] Implement RTC for MBC3
- [X] Implement SRAM save & restore
- [ ] Pass `dmg-acid2` test ROM
- [ ] Implement Sound/APU
- [ ] Implement GBC

## Maybe Never?

Just being realistic about my likelihood of getting to these:

- [ ] FIFO-based rendering PPU (currently scanline)
- [ ] Implement emulation for every known DMG bug
- [ ] Implement SGB mode
- [ ] Implement MBC2
- [ ] Implement MBC6
- [ ] Implement MBC7
- [ ] Implement any multicarts or Hudson carts
- [ ] Implement (any) accessories

## Inspiration Material

* [DMG-01](https://rylev.github.io/DMG-01/public/book/introduction.html)
* [Gameboy Doctor](https://github.com/robert/gameboy-doctor)
* [Let's Write a Gameboy Emulator in Python](https://www.inspiredpython.com/course/game-boy-emulator/let-s-write-a-game-boy-emulator-in-python)
* [Pandocs](https://gbdev.io/pandocs/About.html)
* [Writing a Gameboy Emulator in Rust](https://yushiomote.org/posts/gameboy-emu)
