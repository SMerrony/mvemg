// eaglePC.go

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
	"mvemg/memory"
	"mvemg/util"
)

func eaglePC(cpuPtr *CPUT, iPtr *decodedInstrT) bool {

	var (
		wd                     dg.WordT
		dwd, tmp32b            dg.DwordT
		tmpAddr                dg.PhysAddrT
		s32a, s32b             int32
		noAccModeInd2Word      noAccModeInd2WordT
		noAccModeInd3Word      noAccModeInd3WordT
		noAccModeInd3WordXcall noAccModeInd3WordXcallT
		noAccModeInd4Word      noAccModeInd4WordT
		oneAccImm2Word         oneAccImm2WordT
		twoAcc1Word            twoAcc1WordT
		split8bitDisp          split8bitDispT
		wskb                   wskbT
	)

	switch iPtr.mnemonic {

	case "LCALL": // FIXME - LCALL only handling trivial case, no checking
		noAccModeInd4Word = iPtr.variant.(noAccModeInd4WordT)
		cpuPtr.ac[3] = dg.DwordT(cpuPtr.pc) + 4
		if noAccModeInd4Word.argCount > 0 {
			dwd = dg.DwordT(noAccModeInd4Word.argCount) & 0x00007fff
		} else {
			// TODO PSR
			dwd = dg.DwordT(noAccModeInd4Word.argCount)
		}
		memory.WsPush(0, dwd)
		cpuPtr.pc = resolve32bitEffAddr(cpuPtr, noAccModeInd4Word.ind, noAccModeInd4Word.mode, noAccModeInd4Word.disp31)
		cpuPtr.ovk = false

	case "LJMP":
		noAccModeInd3Word = iPtr.variant.(noAccModeInd3WordT)
		cpuPtr.pc = resolve32bitEffAddr(cpuPtr, noAccModeInd3Word.ind, noAccModeInd3Word.mode, noAccModeInd3Word.disp31)

	case "LJSR":
		noAccModeInd3Word = iPtr.variant.(noAccModeInd3WordT)
		cpuPtr.ac[3] = dg.DwordT(cpuPtr.pc) + 3
		cpuPtr.pc = resolve32bitEffAddr(cpuPtr, noAccModeInd3Word.ind, noAccModeInd3Word.mode, noAccModeInd3Word.disp31)

	case "LNISZ":
		noAccModeInd3Word = iPtr.variant.(noAccModeInd3WordT)
		// unsigned narrow increment and skip if zero
		tmpAddr = resolve32bitEffAddr(cpuPtr, noAccModeInd3Word.ind, noAccModeInd3Word.mode, noAccModeInd3Word.disp31)
		wd = memory.ReadWord(tmpAddr) + 1
		memory.WriteWord(tmpAddr, wd)
		if wd == 0 {
			cpuPtr.pc += 4
		} else {
			cpuPtr.pc += 3
		}

	case "LPSHJ":
		noAccModeInd3Word = iPtr.variant.(noAccModeInd3WordT)
		memory.WsPush(0, dg.DwordT(cpuPtr.pc)+3)
		cpuPtr.pc = resolve32bitEffAddr(cpuPtr, noAccModeInd3Word.ind, noAccModeInd3Word.mode, noAccModeInd3Word.disp31)

	case "LWDSZ":
		noAccModeInd3Word = iPtr.variant.(noAccModeInd3WordT)
		// unsigned wide decrement and skip if zero
		tmpAddr = resolve32bitEffAddr(cpuPtr, noAccModeInd3Word.ind, noAccModeInd3Word.mode, noAccModeInd3Word.disp31)
		tmp32b = memory.ReadDWord(tmpAddr) - 1
		memory.WriteDWord(tmpAddr, tmp32b)
		if tmp32b == 0 {
			cpuPtr.pc += 4
		} else {
			cpuPtr.pc += 3
		}

	case "WBR":
		//		if iPtr.disp > 0 {
		//			cpuPtr.pc += dg_phys_addr(iPtr.disp)
		//		} else {
		//			cpuPtr.pc -= dg_phys_addr(iPtr.disp)
		//		}
		split8bitDisp = iPtr.variant.(split8bitDispT)
		cpuPtr.pc += dg.PhysAddrT(int32(split8bitDisp.disp8))

	case "WPOPJ":
		dwd = memory.WsPop(0)
		cpuPtr.pc = dg.PhysAddrT(dwd) & 0x0fffffff

	case "WRTN": // FIXME incomplete: handle PSR and rings
		wspSav := memory.ReadDWord(memory.WspLoc)
		wfpSav := memory.ReadDWord(memory.WfpLoc)
		// set WSP equal to WFP
		memory.WriteDWord(memory.WspLoc, wfpSav)
		// pop off 6 double words
		dwd = memory.WsPop(0) // 1
		cpuPtr.carry = util.TestDWbit(dwd, 0)
		cpuPtr.pc = dg.PhysAddrT(dwd & 0x7fffffff)
		cpuPtr.ac[3] = memory.WsPop(0) // 2
		// replace WFP with popped value of AC3
		memory.WriteDWord(memory.WfpLoc, cpuPtr.ac[3])
		cpuPtr.ac[2] = memory.WsPop(0) // 3
		cpuPtr.ac[1] = memory.WsPop(0) // 4
		cpuPtr.ac[0] = memory.WsPop(0) // 5
		dwd = memory.WsPop(0)          // 6
		// TODO Set PSR
		wsFramSz2 := int(dwd&0x00007fff) * 2
		memory.WriteDWord(memory.WspLoc, wspSav-dg.DwordT(wsFramSz2)-12)

	case "WSEQ": // Signedness doen't matter for equality testing
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		if twoAcc1Word.acd == twoAcc1Word.acs {
			tmp32b = 0
		} else {
			tmp32b = cpuPtr.ac[twoAcc1Word.acd]
		}
		if cpuPtr.ac[twoAcc1Word.acs] == tmp32b {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSEQI":
		oneAccImm2Word = iPtr.variant.(oneAccImm2WordT)
		tmp32b = dg.DwordT(int32(oneAccImm2Word.immS16))
		if cpuPtr.ac[oneAccImm2Word.acd] == tmp32b {
			cpuPtr.pc += 3
		} else {
			cpuPtr.pc += 2
		}

	case "WSGE": // wide signed
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		if twoAcc1Word.acd == twoAcc1Word.acs {
			s32a = 0
		} else {
			s32a = int32(cpuPtr.ac[twoAcc1Word.acd]) // this does the right thing in Go
		}
		s32b = int32(cpuPtr.ac[twoAcc1Word.acs])
		if s32b >= s32a {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSGT":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		if twoAcc1Word.acd == twoAcc1Word.acs {
			s32a = 0
		} else {
			s32a = int32(cpuPtr.ac[twoAcc1Word.acd]) // this does the right thing in Go
		}
		s32b = int32(cpuPtr.ac[twoAcc1Word.acs])
		if s32b > s32a {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSKBO":
		wskb = iPtr.variant.(wskbT)
		if util.TestDWbit(cpuPtr.ac[0], wskb.bitNum) {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSKBZ":
		wskb = iPtr.variant.(wskbT)
		if !util.TestDWbit(cpuPtr.ac[0], wskb.bitNum) {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSLE":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		if twoAcc1Word.acd == twoAcc1Word.acs {
			s32a = 0
		} else {
			s32a = int32(cpuPtr.ac[twoAcc1Word.acd]) // this does the right thing in Go
		}
		s32b = int32(cpuPtr.ac[twoAcc1Word.acs])
		if s32b <= s32a {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSLT":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		if twoAcc1Word.acd == twoAcc1Word.acs {
			s32a = 0
		} else {
			s32a = int32(cpuPtr.ac[twoAcc1Word.acd]) // this does the right thing in Go
		}
		s32b = int32(cpuPtr.ac[twoAcc1Word.acs])
		if s32b < s32a {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSNE":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		if twoAcc1Word.acd == twoAcc1Word.acs {
			tmp32b = 0
		} else {
			tmp32b = cpuPtr.ac[twoAcc1Word.acd]
		}
		if cpuPtr.ac[twoAcc1Word.acs] != tmp32b {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "XCALL":
		noAccModeInd3WordXcall = iPtr.variant.(noAccModeInd3WordXcallT)
		// FIXME - only handling the trivial case so far
		cpuPtr.ac[3] = dg.DwordT(cpuPtr.pc) + 3
		if noAccModeInd3WordXcall.argCount > 0 {
			dwd = dg.DwordT(noAccModeInd3WordXcall.argCount) & 0x00007fff
		} else {
			// TODO PSR
			dwd = dg.DwordT(noAccModeInd3WordXcall.argCount)
		}
		memory.WsPush(0, dwd)
		cpuPtr.pc = resolve16bitEagleAddr(cpuPtr, noAccModeInd3WordXcall.ind, noAccModeInd3WordXcall.mode,
			noAccModeInd3WordXcall.disp15)

	case "XJMP":
		noAccModeInd2Word = iPtr.variant.(noAccModeInd2WordT)
		cpuPtr.pc = resolve16bitEagleAddr(cpuPtr, noAccModeInd2Word.ind, noAccModeInd2Word.mode, noAccModeInd2Word.disp15)

	case "XJSR":
		noAccModeInd2Word = iPtr.variant.(noAccModeInd2WordT)
		cpuPtr.ac[3] = dg.DwordT(cpuPtr.pc + 2) // TODO Check this, PoP is self-contradictory on p.11-642
		cpuPtr.pc = resolve16bitEagleAddr(cpuPtr, noAccModeInd2Word.ind, noAccModeInd2Word.mode, noAccModeInd2Word.disp15)

	case "XNISZ": // unsigned narrow increment and skip if zero
		noAccModeInd2Word = iPtr.variant.(noAccModeInd2WordT)
		tmpAddr = resolve16bitEagleAddr(cpuPtr, noAccModeInd2Word.ind, noAccModeInd2Word.mode, noAccModeInd2Word.disp15)
		wd = memory.ReadWord(tmpAddr)
		wd++ // N.B. have checked that 0xffff + 1 == 0 in Go
		memory.WriteWord(tmpAddr, wd)
		if wd == 0 {
			cpuPtr.pc += 3
		} else {
			cpuPtr.pc += 2
		}

	case "XWDSZ":
		noAccModeInd2Word = iPtr.variant.(noAccModeInd2WordT)
		tmpAddr = resolve16bitEagleAddr(cpuPtr, noAccModeInd2Word.ind, noAccModeInd2Word.mode, noAccModeInd2Word.disp15)
		dwd = memory.ReadDWord(tmpAddr)
		dwd--
		memory.WriteDWord(tmpAddr, dwd)
		if dwd == 0 {
			cpuPtr.pc += 3
		} else {
			cpuPtr.pc += 2
		}

	default:
		log.Fatalf("ERROR: EAGLE_PC instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	return true
}
