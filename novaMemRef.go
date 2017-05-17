// novaMemRef.go
package main

import (
	"log"
)

func novaMemRef(cpuPtr *Cpu, iPtr *DecodedInstr) bool {

	var shifter dg_word
	var effAddr dg_phys_addr

	switch iPtr.mnemonic {
	case "LDA":
		effAddr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		shifter = memReadWord(effAddr)
		cpuPtr.ac[iPtr.acd] = 0x0000ffff & dg_dword(shifter)
	case "STA":
		shifter = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		effAddr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		memWriteWord(effAddr, shifter)
	default:
		log.Printf("ERROR: NOVA_MEMREF instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}
	cpuPtr.pc++
	return true
}
