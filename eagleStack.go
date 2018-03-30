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

////////////////////////////////////////////////////////////////
// N.B. Be sure to use Double-Word memory references here... //
//////////////////////////////////////////////////////////////

package main

import (
	"log"

	"github.com/SMerrony/dgemug/logging"

	"github.com/SMerrony/dgemug/util"

	"github.com/SMerrony/dgemug/memory"

	"github.com/SMerrony/dgemug/dg"
)

func eagleStack(cpuPtr *CPUT, iPtr *decodedInstrT) bool {

	var (
		firstAc, lastAc, thisAc int
		acsUp                   = [8]int{0, 1, 2, 3, 0, 1, 2, 3}
		tmpDwd                  dg.DwordT
		tmpQwd                  dg.QwordT
		immMode2Word            immMode2WordT
		noAccMode2Word          noAccMode2WordT
		noAccMode3Word          noAccMode3WordT
		noAccModeInd2Word       noAccModeInd2WordT
		noAccModeInd3Word       noAccModeInd3WordT
		oneAcc1Word             oneAcc1WordT
		twoAcc1Word             twoAcc1WordT
		unique2Word             unique2WordT
	)

	switch iPtr.ix {

	// N.B. DSZTS and ISZTS are in eaglePC

	case instrLDAFP:
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		cpuPtr.ac[oneAcc1Word.acd] = memory.ReadDWord(memory.WfpLoc)

	case instrLDASB:
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		cpuPtr.ac[oneAcc1Word.acd] = memory.ReadDWord(memory.WsbLoc)

	case instrLDASL:
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		cpuPtr.ac[oneAcc1Word.acd] = memory.ReadDWord(memory.WslLoc)

	case instrLDASP:
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		cpuPtr.ac[oneAcc1Word.acd] = memory.ReadDWord(memory.WspLoc)

	case instrLPEF:
		noAccModeInd3Word = iPtr.variant.(noAccModeInd3WordT)
		memory.WsPush(0, dg.DwordT(resolve32bitEffAddr(cpuPtr, noAccModeInd3Word.ind, noAccModeInd3Word.mode, noAccModeInd3Word.disp31)))

	case instrLPEFB:
		noAccMode3Word = iPtr.variant.(noAccMode3WordT)
		eff := dg.DwordT(noAccMode3Word.immU32)
		switch noAccMode3Word.mode {
		case absoluteMode: // do nothing
		case pcMode:
			eff += dg.DwordT(cpuPtr.pc)
		case ac2Mode:
			eff += cpuPtr.ac[2]
		case ac3Mode:
			eff += cpuPtr.ac[3]
		}
		memory.WsPush(0, eff)

	case instrSTAFP:
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		// FIXME handle segments
		memory.WriteDWord(memory.WfpLoc, cpuPtr.ac[oneAcc1Word.acd])

	case instrSTASB:
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		// FIXME handle segments
		memory.WriteDWord(memory.WsbLoc, cpuPtr.ac[oneAcc1Word.acd])

	case instrSTASL:
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		// FIXME handle segments
		memory.WriteDWord(memory.WslLoc, cpuPtr.ac[oneAcc1Word.acd])

	case instrSTASP:
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		// FIXME handle segments
		memory.WriteDWord(memory.WspLoc, cpuPtr.ac[oneAcc1Word.acd])
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... STASP set WSP to %d\n", cpuPtr.ac[oneAcc1Word.acd])
		}

	case instrSTATS:
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		// FIXME handle segments
		memory.WriteDWord(dg.PhysAddrT(memory.ReadDWord(memory.WslLoc)), cpuPtr.ac[oneAcc1Word.acd])

	case instrWFPOP:
		cpuPtr.fpac[3] = memory.WsPopQWord(0)
		cpuPtr.fpac[2] = memory.WsPopQWord(0)
		cpuPtr.fpac[1] = memory.WsPopQWord(0)
		cpuPtr.fpac[0] = memory.WsPopQWord(0)
		tmpQwd = memory.WsPopQWord(0)
		cpuPtr.fpsr = 0
		any := false
		// set the ANY bit?
		if memory.GetQwbits(tmpQwd, 1, 4) != 0 {
			memory.SetQwbit(&cpuPtr.fpsr, 0)
			any = true
		}
		// copy bits 1-11
		for b := 1; b <= 11; b++ {
			if memory.TestQwbit(tmpQwd, b) {
				memory.SetQwbit(&cpuPtr.fpsr, uint(b))
			}
		}
		// bits 28-31
		if any {
			for b := 28; b <= 31; b++ {
				if memory.TestQwbit(tmpQwd, b) {
					memory.SetQwbit(&cpuPtr.fpsr, uint(b))
				}
			}
			for b := 33; b <= 63; b++ {
				if memory.TestQwbit(tmpQwd, b) {
					memory.SetQwbit(&cpuPtr.fpsr, uint(b))
				}
			}
		}

	case instrWMSP:
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		tmpDwd = cpuPtr.ac[oneAcc1Word.acd] << 1
		tmpDwd += memory.ReadDWord(memory.WspLoc)
		memory.WriteDWord(memory.WspLoc, tmpDwd)

	case instrWPOP:
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		firstAc = twoAcc1Word.acs
		lastAc = twoAcc1Word.acd
		if lastAc > firstAc {
			firstAc += 4
		}
		for thisAc = firstAc; thisAc >= lastAc; thisAc-- {
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "... wide popping AC%d\n", acsUp[thisAc])
			}
			cpuPtr.ac[acsUp[thisAc]] = memory.WsPop(0)
		}

	case instrWPSH:
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		firstAc = twoAcc1Word.acs
		lastAc = twoAcc1Word.acd
		if lastAc < firstAc {
			lastAc += 4
		}
		for thisAc = firstAc; thisAc <= lastAc; thisAc++ {
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "... wide pushing AC%d\n", acsUp[thisAc])
			}
			memory.WsPush(0, cpuPtr.ac[acsUp[thisAc]])
		}

	// N.B. WRTN is in eaglePC

	case instrWSAVR:
		unique2Word = iPtr.variant.(unique2WordT)
		wsav(cpuPtr, &unique2Word)
		cpuSetOVK(false)

	case instrWSAVS:
		unique2Word = iPtr.variant.(unique2WordT)
		wsav(cpuPtr, &unique2Word)
		cpuSetOVK(true)

	case instrWSSVR:
		unique2Word = iPtr.variant.(unique2WordT)
		wssav(cpuPtr, &unique2Word)
		cpuSetOVK(false)
		cpuSetOVR(false)

	case instrXPEF:
		noAccModeInd2Word = iPtr.variant.(noAccModeInd2WordT)
		// FIXME segment handling, check for overflow
		memory.WsPush(0, dg.DwordT(resolve16bitEagleAddr(cpuPtr, noAccModeInd2Word.ind, noAccModeInd2Word.mode, noAccModeInd2Word.disp15)))

	case instrXPEFB:
		noAccMode2Word = iPtr.variant.(noAccMode2WordT)
		// FIXME segment handling, check for overflow
		eff := dg.DwordT(noAccMode2Word.disp16)
		switch noAccMode2Word.mode {
		case absoluteMode: // do nothing
		case pcMode:
			eff += dg.DwordT(cpuPtr.pc)
		case ac2Mode:
			eff += cpuPtr.ac[2]
		case ac3Mode:
			eff += cpuPtr.ac[3]
		}
		memory.WsPush(0, eff)

	case instrXPSHJ:
		// FIXME segment handling, check for overflow
		immMode2Word = iPtr.variant.(immMode2WordT)
		memory.WsPush(0, dg.DwordT(cpuPtr.pc+2))
		cpuPtr.pc = resolve16bitEagleAddr(cpuPtr, immMode2Word.ind, immMode2Word.mode, immMode2Word.disp15)
		return true

	default:
		log.Fatalf("ERROR: EAGLE_STACK instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg.PhysAddrT(iPtr.instrLength)
	return true
}

