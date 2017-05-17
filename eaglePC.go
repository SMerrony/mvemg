// eaglePC.go
package main

import (
	"log"
)

func eaglePC(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	//var addr dg_phys_addr
	//var wd dg_word

	switch iPtr.mnemonic {

	case "WBR":
		cpuPtr.pc = dg_phys_addr(iPtr.disp)

	case "XJMP":
		cpuPtr.pc = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)

	case "XJSR":
		cpuPtr.ac[3] = dg_dword(cpuPtr.pc + 2) // TODO Check this, PoP is self-contradictory on p.11-642
		cpuPtr.pc = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)

	default:
		log.Printf("ERROR: EAGLE_PC instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	return true
}
