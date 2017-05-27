// eclipseOp.go
package main

import (
	"log"
)

func eclipseOp(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	var (
		addr dg_phys_addr
		wd   dg_word
		dwd  dg_dword
	)

	switch iPtr.mnemonic {

	case "ELEF":
		cpuPtr.ac[iPtr.acd] = dg_dword(resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp))

	case "ESTA":
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		memWriteWord(addr, dwordGetLowerWord(cpuPtr.ac[iPtr.acd]))

	case "HXL":
		dwd = cpuPtr.ac[iPtr.acd] << (uint32(iPtr.immVal) * 4)
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	case "HXR":
		dwd = cpuPtr.ac[iPtr.acd] >> (uint32(iPtr.immVal) * 4)
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	case "SBI": // unsigned
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		wd -= dg_word(iPtr.immVal) // TODO this is a signed int, is this OK?
		cpuPtr.ac[iPtr.acd] = dg_dword(wd)

	default:
		log.Printf("ERROR: ECLIPSE_OP instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
