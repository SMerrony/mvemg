// eaglePC.go
package main

import (
	"log"
)

func eaglePC(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	//var addr dg_phys_addr
	var tmp32b dg_dword

	switch iPtr.mnemonic {

	case "WBR":
		//		if iPtr.disp > 0 {
		//			cpuPtr.pc += dg_phys_addr(iPtr.disp)
		//		} else {
		//			cpuPtr.pc -= dg_phys_addr(iPtr.disp)
		//		}
		cpuPtr.pc += dg_phys_addr(iPtr.disp)

	case "WSEQ":
		if iPtr.acd == iPtr.acs {
			tmp32b = 0
		} else {
			tmp32b = cpuPtr.ac[iPtr.acd]
		}
		if cpuPtr.ac[iPtr.acs] == tmp32b {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

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
