// novaPC.go
package main

import (
	"log"
)

func novaPC(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	switch iPtr.mnemonic {

	case "JMP":
		// disp is only 8-bit, but same resolution code
		cpuPtr.pc = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)

	case "JSR":
		cpuPtr.ac[3] = dg_dword(cpuPtr.pc + 1)
		cpuPtr.pc = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)

	default:
		log.Printf("ERROR: NOVA_PC instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}
	return true
}