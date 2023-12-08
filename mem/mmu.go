package mem

const RAMSize = 0xFFFF

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

type MemHandler interface {
	OnRead(mmu *MMU, addr uint16) (replacement byte, passthrough bool)
	OnWrite(mmu *MMU, addr uint16, value byte) (replacement byte, passthrough bool, block bool)
}

type MemHandlerHandle struct {
	val uint
}

type MemRegion struct {
	Start uint16
	End   uint16
}

type MemAddressable interface {
	Read8(addr uint16) byte
	Write8(addr uint16, value byte)
	Read16(addr uint16) uint16
	Write16(addr uint16, value uint16)
}

func NewMMU() *MMU {
	return &MMU{
		ram:           make([]byte, RAMSize),
		handleCounter: 0,
		handles:       map[MemHandlerHandle]MemRegion{},
		handlers:      map[uint16][]MMUHandler{},
	}
}

func (mmu *MMU) AddHandler(region MemRegion, handler MemHandler) MemHandlerHandle {
	handle := mmu.nextHandle()

	mmu.handles[handle] = region

	for i := region.Start; i <= region.End; i++ {
		val, exist := mmu.handlers[i]

		if !exist {
			val = make([]MMUHandler, 0)
			mmu.handlers[i] = val
		}

		mmu.handlers[i] = append(val, MMUHandler{handle: handle, handler: handler})
	}

	return handle
}

func (mmu *MMU) RemoveHandler(handle MemHandlerHandle) {
	region, exist := mmu.handles[handle]

	if !exist {
		return
	}

	delete(mmu.handles, handle)

	for addr := region.Start; addr <= region.End; addr++ {
		val, exist := mmu.handlers[addr]

		if exist {
			newVal := make([]MMUHandler, len(val)-1)

			for i := 0; i < len(val); i++ {
				if val[i].handle != handle {
					newVal = append(newVal, val[i])
				}
			}

			mmu.handlers[addr] = newVal
		}
	}
}

func (mmu *MMU) Read8(addr uint16) byte {
	addrHandlers, handlersExist := mmu.handlers[addr]

	if handlersExist {
		for i := 0; i < len(addrHandlers); i++ {
			handler := addrHandlers[i]

			replacement, passthrough := handler.handler.OnRead(mmu, addr)

			if !passthrough {
				return replacement
			}
		}
	}

	// TODO: Handle Echo RAM ?
	// 			 What reads from Echo RAM?
	return mmu.ram[addr]
}

func (mmu *MMU) Write8(addr uint16, value byte) {
	addrHandlers, handlersExist := mmu.handlers[addr]

	if handlersExist {
		for i := 0; i < len(addrHandlers); i++ {
			handler := addrHandlers[i]

			replacement, passthrough, block := handler.handler.OnWrite(mmu, addr, value)

			if block {
				return
			}

			if !passthrough {
				mmu.ram[addr] = replacement
				return
			}
		}
	}

	// TODO: Handle Echo RAM
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
