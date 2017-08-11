// eclipseStack.go

// Copyright (C) 2017  Steve Merrony

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package main

import (
	"log"
	"mvemg/logging"
)

func eclipseStack(cpuPtr *CPU, iPtr *decodedInstrT) bool {
	var (
		addr                DgPhysAddrT
		first, last, thisAc int
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
				logging.DebugPrint(logging.DebugLog, "... narrow popping AC%d\n", acsUp[thisAc])
			}
			cpuPtr.ac[acsUp[thisAc]] = DgDwordT(nsPop(0))
		}

	case "POPJ":
		addr = DgPhysAddrT(nsPop(0))
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
				logging.DebugPrint(logging.DebugLog, "... narrow pushing AC%d\n", acsUp[thisAc])
			}
			nsPush(0, dwordGetLowerWord(cpuPtr.ac[acsUp[thisAc]]))
		}

	case "PSHJ":
		nsPush(0, DgWordT(cpuPtr.pc)+2)
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)
		cpuPtr.pc = addr
		return true // because PC set

	case "RTN":
		// complement of SAVE
		memWriteWord(NSP_LOC, memReadWord(NFP_LOC)) // ???
		word := nsPop(0)
		cpuPtr.carry = testWbit(word, 0)
		cpuPtr.pc = DgPhysAddrT(word) & 0x7fff
		//nfpSave := nsPop(0)               // 1
		cpuPtr.ac[3] = DgDwordT(nsPop(0)) // 2
		cpuPtr.ac[2] = DgDwordT(nsPop(0)) // 3
		cpuPtr.ac[1] = DgDwordT(nsPop(0)) // 4
		cpuPtr.ac[0] = DgDwordT(nsPop(0)) // 5
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
		wdCnt := int(iPtr.immU16)
		if wdCnt > 0 {
			for wd := 0; wd < wdCnt; wd++ {
				nsPush(0, 0) // ...
			}
		}
		cpuPtr.ac[3] = DgDwordT(memReadWord(NSP_LOC)) // ???
		memWriteWord(NFP_LOC, DgWordT(cpuPtr.ac[3]))  // ???
		memWriteWord(NSP_LOC, nspSav+5)
		//cpuPtr.ac[3] = dg_dword(nspSav + 5)

	default:
		log.Fatalf("ERROR: ECLIPSE_STACK instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += DgPhysAddrT(iPtr.instrLength)
	return true
}
