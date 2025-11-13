package devices

import (
	"testing"

	"github.com/maxfierke/gogo-gb/mem"
	"github.com/stretchr/testify/assert"
)

var NULL_MMU = mem.NewMMU([]byte{})

func TestTimerDivider(t *testing.T) {
	assert := assert.New(t)

	timer := NewTimer()
	assert.Equal(uint8(0), timer.divider)

	timer.Step(12, &InterruptController{})
	assert.Equal(uint8(12), timer.divider)

	op := timer.OnWrite(NULL_MMU, REG_TIMER_DIV, 0xFF)
	assert.Equal(mem.WriteBlock(), op)
	assert.Equal(uint8(0), timer.divider)
}

func TestTimerTIMA(t *testing.T) {
	assert := assert.New(t)

	// TIMER_CLK_SEL_CPU_DIV_256
	timer := NewTimer()
	ic := &InterruptController{}

	op := timer.OnWrite(NULL_MMU, REG_TIMER_TAC, 0x07)
	assert.Equal(mem.WriteBlock(), op)
	assert.True(timer.incCounter)
	assert.Equal(TimerClockSelector(TIMER_CLK_SEL_CPU_DIV_256), timer.freqSel)

	for range 255 {
		timer.Step(1, ic)
	}
	assert.Equal(uint8(0), timer.counter)

	// Should increase once we hit frequency divider
	timer.Step(1, ic)
	assert.Equal(uint8(1), timer.counter)

	// TIMER_CLK_SEL_CPU_DIV_1024
	timer = NewTimer()
	ic = &InterruptController{}

	op = timer.OnWrite(NULL_MMU, REG_TIMER_TAC, 0x04)
	assert.Equal(mem.WriteBlock(), op)
	assert.True(timer.incCounter)
	assert.Equal(TimerClockSelector(TIMER_CLK_SEL_CPU_DIV_1024), timer.freqSel)

	for range 4096 {
		timer.Step(1, ic)
	}
	assert.Equal(uint8(4), timer.counter)

	// TIMER_CLK_SEL_CPU_DIV_16
	timer = NewTimer()
	ic = &InterruptController{}

	op = timer.OnWrite(NULL_MMU, REG_TIMER_TAC, 0x05)
	assert.Equal(mem.WriteBlock(), op)
	assert.True(timer.incCounter)
	assert.Equal(TimerClockSelector(TIMER_CLK_SEL_CPU_DIV_16), timer.freqSel)

	timer.Step(64, ic)
	assert.Equal(uint8(4), timer.counter)

	// TIMER_CLK_SEL_CPU_DIV_64
	timer = NewTimer()
	ic = &InterruptController{}

	op = timer.OnWrite(NULL_MMU, REG_TIMER_TAC, 0x06)
	assert.Equal(mem.WriteBlock(), op)
	assert.True(timer.incCounter)
	assert.Equal(TimerClockSelector(TIMER_CLK_SEL_CPU_DIV_64), timer.freqSel)

	for range 256 {
		timer.Step(1, ic)
	}
	assert.Equal(uint8(4), timer.counter)
}

func TestTimerInterrupt(t *testing.T) {
	assert := assert.New(t)

	timer := NewTimer()
	ic := &InterruptController{}

	// Enable all Interrupts
	ic.OnWrite(NULL_MMU, REG_IE, 0xFF)

	op := timer.OnWrite(NULL_MMU, REG_TIMER_TAC, 0x05)
	assert.Equal(mem.WriteBlock(), op)
	assert.True(timer.incCounter)
	assert.Equal(TimerClockSelector(TIMER_CLK_SEL_CPU_DIV_16), timer.freqSel)

	for range 255 {
		timer.Step(16, ic)
	}
	assert.Equal(uint8(0xFF), timer.counter)
	assert.Equal(INT_NONE, ic.NextRequest())

	timer.Step(16, ic)
	assert.Equal(uint8(0x0), timer.counter)
	assert.Equal(INT_TIMER, ic.NextRequest())
}
