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
)

func eagleOp(cpuPtr *CPU, iPtr *decodedInstrT) bool {
	//var addr dg_phys_addr

	var (
		wd       dg_word
		dwd      dg_dword
		res, s32 int32
		s16      int16
	)

	switch iPtr.mnemonic {

	case "ADDI":
		// signed 16-bit add immediate
		s16 = int16(dwordGetLowerWord(cpuPtr.ac[iPtr.acd]))
		s16 += int16(iPtr.immS16)
		cpuPtr.ac[iPtr.acd] = dg_dword(s16) & 0X0000FFFF

	case "ANDI":
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		cpuPtr.ac[iPtr.acd] = dg_dword(wd&iPtr.immWord) & 0x0000ffff

	case "CRYTC":
		cpuPtr.carry = !cpuPtr.carry

	case "CRYTO":
		cpuPtr.carry = true

	case "CRYTZ":
		cpuPtr.carry = false

	case "LLEF":
		cpuPtr.ac[iPtr.acd] = dg_dword(resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp31))

	case "NADD": // signed add
		s16 = int16(cpuPtr.ac[iPtr.acd]) + int16(cpuPtr.ac[iPtr.acs])
		cpuPtr.ac[iPtr.acd] = dg_dword(s16)

	case "NLDAI":
		cpuPtr.ac[iPtr.acd] = dg_dword(int32(iPtr.immS16))

	case "NSUB": // signed subtract
		s16 = int16(cpuPtr.ac[iPtr.acd]) - int16(cpuPtr.ac[iPtr.acs])
		cpuPtr.ac[iPtr.acd] = dg_dword(s16)

	case "SSPT": /* NO-OP - see p.8-5 of MV/10000 Sys Func Chars */
		log.Println("INFO: SSPT is a No-Op on this machine, continuing")

	case "WADD":
		res = int32(cpuPtr.ac[iPtr.acs]) + int32(cpuPtr.ac[iPtr.acd])
		cpuPtr.ac[iPtr.acd] = dg_dword(res)
		// FIXME - handle overflow and carry

	case "WADC":
		dwd = ^cpuPtr.ac[iPtr.acs]
		cpuPtr.ac[iPtr.acd] = dg_dword(int32(cpuPtr.ac[iPtr.acd]) + int32(dwd))
		// FIXME - handle overflow and carry

	case "WADI":
		s32 = int32(cpuPtr.ac[iPtr.acd]) + int32(iPtr.immU16)
		cpuPtr.ac[iPtr.acd] = dg_dword(s32)

	case "WAND":
		cpuPtr.ac[iPtr.acd] &= cpuPtr.ac[iPtr.acs]

	case "WANDI":
		cpuPtr.ac[iPtr.acd] &= iPtr.immDword

	case "WCOM":
		cpuPtr.ac[iPtr.acd] = ^cpuPtr.ac[iPtr.acs]

	case "WINC":
		cpuPtr.ac[iPtr.acd] = cpuPtr.ac[iPtr.acs] + 1

	case "WIORI":
		cpuPtr.ac[iPtr.acd] |= iPtr.immDword

	case "WLDAI":
		cpuPtr.ac[iPtr.acd] = iPtr.immDword

	case "WLSHI":
		shiftAmt8 := int8(iPtr.immS16 & 0x0ff)
		if shiftAmt8 < 0 { // shift right
			shiftAmt8 *= -1
			dwd = cpuPtr.ac[iPtr.acd] >> uint(shiftAmt8)
			cpuPtr.ac[iPtr.acd] = dwd
		}
		if shiftAmt8 > 0 { // shift left
			dwd = cpuPtr.ac[iPtr.acd] << uint(shiftAmt8)
			cpuPtr.ac[iPtr.acd] = dwd
		}

	case "WLSI":
		cpuPtr.ac[iPtr.acd] = cpuPtr.ac[iPtr.acd] << iPtr.immU16

	case "WMOV":
		cpuPtr.ac[iPtr.acd] = cpuPtr.ac[iPtr.acs]

	case "WMOVR":
		cpuPtr.ac[iPtr.acd] = cpuPtr.ac[iPtr.acd] >> 1

	case "WNADI": //signed 16-bit
		s32 = int32(cpuPtr.ac[iPtr.acd]) + int32(iPtr.immS16)
		cpuPtr.ac[iPtr.acd] = dg_dword(s32)

	case "WNEG":
		// FIXME WNEG - handle CARRY/OVR
		s32 = -int32(cpuPtr.ac[iPtr.acs])
		cpuPtr.ac[iPtr.acd] = dg_dword(s32)

	case "WSBI":
		s32 = int32(cpuPtr.ac[iPtr.acd]) - int32(iPtr.immU16)
		cpuPtr.ac[iPtr.acd] = dg_dword(s32)
		// FIXME - handle overflow and carry

	case "WSUB":
		res = int32(cpuPtr.ac[iPtr.acd]) - int32(cpuPtr.ac[iPtr.acs])
		cpuPtr.ac[iPtr.acd] = dg_dword(res)
		// FIXME - handle overflow and carry

	case "ZEX":
		cpuPtr.ac[iPtr.acd] = 0 | dg_dword(dwordGetLowerWord(cpuPtr.ac[iPtr.acs]))

	default:
		log.Fatalf("ERROR: EAGLE_OP instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
