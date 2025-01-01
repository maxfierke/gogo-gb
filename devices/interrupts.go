package devices

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/mem"
)

const (
	REG_IE = 0xFFFF
	REG_IF = 0xFF0F
)

type IRQ byte

const (
	INT_NONE   IRQ = 0x00
	INT_VBLANK IRQ = 0x40
	INT_STAT   IRQ = 0x48
	INT_TIMER  IRQ = 0x50
	INT_SERIAL IRQ = 0x58
	INT_JOYPAD IRQ = 0x60
)

const (
	vblankFlagBit = 0
	lcdFlagBit    = 1
	timerFlagBit  = 2
	serialFlagBit = 3
	joypadFlagBit = 4
)

type InterruptLine struct {
	vblank bool
	lcd    bool
	timer  bool
	serial bool
	joypad bool
}

func (il *InterruptLine) Read() uint8 {
	var (
		vblank uint8
		lcd    uint8
		timer  uint8
		serial uint8
		joypad uint8
	)

	if il.vblank {
		vblank = 1 << vblankFlagBit
	}

	if il.lcd {
		lcd = 1 << lcdFlagBit
	}

	if il.timer {
		timer = 1 << timerFlagBit
	}

	if il.serial {
		serial = 1 << serialFlagBit
	}

	if il.joypad {
		joypad = 1 << joypadFlagBit
	}

	return (vblank | lcd | timer | serial | joypad)
}

func (il *InterruptLine) Write(value uint8) {
	il.vblank = ((value >> vblankFlagBit) & 0b1) != 0
	il.lcd = ((value >> lcdFlagBit) & 0b1) != 0
	il.timer = ((value >> timerFlagBit) & 0b1) != 0
	il.serial = ((value >> serialFlagBit) & 0b1) != 0
	il.joypad = ((value >> joypadFlagBit) & 0b1) != 0
}

type InterruptController struct {
	enabled   InterruptLine
	requested InterruptLine
}

func NewInterruptController() *InterruptController {
	return &InterruptController{
		enabled:   InterruptLine{},
		requested: InterruptLine{},
	}
}

func (ic *InterruptController) ConsumeRequest() IRQ {
	nextReq := ic.NextRequest()

	// https://gbdev.io/pandocs/Interrupts.html#interrupt-priorities
	if nextReq == INT_VBLANK {
		ic.requested.vblank = false
	}

	if nextReq == INT_STAT {
		ic.requested.lcd = false
	}

	if nextReq == INT_JOYPAD {
		ic.requested.joypad = false
	}

	if nextReq == INT_SERIAL {
		ic.requested.serial = false
	}

	if nextReq == INT_TIMER {
		ic.requested.timer = false
	}

	return nextReq
}

func (ic *InterruptController) NextRequest() IRQ {
	if ic.enabled.vblank && ic.requested.vblank {
		return INT_VBLANK
	}

	if ic.enabled.lcd && ic.requested.lcd {
		return INT_STAT
	}

	if ic.enabled.timer && ic.requested.timer {
		return INT_TIMER
	}

	if ic.enabled.serial && ic.requested.serial {
		return INT_SERIAL
	}

	if ic.enabled.joypad && ic.requested.joypad {
		return INT_JOYPAD
	}

	return INT_NONE
}

func (ic *InterruptController) Reset() {
	ic.enabled.Write(0x00)
	ic.requested.Write(0x00)
}

func (ic *InterruptController) RequestLCD() {
	ic.requested.lcd = true
}

func (ic *InterruptController) RequestJoypad() {
	ic.requested.joypad = true
}

func (ic *InterruptController) RequestSerial() {
	ic.requested.serial = true
}

func (ic *InterruptController) RequestTimer() {
	ic.requested.timer = true
}

func (ic *InterruptController) RequestVBlank() {
	ic.requested.vblank = true
}

func (ic *InterruptController) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr == REG_IE {
		return mem.ReadReplace(ic.enabled.Read())
	} else if addr == REG_IF {
		return mem.ReadReplace(ic.requested.Read())
	}

	return mem.ReadPassthrough()
}

func (ic *InterruptController) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_IE {
		ic.enabled.Write(value)
		return mem.WriteBlock()
	} else if addr == REG_IF {
		ic.requested.Write(value)
		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for interrupts", value, addr))
}
