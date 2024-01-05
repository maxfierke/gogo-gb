package devices

import (
	"fmt"
	"os"

	"github.com/maxfierke/gogo-gb/mem"
)

const (
	REG_SERIAL_SB = 0xFF01
	REG_SERIAL_SC = 0xFF02

	SC_TRANSFER_EN = 0x80
	SC_CLK_SPD     = 0x2
	SC_CLK_SEL     = 0x1

	SC_CLK_EXT = 0x0
	SC_CLK_INT = 0x1
)

type SerialCtrl struct {
	transferEnabled bool
	clockSpeedDbl   bool
	clockInternal   bool
}

func (sc *SerialCtrl) Read() byte {
	value := byte(0x0)

	if sc.transferEnabled {
		value &= SC_TRANSFER_EN
	}

	if sc.clockSpeedDbl {
		value &= SC_CLK_SPD
	}

	if sc.clockInternal {
		value &= SC_CLK_INT
	}

	return value
}

func (sc *SerialCtrl) Write(value byte) {
	sc.transferEnabled = (value & SC_TRANSFER_EN) != 0
	sc.clockSpeedDbl = (value & SC_CLK_SPD) != 0
	sc.clockInternal = (value & SC_CLK_SEL) == SC_CLK_INT
}

func (sc *SerialCtrl) IsTransferEnabled() bool {
	return sc.transferEnabled
}

func (sc *SerialCtrl) SetTransferEnabled(enabled bool) {
	sc.transferEnabled = enabled
}

func (sc *SerialCtrl) IsClockSpeedDbl() bool {
	return sc.clockSpeedDbl
}

func (sc *SerialCtrl) SetClockSpeedDbl(enabled bool) {
	sc.clockSpeedDbl = enabled
}

func (sc *SerialCtrl) IsClockInternal() bool {
	return sc.clockInternal
}

func (sc *SerialCtrl) SetClockInternal(enabled bool) {
	sc.clockInternal = enabled
}

type SerialPort struct {
	clk  uint
	ctrl SerialCtrl
	recv byte
	buf  byte
}

func NewSerialPort() *SerialPort {
	return &SerialPort{}
}

func (sp *SerialPort) Step(cycles uint, ic *InterruptController) {
	if !sp.ctrl.IsTransferEnabled() {
		return
	}

	if sp.ctrl.IsClockInternal() {
		if sp.clk < cycles {
			sp.buf = sp.recv
			sp.ctrl.SetTransferEnabled(false)
			ic.RequestSerial()
		} else {
			sp.clk -= cycles
		}
	} else {
		// TODO: Implement external clock
	}
}

func (sp *SerialPort) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr == REG_SERIAL_SB {
		return mem.ReadReplace(sp.buf)
	} else if addr == REG_SERIAL_SC {
		return mem.ReadReplace(sp.ctrl.Read())
	}

	return mem.ReadPassthrough()
}

func (sp *SerialPort) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_SERIAL_SB {
		sp.buf = value
		return mem.WriteBlock()
	} else if addr == REG_SERIAL_SC {
		sp.ctrl.Write(value)

		if sp.ctrl.IsTransferEnabled() && sp.ctrl.IsClockInternal() {
			// TODO: derive this somehow and factor in GBC speeds when relevant
			sp.clk = 8192

			// TODO: Write this to some IO writer
			os.Stderr.Write([]byte{sp.buf})

			// TODO: Read this from somewhere
			sp.recv = 0xFF
		}

		return mem.WriteBlock()
	}

	panic(fmt.Sprintf("Attempting to write 0x%02X @ 0x%04X, which is out-of-bounds for serial port", value, addr))
}