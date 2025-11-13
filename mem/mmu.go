package mem

type MMU struct {
	ram           []byte
	handleCounter uint
	handles       map[MemHandlerHandle]MemRegion
	handlers      map[uint16][]MMUHandler
}

type MMUHandler struct {
	handle  MemHandlerHandle
	handler MemHandler
}

type MemRead struct {
	replacement byte
	passthrough bool
}

func ReadReplace(replacement byte) MemRead {
	return MemRead{
		replacement: replacement,
		passthrough: false,
	}
}

func ReadPassthrough() MemRead {
	return MemRead{
		replacement: 0x00,
		passthrough: true,
	}
}

type MemWrite struct {
	replacement byte
	passthrough bool
	blocked     bool
}

func WriteReplace(value byte) MemWrite {
	return MemWrite{replacement: value, passthrough: false, blocked: false}
}

func WritePassthrough() MemWrite {
	return MemWrite{replacement: 0x00, passthrough: true, blocked: false}
}

func WriteBlock() MemWrite {
	return MemWrite{replacement: 0x00, passthrough: false, blocked: true}
}

type MemHandler interface {
	OnRead(mmu *MMU, addr uint16) MemRead
	OnWrite(mmu *MMU, addr uint16, value byte) MemWrite
}

type MemHandlerHandle struct {
	val uint
}

type MemRegion struct {
	Start uint16
	End   uint16
}

func (region *MemRegion) Contains(addr uint16, exclusive bool) bool {
	if exclusive {
		return region.Start <= addr && addr < region.End
	} else {
		return region.Start <= addr && addr <= region.End
	}
}

type MemBus interface {
	Read8(addr uint16) byte
	Write8(addr uint16, value byte)
	Read16(addr uint16) uint16
	Write16(addr uint16, value uint16)
}

type EchoRegion struct{}

func NewEchoRegion() *EchoRegion {
	return &EchoRegion{}
}

func (umr *EchoRegion) OnRead(mmu *MMU, addr uint16) MemRead {
	// Echo mirrors 0xC000
	return ReadReplace(mmu.Read8(addr - 0x2000))
}

func (umr *EchoRegion) OnWrite(mmu *MMU, addr uint16, value byte) MemWrite {
	return WriteBlock()
}

type UnmappedRegion struct{}

func NewUnmappedRegion() *UnmappedRegion {
	return &UnmappedRegion{}
}

func (umr *UnmappedRegion) OnRead(mmu *MMU, addr uint16) MemRead {
	return ReadReplace(0x00)
}

func (umr *UnmappedRegion) OnWrite(mmu *MMU, addr uint16, value byte) MemWrite {
	return WriteBlock()
}

func NewMMU(ram []byte) *MMU {
	return &MMU{
		ram:           ram,
		handleCounter: 0,
		handles:       map[MemHandlerHandle]MemRegion{},
		handlers:      map[uint16][]MMUHandler{},
	}
}

func (mmu *MMU) AddHandler(region MemRegion, handler MemHandler) MemHandlerHandle {
	handle := mmu.nextHandle()

	mmu.handles[handle] = region

	for addr := uint(region.Start); addr <= uint(region.End); addr++ {
		addr16 := uint16(addr)
		val, exist := mmu.handlers[addr16]

		if !exist {
			val = make([]MMUHandler, 0)
			mmu.handlers[addr16] = val
		}

		mmu.handlers[addr16] = append(val, MMUHandler{handle: handle, handler: handler})
	}

	return handle
}

func (mmu *MMU) RemoveHandler(handle MemHandlerHandle) {
	region, exist := mmu.handles[handle]

	if !exist {
		return
	}

	delete(mmu.handles, handle)

	for addr := uint(region.Start); addr <= uint(region.End); addr++ {
		addr16 := uint16(addr)
		val, exist := mmu.handlers[addr16]

		if exist {
			newVal := make([]MMUHandler, len(val)-1)

			for i := range val {
				if val[i].handle != handle {
					newVal = append(newVal, val[i])
				}
			}

			mmu.handlers[addr16] = newVal
		}
	}
}

func (mmu *MMU) Read8(addr uint16) byte {
	addrHandlers, handlersExist := mmu.handlers[addr]

	if handlersExist {
		for i := range addrHandlers {
			handler := addrHandlers[i]

			memread := handler.handler.OnRead(mmu, addr)

			if !memread.passthrough {
				return memread.replacement
			}
		}
	}

	return mmu.ram[addr]
}

func (mmu *MMU) Write8(addr uint16, value byte) {
	addrHandlers, handlersExist := mmu.handlers[addr]

	if handlersExist {
		for i := range addrHandlers {
			handler := addrHandlers[i]

			memwrite := handler.handler.OnWrite(mmu, addr, value)

			if memwrite.blocked {
				return
			}

			if !memwrite.passthrough {
				mmu.ram[addr] = memwrite.replacement

				return
			}
		}
	}

	mmu.ram[addr] = value
}

func (mmu *MMU) Read16(addr uint16) uint16 {
	low := mmu.Read8(addr)
	high := mmu.Read8(addr + 1)

	return uint16(high)<<8 | uint16(low)
}

func (mmu *MMU) Write16(addr uint16, value uint16) {
	mmu.Write8(addr, uint8(value))
	mmu.Write8(addr+1, uint8(value>>8))
}

func (mmu *MMU) nextHandle() MemHandlerHandle {
	handle := MemHandlerHandle{
		val: mmu.handleCounter,
	}

	mmu.handleCounter += 1

	return handle
}
