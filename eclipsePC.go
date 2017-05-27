// eclipsePC.go
package main

import (
	"fmt"
	"log"
)

func eclipsePC(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	var (
		addr, inc dg_phys_addr
		acs, h, l int16
	//wd   dg_word
	//dwd dg_dword
	)

	switch iPtr.mnemonic {

	case "CLM": // signed compare to limits
		acs = int16(dwordGetLowerWord(cpuPtr.ac[iPtr.acs]))
		if iPtr.acs == iPtr.acd {
			l = int16(memReadWord(cpuPtr.pc + 1))
			h = int16(memReadWord(cpuPtr.pc + 2))
			if acs < l || acs > h {
				inc = 3
			} else {
				inc = 4
			}
		} else {
			l = int16(memReadWord(dg_phys_addr(dwordGetLowerWord(cpuPtr.ac[iPtr.acd]))))
			h = int16(memReadWord(dg_phys_addr(dwordGetLowerWord(cpuPtr.ac[iPtr.acd]) + 1)))
			if acs < l || acs > h {
				inc = 1
			} else {
				inc = 2
			}
		}
		debugPrint(SYSTEM_LOG, fmt.Sprintf("CLM compared %d with limits %d and %d, moving PC by %d\n", acs, l, h, inc))
		cpuPtr.pc += inc

	case "EJMP":
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		cpuPtr.pc = addr

	case "EJSR":
		cpuPtr.ac[3] = dg_dword(cpuPtr.pc) + 2
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		cpuPtr.pc = addr

	default:
		log.Printf("ERROR: ECLIPSE_PC instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
