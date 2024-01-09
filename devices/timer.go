package devices

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/mem"
)

const (
	REG_TIMER_DIV  = 0xFF04
	REG_TIMER_TIMA = 0xFF05
	REG_TIMER_TMA  = 0xFF06
	REG_TIMER_TAC  = 0xFF07

	TIMER_DIV_CLOCK_CYCLES = 256

	TIMER_CLK_EN_MASK  = 1 << 2
	TIMER_CLK_SEL_MASK = 0x3

	TIMER_CLK_SEL_CPU_DIV_1024 = 0x00
	TIMER_CLK_SEL_CPU_DIV_16   = 0x01
	TIMER_CLK_SEL_CPU_DIV_64   = 0x02
	TIMER_CLK_SEL_CPU_DIV_256  = 0x03

	TIMER_TIMA_CLK_1024 = 1024
	TIMER_TIMA_CLK_16   = 16
	TIMER_TIMA_CLK_64   = 64
	TIMER_TIMA_CLK_256  = 256
)

type Timer struct {
	divider    uint8
	counter    uint8
	modulo     uint8
	incCounter bool
	freqSel    byte

	dividerClk uint
	counterClk uint
}

func NewTimer() *Timer {
	return &Timer{}
}

func (timer *Timer) FreqDivider() uint {
	switch timer.freqSel {
	case TIMER_CLK_SEL_CPU_DIV_1024:
		return TIMER_TIMA_CLK_1024
	case TIMER_CLK_SEL_CPU_DIV_16:
		return TIMER_TIMA_CLK_16
	case TIMER_CLK_SEL_CPU_DIV_64:
		return TIMER_TIMA_CLK_64
	case TIMER_CLK_SEL_CPU_DIV_256:
		return TIMER_TIMA_CLK_256
	default:
		panic(fmt.Sprintf("Unexpected value for frequency divider: %v", timer.freqSel))
	}
}

func (timer *Timer) Step(cycles uint8, ic *InterruptController) {
	if timer.dividerClk < uint(cycles) {
		timer.divider += 1
		remainingCycles := uint(cycles) - timer.dividerClk
		timer.dividerClk = TIMER_DIV_CLOCK_CYCLES - remainingCycles
	} else {
		timer.dividerClk -= uint(cycles)
	}

	if !timer.incCounter {
		return
	}

	if timer.counterClk < uint(cycles) {
		remainingCycles := uint(cycles) - timer.counterClk

		if timer.counter == 0xFF {
			timer.counter = timer.modulo
			ic.RequestTimer()
		} else {
			timer.counter += 1
		}

		timer.counterClk = timer.FreqDivider() - remainingCycles
	} else {
		timer.counterClk -= uint(cycles)
	}
}

func (timer *Timer) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr == REG_TIMER_DIV {
		return mem.ReadReplace(timer.divider)
	} else if addr == REG_TIMER_TIMA {
		return mem.ReadReplace(timer.counter)
	} else if addr == REG_TIMER_TMA {
		return mem.ReadReplace(timer.modulo)
	} else if addr == REG_TIMER_TAC {
		tac := timer.freqSel

		if timer.incCounter {
			tac |= TIMER_CLK_EN_MASK
		}

		return mem.ReadReplace(tac)
	}

	return mem.ReadPassthrough()
}

func (timer *Timer) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_TIMER_DIV {
		timer.divider = 0
	} else if addr == REG_TIMER_TIMA {
		timer.counter = value
	} else if addr == REG_TIMER_TMA {
		timer.modulo = value
	} else if addr == REG_TIMER_TAC {

		enableCounter := (value & TIMER_CLK_EN_MASK) == TIMER_CLK_EN_MASK

		if enableCounter && !timer.incCounter {
			timer.incCounter = true
			timer.freqSel = (value & TIMER_CLK_SEL_MASK)
		} else {
			timer.incCounter = enableCounter
		}
	} else {
		panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for timer", value, addr))
	}

	return mem.WriteBlock()
}
