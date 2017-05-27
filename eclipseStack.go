// eclipseStack.go
package main

import (
	"fmt"
	"log"
)

func eclipseStack(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	var (
		addr                dg_phys_addr
		first, last, thisAc int
	//wd   dg_word
	//dwd dg_dword
	)
	acsUp := [8]int{0, 1, 2, 3, 0, 1, 2, 3}

	switch iPtr.mnemonic {

	case "POP":
		first = iPtr.acs
		last = iPtr.acd
		if last > first {
			first += 4
		}
		for thisAc = first; thisAc >= last; thisAc-- {
			debugPrint(SYSTEM_LOG, fmt.Sprintf("... popping AC%d\n", acsUp[thisAc]))
			cpuPtr.ac[acsUp[thisAc]] = dg_dword(nsPop(0))
		}

	case "POPJ":
		addr = dg_phys_addr(nsPop(0))
		cpuPtr.pc = addr
		return true

	case "PSH":
		first = iPtr.acs
		last = iPtr.acd
		if last < first {
			last += 4
		}
		for thisAc = first; thisAc <= last; thisAc++ {
			debugPrint(SYSTEM_LOG, fmt.Sprintf("... pushing AC%d\n", acsUp[thisAc]))
			nsPush(0, dwordGetLowerWord(cpuPtr.ac[acsUp[thisAc]]))
		}

	case "PSHJ":
		nsPush(0, dg_word(cpuPtr.pc)+2)
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		cpuPtr.pc = addr
		return true

	default:
		log.Printf("ERROR: ECLIPSE_STACK instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