// wsav is common to WSAVR and WSAVS
func wsav(cpuPtr *CPUT, u2wd *unique2WordT) {
	wfpSav := memory.ReadDWord(memory.WfpLoc)
	memory.WsPush(0, cpuPtr.ac[0]) // 1
	memory.WsPush(0, cpuPtr.ac[1]) // 2
	memory.WsPush(0, cpuPtr.ac[2]) // 3
	memory.WsPush(0, wfpSav)       // 4
	dwd := cpuPtr.ac[3] & 0x7fffffff
	if cpuPtr.carry {
		dwd |= 0x80000000
	}
	memory.WsPush(0, dwd) // 5
	memory.WriteDWord(memory.WfpLoc, memory.ReadDWord(memory.WspLoc))
	cpuPtr.ac[3] = memory.ReadDWord(memory.WspLoc)
	dwdCnt := uint(u2wd.immU16)
	if dwdCnt > 0 {
		memory.AdvanceWSP(dwdCnt)
	}
}

// wssav is common to WSSVR and WSSVS
func wssav(cpuPtr *CPUT, u2wd *unique2WordT) {
	wfpSav := memory.ReadDWord(memory.WfpLoc)
	memory.WsPush(0, util.DwordFromTwoWords(cpuPtr.psr, 0)) // 1
	memory.WsPush(0, cpuPtr.ac[0])                          // 2
	memory.WsPush(0, cpuPtr.ac[1])                          // 3
	memory.WsPush(0, cpuPtr.ac[2])                          // 4
	memory.WsPush(0, wfpSav)                                // 5
	dwd := cpuPtr.ac[3] & 0x7fffffff
	if cpuPtr.carry {
		dwd |= 0x80000000
	}
	memory.WsPush(0, dwd) // 6
	memory.WriteDWord(memory.WfpLoc, memory.ReadDWord(memory.WspLoc))
	cpuPtr.ac[3] = memory.ReadDWord(memory.WspLoc)
	dwdCnt := uint(u2wd.immU16)
	if dwdCnt > 0 {
		memory.AdvanceWSP(dwdCnt)
	}
}
