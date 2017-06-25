// eclipseStack.go
package main

import (
	"log"
	"mvemg/logging"
)

func eclipseStack(cpuPtr *CPU, iPtr *DecodedInstr) bool {
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
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "... popping AC%d\n", acsUp[thisAc])
			}
			cpuPtr.ac[acsUp[thisAc]] = dg_dword(nsPop(0))
		}

	case "POPJ":
		addr = dg_phys_addr(nsPop(0))
		cpuPtr.pc = addr
		return true // because PC set

	case "PSH":
		first = iPtr.acs
		last = iPtr.acd
		if last < first {
			last += 4
		}
		for thisAc = first; thisAc <= last; thisAc++ {
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "... pushing AC%d\n", acsUp[thisAc])
			}
			nsPush(0, dwordGetLowerWord(cpuPtr.ac[acsUp[thisAc]]))
		}

	case "PSHJ":
		nsPush(0, dg_word(cpuPtr.pc)+2)
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		cpuPtr.pc = addr
		return true // because PC set

	case "RTN":
		// complement of SAVE
		memWriteWord(NSP_LOC, memReadWord(NFP_LOC)) // ???
		word := nsPop(0)
		cpuPtr.carry = testWbit(word, 0)
		cpuPtr.pc = dg_phys_addr(word) & 0x7fff
		//nfpSave := nsPop(0)               // 1
		cpuPtr.ac[3] = dg_dword(nsPop(0)) // 2
		cpuPtr.ac[2] = dg_dword(nsPop(0)) // 3
		cpuPtr.ac[1] = dg_dword(nsPop(0)) // 4
		cpuPtr.ac[0] = dg_dword(nsPop(0)) // 5
		memWriteWord(NFP_LOC, dwordGetLowerWord(cpuPtr.ac[3]))
		return true // because PC set

		//		nfpSav := memReadWord(NFP_LOC)
		//		pwd1 := nsPop(0) // 1
		//		cpuPtr.carry = testWbit(pwd1, 0)
		//		cpuPtr.pc = dg_phys_addr(pwd1 & 0x07fff)
		//		cpuPtr.ac[3] = dg_dword(nsPop(0)) // 2
		//		cpuPtr.ac[2] = dg_dword(nsPop(0)) // 3
		//		cpuPtr.ac[1] = dg_dword(nsPop(0)) // 4
		//		cpuPtr.ac[0] = dg_dword(nsPop(0)) // 5
		//		memWriteWord(NSP_LOC, nfpSav-5)
		//		memWriteWord(NFP_LOC, dwordGetLowerWord(cpuPtr.ac[3]))

		//return true // because PC set

	case "SAVE":
		nfpSav := memReadWord(NFP_LOC)
		nspSav := memReadWord(NSP_LOC)
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
		memWriteWord(NFP_LOC, dg_word(cpuPtr.ac[3]))  // ???
		memWriteWord(NSP_LOC, nspSav+5)
		//cpuPtr.ac[3] = dg_dword(nspSav + 5)

	default:
		log.Fatalf("ERROR: ECLIPSE_STACK instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
