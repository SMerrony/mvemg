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

	case "SAVE":
		nfpSav := memReadWord(NFP_LOC)
		nsPush(0, dwordGetLowerWord(cpuPtr.ac[0])) // 1
		nsPush(0, dwordGetLowerWord(cpuPtr.ac[1])) // 2
		nsPush(0, dwordGetLowerWord(cpuPtr.ac[2])) // 3
		nsPush(0, nfpSav)                          // 4
		word := dwordGetLowerWord(cpuPtr.ac[3])
		if cpuPtr.carry {
			word |= 0x8000
		} else {
			word &= 0x7fff
		}
		nsPush(0, word) // 5
		wdCnt := int(iPtr.imm16b)
		if wdCnt > 0 {
			for wd := 0; wd < wdCnt; wd++ {
				nsPush(0, 0) // ...
			}
		}
		cpuPtr.ac[3] = dg_dword(memReadWord(NSP_LOC)) // ???
		memWriteWord(NFP_LOC, dg_word(cpuPtr.ac[3]))

	default:
		log.Printf("ERROR: ECLIPSE_STACK instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
