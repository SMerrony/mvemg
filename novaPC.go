// novaPC.go
package main

func novaPC(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	switch iPtr.mnemonic {

	case "JMP":
		// disp is only 8-bit, but same resolution code
		cpuPtr.pc = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)

	case "JSR":
		tmpPC := dg_dword(cpuPtr.pc + 1)
		cpuPtr.pc = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		cpuPtr.ac[3] = tmpPC

	default:
		debugPrint(DEBUG_LOG, "ERROR: NOVA_PC instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}
	return true
}
