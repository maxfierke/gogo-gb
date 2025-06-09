package mem

func ReadBankAddr(memory []byte, banksRegion MemRegion, bankSize uint16, currentBank uint16, addr uint16) byte {
	bankBaseAddr := uint(currentBank) * uint(bankSize)
	bankSlotAddr := uint(addr) - uint(banksRegion.Start)
	memoryAddrMask := uint(len(memory) - 1)
	memoryAddr := (bankBaseAddr + bankSlotAddr) & memoryAddrMask
	return memory[memoryAddr]
}

func WriteBankAddr(memory []byte, banksRegion MemRegion, bankSize uint16, currentBank uint16, addr uint16, value byte) {
	bankBaseAddr := uint(currentBank) * uint(bankSize)
	bankSlotAddr := uint(addr) - uint(banksRegion.Start)
	memoryAddrMask := uint(len(memory) - 1)
	memoryAddr := (bankBaseAddr + bankSlotAddr) & memoryAddrMask
	memory[memoryAddr] = value
}
