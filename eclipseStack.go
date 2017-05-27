// eclipseStack.go
package main

import (
	"fmt"
	"log"
)

func eclipseStack(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	var (
		addr, inc dg_phys_addr
		acs, h, l int16
	//wd   dg_word
	//dwd dg_dword
	)
	acsUp := [8]int{0, 1, 2, 3, 0, 1, 2, 3}

	switch iPtr.mnemonic {

	default:
		log.Printf("ERROR: ECLIPSE_STACK instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
