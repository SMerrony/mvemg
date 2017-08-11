// novaPC.go
package main

import "log"

func novaPC(cpuPtr *CPU, iPtr *decodedInstrT) bool {
	switch iPtr.mnemonic {

	case "JMP":
		// disp is only 8-bit, but same resolution code
		cpuPtr.pc = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)

	case "JSR":
		tmpPC := DgDwordT(cpuPtr.pc + 1)
		cpuPtr.pc = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)
		cpuPtr.ac[3] = tmpPC

	default:
		log.Fatalf("ERROR: NOVA_PC instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}
	return true
}
