package devices

import (
	"sync"

	"github.com/maxfierke/gogo-gb/bits"
	"github.com/maxfierke/gogo-gb/mem"
)

const (
	REG_JOYP = 0xFF00
)

const (
	REG_JOYP_BIT_A_RIGHT = iota
	REG_JOYP_BIT_B_LEFT
	REG_JOYP_BIT_SELECT_UP
	REG_JOYP_BIT_START_DOWN
	REG_JOYP_BIT_DPAD_SEL
	REG_JOYP_BIT_BUTTONS_SEL
)

type JoypadInputs struct {
	A      bool
	B      bool
	Up     bool
	Down   bool
	Left   bool
	Right  bool
	Start  bool
	Select bool
}

func (ji JoypadInputs) AnyPressed() bool {
	return ji.A || ji.B || ji.Up || ji.Down || ji.Left || ji.Right || ji.Start || ji.Select
}

type Joypad struct {
	readButtons  bool
	readDPad     bool
	inputState   JoypadInputs
	inputStateMu sync.Mutex

	ic *InterruptController
}

func NewJoypad(ic *InterruptController) *Joypad {
	return &Joypad{
		ic: ic,
	}
}

func (j *Joypad) ReceiveInputs(inputs JoypadInputs) {
	j.inputStateMu.Lock()
	defer j.inputStateMu.Unlock()

	if inputs.AnyPressed() {
		j.ic.RequestJoypad()
	}

	j.inputState = inputs
}

func (j *Joypad) OnRead(mmu *mem.MMU, addr uint16) mem.MemRead {
	if addr == REG_JOYP {
		j.inputStateMu.Lock()
		defer j.inputStateMu.Unlock()

		var (
			readButtons uint8
			readDPad    uint8
			startDown   uint8
			selectUp    uint8
			bLeft       uint8
			aRight      uint8
		)

		if j.readButtons {
			readButtons = 1 << REG_JOYP_BIT_BUTTONS_SEL
		}

		if j.readDPad {
			readDPad = 1 << REG_JOYP_BIT_DPAD_SEL
		}

		if j.inputState.Start && j.readButtons {
			startDown = 1 << REG_JOYP_BIT_START_DOWN
		}

		if j.inputState.Down && j.readDPad {
			startDown |= 1 << REG_JOYP_BIT_START_DOWN
		}

		if j.inputState.Select && j.readButtons {
			selectUp = 1 << REG_JOYP_BIT_SELECT_UP
		}

		if j.inputState.Up && j.readDPad {
			selectUp |= 1 << REG_JOYP_BIT_SELECT_UP
		}

		if j.inputState.B && j.readButtons {
			bLeft = 1 << REG_JOYP_BIT_B_LEFT
		}

		if j.inputState.Left && j.readDPad {
			bLeft |= 1 << REG_JOYP_BIT_B_LEFT
		}

		if j.inputState.A && j.readButtons {
			aRight = 1 << REG_JOYP_BIT_A_RIGHT
		}

		if j.inputState.Right && j.readDPad {
			aRight |= 1 << REG_JOYP_BIT_A_RIGHT
		}

		readByte := (readButtons | readDPad | startDown | selectUp | bLeft | aRight) ^ 0xFF

		return mem.ReadReplace(readByte)
	}

	return mem.ReadPassthrough()
}

func (j *Joypad) OnWrite(mmu *mem.MMU, addr uint16, value byte) mem.MemWrite {
	if addr == REG_JOYP {
		j.readButtons = bits.Read(value, REG_JOYP_BIT_BUTTONS_SEL) == 0
		j.readDPad = bits.Read(value, REG_JOYP_BIT_DPAD_SEL) == 0
	}

	return mem.WriteBlock()
}
