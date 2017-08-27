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
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
	"mvemg/util"
)

func eclipseStack(cpuPtr *CPUT, iPtr *decodedInstrT) bool {
	var (
		addr                dg.PhysAddrT
		first, last, thisAc int
		noAccModeInd2Word   noAccModeInd2WordT
		twoAcc1Word         twoAcc1WordT
		unique2Word         unique2WordT
	)
	acsUp := [8]int{0, 1, 2, 3, 0, 1, 2, 3}

	switch iPtr.mnemonic {

	case "POP":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		first = twoAcc1Word.acs
		last = twoAcc1Word.acd
		if last > first {
			first += 4
		}
		for thisAc = first; thisAc >= last; thisAc-- {
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "... narrow popping AC%d\n", acsUp[thisAc])
			}
			cpuPtr.ac[acsUp[thisAc]] = dg.DwordT(memory.NsPop(0))
		}

	case "POPJ":
		addr = dg.PhysAddrT(memory.NsPop(0))
		cpuPtr.pc = addr
		return true // because PC set

	case "PSH":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		first = twoAcc1Word.acs
		last = twoAcc1Word.acd
		if last < first {
			last += 4
		}
		for thisAc = first; thisAc <= last; thisAc++ {
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "... narrow pushing AC%d\n", acsUp[thisAc])
			}
			memory.NsPush(0, util.DWordGetLowerWord(cpuPtr.ac[acsUp[thisAc]]))
		}

	case "PSHJ":
		noAccModeInd2Word = iPtr.variant.(noAccModeInd2WordT)
		memory.NsPush(0, dg.WordT(cpuPtr.pc)+2)
		addr = resolve16bitEclipseAddr(cpuPtr, noAccModeInd2Word.ind, noAccModeInd2Word.mode, noAccModeInd2Word.disp15)
		cpuPtr.pc = addr
		return true // because PC set

	case "RTN":
		// complement of SAVE
		memory.WriteWord(memory.NspLoc, memory.ReadWord(memory.NfpLoc)) // ???
		word := memory.NsPop(0)
		cpuPtr.carry = util.TestWbit(word, 0)
		cpuPtr.pc = dg.PhysAddrT(word) & 0x7fff
		//nfpSave := memory.NsPop(0)               // 1
		cpuPtr.ac[3] = dg.DwordT(memory.NsPop(0)) // 2
		cpuPtr.ac[2] = dg.DwordT(memory.NsPop(0)) // 3
		cpuPtr.ac[1] = dg.DwordT(memory.NsPop(0)) // 4
		cpuPtr.ac[0] = dg.DwordT(memory.NsPop(0)) // 5
		memory.WriteWord(memory.NfpLoc, util.DWordGetLowerWord(cpuPtr.ac[3]))
		return true // because PC set

		//		nfpSav := memory.ReadWord(NFP_LOC)
		//		pwd1 := memory.NsPop(0) // 1
		//		cpuPtr.carry = util.TestWbit(pwd1, 0)
		//		cpuPtr.pc = dg_phys_addr(pwd1 & 0x07fff)
		//		cpuPtr.ac[3] = dg_dword(memory.NsPop(0)) // 2
		//		cpuPtr.ac[2] = dg_dword(memory.NsPop(0)) // 3
		//		cpuPtr.ac[1] = dg_dword(memory.NsPop(0)) // 4
		//		cpuPtr.ac[0] = dg_dword(memory.NsPop(0)) // 5
		//		WriteWord(NSP_LOC, nfpSav-5)
		//		WriteWord(NFP_LOC, util.DWordGetLowerWord(cpuPtr.ac[3]))

		//return true // because PC set

	case "SAVE":
		unique2Word = iPtr.variant.(unique2WordT)
		nfpSav := memory.ReadWord(memory.NfpLoc)
		nspSav := memory.ReadWord(memory.NspLoc)
		memory.NsPush(0, util.DWordGetLowerWord(cpuPtr.ac[0])) // 1
		memory.NsPush(0, util.DWordGetLowerWord(cpuPtr.ac[1])) // 2
		memory.NsPush(0, util.DWordGetLowerWord(cpuPtr.ac[2])) // 3
		memory.NsPush(0, nfpSav)                               // 4
		word := util.DWordGetLowerWord(cpuPtr.ac[3])
		if cpuPtr.carry {
			word |= 0x8000
		} else {
			word &= 0x7fff
		}
		memory.NsPush(0, word) // 5
		wdCnt := int(unique2Word.immU16)
		if wdCnt > 0 {
			for wd := 0; wd < wdCnt; wd++ {
				memory.NsPush(0, 0) // ...
			}
		}
		cpuPtr.ac[3] = dg.DwordT(memory.ReadWord(memory.NspLoc)) // ???
		memory.WriteWord(memory.NfpLoc, dg.WordT(cpuPtr.ac[3]))  // ???
		memory.WriteWord(memory.NspLoc, nspSav+5)
		//cpuPtr.ac[3] = dg_dword(nspSav + 5)

	default:
		log.Fatalf("ERROR: ECLIPSE_STACK instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg.PhysAddrT(iPtr.instrLength)
	return true
}
