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
)

type TimerClockSelector byte

const (
	TIMER_CLK_SEL_CPU_DIV_1024 TimerClockSelector = 0x00
	TIMER_CLK_SEL_CPU_DIV_16   TimerClockSelector = 0x01
	TIMER_CLK_SEL_CPU_DIV_64   TimerClockSelector = 0x02
	TIMER_CLK_SEL_CPU_DIV_256  TimerClockSelector = 0x03
)

const (
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
	freqSel    TimerClockSelector

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
	// TODO(GBC): Need to take into account CGB double-speed mode
	timer.divider += cycles

	if !timer.incCounter {
		return
	}

	timer.counterClk += uint(cycles)
	requestInt := false

	for timer.counterClk >= timer.FreqDivider() {
		timer.counterClk -= timer.FreqDivider()

		if timer.counter == 0xFF {
			timer.counter = timer.modulo
			requestInt = true
		} else {
			timer.counter += 1
		}
	}

	if requestInt {
		ic.RequestTimer()
	}
}

func (timer *Timer) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	switch addr {
	case REG_TIMER_DIV:
		return mem.ReadReplace(timer.divider)
	case REG_TIMER_TIMA:
		return mem.ReadReplace(timer.counter)
	case REG_TIMER_TMA:
		return mem.ReadReplace(timer.modulo)
	case REG_TIMER_TAC:
		tac := byte(timer.freqSel)

		if timer.incCounter {
			tac |= TIMER_CLK_EN_MASK
		}

		return mem.ReadReplace(tac)
	default:
		return mem.ReadPassthrough()
	}
}

func (timer *Timer) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	switch addr {
	case REG_TIMER_DIV:
		timer.divider = 0
	case REG_TIMER_TIMA:
		timer.counter = value
	case REG_TIMER_TMA:
		timer.modulo = value
	case REG_TIMER_TAC:
		enableCounter := (value & TIMER_CLK_EN_MASK) == TIMER_CLK_EN_MASK

		if enableCounter && !timer.incCounter {
			timer.incCounter = true
			timer.freqSel = TimerClockSelector(value & TIMER_CLK_SEL_MASK)
		} else {
			timer.incCounter = enableCounter
		}
	default:
		panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for timer", value, addr))
	}

	return mem.WriteBlock()
}
