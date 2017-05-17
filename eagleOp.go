// eagleOp.go
package main

import (
	"log"
)

func eagleOp(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	//var addr dg_phys_addr
	//var wd dg_word
	var (
		res int32
	)

	switch iPtr.mnemonic {

	case "WADD":
		res = int32(cpuPtr.ac[iPtr.acs]) + int32(cpuPtr.ac[iPtr.acd])
		cpuPtr.ac[iPtr.acd] = dg_dword(res)
		// FIXME - handle overflow and carry

	default:
		log.Printf("ERROR: EAGLE_OP instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
