// eagleMemRef.go
package main

import (
	"log"
)

func eagleMemRef(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	var (
		addr dg_phys_addr
		wd   dg_word
		dwd  dg_dword
	)

	switch iPtr.mnemonic {

	case "XLDB":
		cpuPtr.ac[iPtr.acd] = dg_dword(memReadByte(resolve16bitEagleAddr(cpuPtr, ' ', iPtr.mode, iPtr.disp), iPtr.loHiBit)) & 0x00ff

	case "XNLDA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		wd = memReadWord(addr)
		cpuPtr.ac[iPtr.acd] = sexWordToDWord(wd) // FIXME check this...

	case "XNSTA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		memWriteWord(addr, wd)
	case "XWLDA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		dwd = memReadDWord(addr)
		cpuPtr.ac[iPtr.acd] = dwd

	case "XWSTA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		dwd = cpuPtr.ac[iPtr.acd]
		memWriteDWord(addr, dwd)

	default:
		log.Printf("ERROR: EAGLE_MEMREF instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
