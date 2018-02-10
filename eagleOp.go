// eagleOp.go

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

func eagleOp(cpuPtr *CPUT, iPtr *decodedInstrT) bool {
	//var addr dg_phys_addr

	var (
		wd                 dg.WordT
		dwd                dg.DwordT
		lobyte             bool
		res, s32           int32
		s16                int16
		s64                int64
		addr               dg.PhysAddrT
		immOneAcc          immOneAccT
		oneAcc1Word        oneAcc1WordT
		oneAccImm2Word     oneAccImm2WordT
		oneAccImmWd2Word   oneAccImmWd2WordT
		oneAccImm3Word     oneAccImm3WordT
		oneAccImmDwd3Word  oneAccImmDwd3WordT
		oneAccMode3Word    oneAccMode3WordT // LLDB, LLEFB
		oneAccModeInd3Word oneAccModeInd3WordT
		twoAcc1Word        twoAcc1WordT
	)

	switch iPtr.mnemonic {

	case "ADDI":
		oneAccImm2Word = iPtr.variant.(oneAccImm2WordT)
		// signed 16-bit add immediate
		s16 = int16(util.DWordGetLowerWord(cpuPtr.ac[oneAccImm2Word.acd]))
		s16 += oneAccImm2Word.immS16
		cpuPtr.ac[oneAccImm2Word.acd] = dg.DwordT(s16) & 0X0000FFFF

	case "ANDI":
		oneAccImmWd2Word = iPtr.variant.(oneAccImmWd2WordT)
		wd = util.DWordGetLowerWord(cpuPtr.ac[oneAccImmWd2Word.acd])
		cpuPtr.ac[oneAccImmWd2Word.acd] = dg.DwordT(wd&oneAccImmWd2Word.immWord) & 0x0000ffff

	case "CRYTC":
		cpuPtr.carry = !cpuPtr.carry

	case "CRYTO":
		cpuPtr.carry = true

	case "CRYTZ":
		cpuPtr.carry = false

	case "CVWN":
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		dwd = cpuPtr.ac[oneAcc1Word.acd]
		if dwd>>16 != 0 && dwd>>16 != 0xffff {
			cpuPtr.SetOVR(true)
		}
		if util.TestDWbit(dwd, 16) {
			cpuPtr.ac[oneAcc1Word.acd] |= 0xffff0000
		} else {
			cpuPtr.ac[oneAcc1Word.acd] &= 0x0000ffff
		}

	case "LLDB":
		oneAccMode3Word = iPtr.variant.(oneAccMode3WordT)
		addr = resolve32bitEffAddr(cpuPtr, ' ', oneAccMode3Word.mode, oneAccMode3Word.disp31>>1)
		lobyte = util.TestDWbit(dg.DwordT(oneAccMode3Word.disp31), 31)
		cpuPtr.ac[oneAccMode3Word.acd] = dg.DwordT(memory.ReadByte(addr, lobyte))

	case "LLEF":
		oneAccModeInd3Word = iPtr.variant.(oneAccModeInd3WordT)
		cpuPtr.ac[oneAccModeInd3Word.acd] = dg.DwordT(
			resolve32bitEffAddr(cpuPtr, oneAccModeInd3Word.ind, oneAccModeInd3Word.mode, oneAccModeInd3Word.disp31))

	case "LLEFB":
		oneAccMode3Word = iPtr.variant.(oneAccMode3WordT)
		addr = resolve32bitEffAddr(cpuPtr, ' ', oneAccMode3Word.mode, oneAccMode3Word.disp31>>1)
		addr <<= 1
		if util.TestDWbit(dg.DwordT(oneAccMode3Word.disp31), 31) {
			addr |= 1
		}
		cpuPtr.ac[oneAccMode3Word.acd] = dg.DwordT(addr)

	case "LPSR":
		cpuPtr.ac[0] = dg.DwordT(cpuPtr.psr)

	case "NADD": // signed add
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		s16 = int16(cpuPtr.ac[twoAcc1Word.acd]) + int16(cpuPtr.ac[twoAcc1Word.acs])
		cpuPtr.ac[twoAcc1Word.acd] = dg.DwordT(s16)

	case "NADDI":
		oneAccImm2Word = iPtr.variant.(oneAccImm2WordT)
		s16 = int16(cpuPtr.ac[oneAccImm2Word.acd])
		s16 += oneAccImm2Word.immS16
		// FIXME handle overflow
		cpuPtr.ac[oneAccImm2Word.acd] = dg.DwordT(s16)

	case "NLDAI":
		oneAccImm2Word = iPtr.variant.(oneAccImm2WordT)
		cpuPtr.ac[oneAccImm2Word.acd] = dg.DwordT(int32(oneAccImm2Word.immS16))

	case "NSUB": // signed subtract
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		s16 = int16(cpuPtr.ac[twoAcc1Word.acd]) - int16(cpuPtr.ac[twoAcc1Word.acs])
		cpuPtr.ac[twoAcc1Word.acd] = dg.DwordT(s16)

	case "SEX": // Sign EXtend
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		cpuPtr.ac[twoAcc1Word.acd] = util.SexWordToDWord(util.DWordGetLowerWord(cpuPtr.ac[twoAcc1Word.acs]))

	case "SSPT": /* NO-OP - see p.8-5 of MV/10000 Sys Func Chars */
		log.Println("INFO: SSPT is a No-Op on this machine, continuing")

	case "WADC":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		dwd = ^cpuPtr.ac[twoAcc1Word.acs]
		cpuPtr.ac[twoAcc1Word.acd] = dg.DwordT(int32(cpuPtr.ac[twoAcc1Word.acd]) + int32(dwd))
		// FIXME - handle overflow and carry

	case "WADD":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		res = int32(cpuPtr.ac[twoAcc1Word.acs]) + int32(cpuPtr.ac[twoAcc1Word.acd])
		cpuPtr.ac[twoAcc1Word.acd] = dg.DwordT(res)
		// FIXME - handle overflow and carry

	case "WADI":
		// FIXME - handle overflow and carry
		immOneAcc = iPtr.variant.(immOneAccT)
		s32 = int32(cpuPtr.ac[immOneAcc.acd]) + int32(immOneAcc.immU16)
		cpuPtr.ac[immOneAcc.acd] = dg.DwordT(s32)

	case "WADDI":
		// FIXME - handle overflow and carry
		oneAccImm3Word = iPtr.variant.(oneAccImm3WordT)
		s32 = int32(cpuPtr.ac[oneAccImm3Word.acd]) + int32(oneAccImm3Word.immU32)
		cpuPtr.ac[oneAccImm3Word.acd] = dg.DwordT(s32)

	case "WAND":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		cpuPtr.ac[twoAcc1Word.acd] &= cpuPtr.ac[twoAcc1Word.acs]

	case "WANDI":
		oneAccImmDwd3Word = iPtr.variant.(oneAccImmDwd3WordT)
		cpuPtr.ac[oneAccImmDwd3Word.acd] &= oneAccImmDwd3Word.immDword

	case "WCOM":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		cpuPtr.ac[twoAcc1Word.acd] = ^cpuPtr.ac[twoAcc1Word.acs]

	case "WDIVS":
		s64 = int64(util.QWordFromTwoDwords(cpuPtr.ac[0], cpuPtr.ac[1]))
		if cpuPtr.ac[2] == 0 {
			cpuPtr.SetOVR(true)
		} else {
			s32 = int32(cpuPtr.ac[2])
			if s64/int64(s32) < -2147483648 || s64/int64(s32) > 2147483647 {
				cpuPtr.SetOVR(true)
			} else {
				cpuPtr.ac[0] = dg.DwordT(s64 % int64(s32))
				cpuPtr.ac[1] = dg.DwordT(s64 / int64(s32))
			}
		}

	case "WINC":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		cpuPtr.ac[twoAcc1Word.acd] = cpuPtr.ac[twoAcc1Word.acs] + 1

	case "WIOR":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		cpuPtr.ac[twoAcc1Word.acd] |= cpuPtr.ac[twoAcc1Word.acs]

	case "WIORI":
		oneAccImmDwd3Word = iPtr.variant.(oneAccImmDwd3WordT)
		cpuPtr.ac[oneAccImmDwd3Word.acd] |= oneAccImmDwd3Word.immDword

	case "WLDAI":
		oneAccImmDwd3Word = iPtr.variant.(oneAccImmDwd3WordT)
		cpuPtr.ac[oneAccImmDwd3Word.acd] = oneAccImmDwd3Word.immDword

	case "WLSHI":
		oneAccImm2Word = iPtr.variant.(oneAccImm2WordT)
		shiftAmt8 := int8(oneAccImm2Word.immS16 & 0x0ff)
		if shiftAmt8 < 0 { // shift right
			shiftAmt8 *= -1
			dwd = cpuPtr.ac[oneAccImm2Word.acd] >> uint(shiftAmt8)
			cpuPtr.ac[oneAccImm2Word.acd] = dwd
		}
		if shiftAmt8 > 0 { // shift left
			dwd = cpuPtr.ac[oneAccImm2Word.acd] << uint(shiftAmt8)
			cpuPtr.ac[oneAccImm2Word.acd] = dwd
		}

	case "WLSI":
		immOneAcc = iPtr.variant.(immOneAccT)
		cpuPtr.ac[immOneAcc.acd] = cpuPtr.ac[immOneAcc.acd] << immOneAcc.immU16

	case "WMOV":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		cpuPtr.ac[twoAcc1Word.acd] = cpuPtr.ac[twoAcc1Word.acs]

	case "WMOVR":
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		cpuPtr.ac[oneAcc1Word.acd] = cpuPtr.ac[oneAcc1Word.acd] >> 1

	case "WMUL":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		res = int32(cpuPtr.ac[twoAcc1Word.acd]) * int32(cpuPtr.ac[twoAcc1Word.acs])
		// FIXME - handle overflow and carry
		cpuPtr.ac[twoAcc1Word.acd] = dg.DwordT(res)

	case "WNADI": //signed 16-bit
		oneAccImm2Word = iPtr.variant.(oneAccImm2WordT)
		s32 = int32(cpuPtr.ac[oneAccImm2Word.acd]) + int32(oneAccImm2Word.immS16)
		cpuPtr.ac[oneAccImm2Word.acd] = dg.DwordT(s32)

	case "WNEG":
		// FIXME WNEG - handle CARRY/OVR
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		cpuPtr.ac[twoAcc1Word.acd] = (^cpuPtr.ac[twoAcc1Word.acs]) + 1

	case "WSBI":
		immOneAcc = iPtr.variant.(immOneAccT)
		s32 = int32(cpuPtr.ac[immOneAcc.acd]) - int32(immOneAcc.immU16)
		cpuPtr.ac[immOneAcc.acd] = dg.DwordT(s32)
		// FIXME - handle overflow and carry

	case "WSUB":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		res = int32(cpuPtr.ac[twoAcc1Word.acd]) - int32(cpuPtr.ac[twoAcc1Word.acs])
		cpuPtr.ac[twoAcc1Word.acd] = dg.DwordT(res)
		// FIXME - handle overflow and carry

	case "ZEX":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		cpuPtr.ac[twoAcc1Word.acd] = 0 | dg.DwordT(util.DWordGetLowerWord(cpuPtr.ac[twoAcc1Word.acs]))

	default:
		log.Fatalf("ERROR: EAGLE_OP instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg.PhysAddrT(iPtr.instrLength)
	return true
}
