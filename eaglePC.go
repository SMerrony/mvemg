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

func eaglePC(cpuPtr *CPU, iPtr *decodedInstrT) bool {

	var (
		wd          dg.WordT
		dwd, tmp32b dg.DwordT
		tmpAddr     dg.PhysAddrT
		s32a, s32b  int32
	)

	switch iPtr.mnemonic {

	case "LCALL": // FIXME - LCALL only handling trivial case, no checking
		cpuPtr.ac[3] = dg.DwordT(cpuPtr.pc) + 4
		if iPtr.argCount > 0 {
			dwd = dg.DwordT(iPtr.argCount) & 0x00007fff
		} else {
			// TODO PSR
			dwd = dg.DwordT(iPtr.argCount)
		}
		memory.WsPush(0, dwd)
		cpuPtr.pc = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp31)
		cpuPtr.ovk = false

	case "LJMP":
		cpuPtr.pc = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp31)

	case "LJSR":
		cpuPtr.ac[3] = dg.DwordT(cpuPtr.pc) + 3
		cpuPtr.pc = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp31)

	case "LNISZ":
		// unsigned narrow increment and skip if zero
		tmpAddr = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp31)
		wd = memory.ReadWord(tmpAddr) + 1
		memory.WriteWord(tmpAddr, wd)
		if wd == 0 {
			cpuPtr.pc += 4
		} else {
			cpuPtr.pc += 3
		}

	case "LPSHJ":
		memory.WsPush(0, dg.DwordT(cpuPtr.pc)+3)
		cpuPtr.pc = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp31)

	case "LWDSZ":
		// unsigned wide decrement and skip if zero
		tmpAddr = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp31)
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
		cpuPtr.pc += dg.PhysAddrT(int32(iPtr.disp8))

	case "WPOPJ":
		dwd = memory.WsPop(0)
		cpuPtr.pc = dg.PhysAddrT(dwd) & 0x0fffffff

	case "WRTN": // FIXME incomplete: handle PSR and rings
		wspSav := memory.ReadDWord(memory.WSP_LOC)
		dwd = memory.WsPop(0) // 1
		cpuPtr.carry = util.TestDWbit(dwd, 0)
		cpuPtr.pc = dg.PhysAddrT(dwd & 0x7fffffff)
		cpuPtr.ac[3] = memory.WsPop(0) // 2
		memory.WriteDWord(memory.WFP_LOC, cpuPtr.ac[3])
		cpuPtr.ac[2] = memory.WsPop(0) // 3
		cpuPtr.ac[1] = memory.WsPop(0) // 4
		cpuPtr.ac[0] = memory.WsPop(0) // 5
		dwd = memory.WsPop(0)          // 6
		wsFramSz2 := int(dwd&0x00007fff) * 2
		memory.WriteDWord(memory.WSP_LOC, wspSav-dg.DwordT(wsFramSz2)-12)

	case "WSEQ": // Signedness doen't matter for equality testing
		if iPtr.acd == iPtr.acs {
			tmp32b = 0
		} else {
			tmp32b = cpuPtr.ac[iPtr.acd]
		}
		if cpuPtr.ac[iPtr.acs] == tmp32b {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSEQI":
		tmp32b = dg.DwordT(int32(iPtr.immS16))
		if cpuPtr.ac[iPtr.acd] == tmp32b {
			cpuPtr.pc += 3
		} else {
			cpuPtr.pc += 2
		}

	case "WSGE": // wide signed
		if iPtr.acd == iPtr.acs {
			s32a = 0
		} else {
			s32a = int32(cpuPtr.ac[iPtr.acd]) // this does the right thing in Go
		}
		s32b = int32(cpuPtr.ac[iPtr.acs])
		if s32b >= s32a {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSGT":
		if iPtr.acd == iPtr.acs {
			s32a = 0
		} else {
			s32a = int32(cpuPtr.ac[iPtr.acd]) // this does the right thing in Go
		}
		s32b = int32(cpuPtr.ac[iPtr.acs])
		if s32b > s32a {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSKBO":
		if util.TestDWbit(cpuPtr.ac[0], iPtr.bitNum) {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSKBZ":
		if !util.TestDWbit(cpuPtr.ac[0], iPtr.bitNum) {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSLE":
		if iPtr.acd == iPtr.acs {
			s32a = 0
		} else {
			s32a = int32(cpuPtr.ac[iPtr.acd]) // this does the right thing in Go
		}
		s32b = int32(cpuPtr.ac[iPtr.acs])
		if s32b <= s32a {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSLT":
		if iPtr.acd == iPtr.acs {
			s32a = 0
		} else {
			s32a = int32(cpuPtr.ac[iPtr.acd]) // this does the right thing in Go
		}
		s32b = int32(cpuPtr.ac[iPtr.acs])
		if s32b < s32a {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSNE":
		if iPtr.acd == iPtr.acs {
			tmp32b = 0
		} else {
			tmp32b = cpuPtr.ac[iPtr.acd]
		}
		if cpuPtr.ac[iPtr.acs] != tmp32b {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "XCALL":
		// FIXME - only handling the trivial case so far
		cpuPtr.ac[3] = dg.DwordT(cpuPtr.pc) + 3
		if iPtr.argCount > 0 {
			dwd = dg.DwordT(iPtr.argCount) & 0x00007fff
		} else {
			// TODO PSR
			dwd = dg.DwordT(iPtr.argCount)
		}
		memory.WsPush(0, dwd)
		cpuPtr.pc = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)

	case "XJMP":
		cpuPtr.pc = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)

	case "XJSR":
		cpuPtr.ac[3] = dg.DwordT(cpuPtr.pc + 2) // TODO Check this, PoP is self-contradictory on p.11-642
		cpuPtr.pc = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)

	case "XNISZ": // unsigned narrow increment and skip if zero
		tmpAddr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)
		wd = memory.ReadWord(tmpAddr)
		wd++ // N.B. have checked that 0xffff + 1 == 0 in Go
		memory.WriteWord(tmpAddr, wd)
		if wd == 0 {
			cpuPtr.pc += 3
		} else {
			cpuPtr.pc += 2
		}

	case "XWDSZ":
		tmpAddr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)
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
