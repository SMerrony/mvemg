// eclipseOp.go
package main

import (
	"log"
)

func eclipseOp(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	var (
		//addr dg_phys_addr
		//wd   dg_word
		dwd dg_dword
	)

	switch iPtr.mnemonic {

	case "HXL":
		dwd = cpuPtr.ac[iPtr.acd] << (uint32(iPtr.immVal) * 4)
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	case "HXR":
		dwd = cpuPtr.ac[iPtr.acd] >> (uint32(iPtr.immVal) * 4)
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	default:
		log.Printf("ERROR: ECLIPSE_OP instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
