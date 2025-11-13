package mem

import "fmt"

const (
	REG_WRAM_SVBK = 0xFF70

	REG_WRAM_SVBK_SEL_MASK = 0x7

	WRAM_BANK_SIZE uint16 = 0x1000
	WRAM_SIZE      uint16 = 0x8000
)

var (
	WRAM_BANK_00 = MemRegion{Start: 0xC000, End: 0xCFFF}
	WRAM_BANKS   = MemRegion{Start: 0xD000, End: 0xDFFF}
)

type WRAM struct {
	curBank uint8
	wram    []byte
}

var _ MemHandler = (*WRAM)(nil)

func NewWRAM() *WRAM {
	return &WRAM{
		wram: make([]byte, WRAM_SIZE),
	}
}

func (w *WRAM) OnRead(mmu *MMU, addr uint16) MemRead {
	if addr == REG_WRAM_SVBK {
		return ReadReplace(max(w.curBank, 1) & REG_WRAM_SVBK_SEL_MASK)
	}

	if WRAM_BANK_00.Contains(addr, false) {
		bankByte := ReadBankAddr(
			w.wram,
			WRAM_BANK_00,
			WRAM_BANK_SIZE,
			0,
			addr,
		)

		return ReadReplace(bankByte)
	} else if WRAM_BANKS.Contains(addr, false) {
		bankByte := ReadBankAddr(
			w.wram,
			WRAM_BANKS,
			WRAM_BANK_SIZE,
			max(uint16(w.curBank), 1),
			addr,
		)

		return ReadReplace(bankByte)
	}

	return ReadPassthrough()
}

func (w *WRAM) OnWrite(mmu *MMU, addr uint16, value byte) MemWrite {
	if addr == REG_WRAM_SVBK {
		w.curBank = value & REG_WRAM_SVBK_SEL_MASK

		return WriteBlock()
	}

	if WRAM_BANK_00.Contains(addr, false) {
		WriteBankAddr(
			w.wram,
			WRAM_BANK_00,
			WRAM_BANK_SIZE,
			0,
			addr,
			value,
		)

		return WriteBlock()
	}

	if WRAM_BANKS.Contains(addr, false) {
		WriteBankAddr(
			w.wram,
			WRAM_BANKS,
			WRAM_BANK_SIZE,
			max(uint16(w.curBank), 1),
			addr,
			value,
		)

		return WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for WRAM", value, addr))
}
