// mvemg project eclipseOp.go

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

func eclipseOp(cpuPtr *CPU, iPtr *decodedInstrT) bool {
	var (
		addr   dg.PhysAddrT
		byt    dg.ByteT
		wd     dg.WordT
		dwd    dg.DwordT
		bitNum uint
	)

	switch iPtr.mnemonic {

	case "ADI": // 16-bit unsigned Add Immediate
		wd = util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])
		wd += dg.WordT(iPtr.immU16) // unsigned arithmetic does wraparound in Go
		cpuPtr.ac[iPtr.acd] = dg.DwordT(wd)

	case "BTO":
		// TODO Handle segment and indirection...
		addr, bitNum = resolveEclipseBitAddr(cpuPtr, iPtr)
		wd = memory.ReadWord(addr)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... BTO Addr: %d, Bit: %d, Before: %s\n",
				addr, bitNum, util.WordToBinStr(wd))
		}
		wd = util.SetWbit(wd, bitNum)
		memory.WriteWord(addr, wd)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... BTO                     Result: %s\n", util.WordToBinStr(wd))
		}

	case "BTZ":
		// TODO Handle segment and indirection...
		addr, bitNum = resolveEclipseBitAddr(cpuPtr, iPtr)
		wd = memory.ReadWord(addr)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... BTZ Addr: %d, Bit: %d, Before: %s\n", addr, bitNum, util.WordToBinStr(wd))
		}
		wd = util.ClearWbit(wd, bitNum)
		memory.WriteWord(addr, wd)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... BTZ                     Result: %s\n",
				util.WordToBinStr(wd))
		}

	case "DIV": // unsigned divide
		uw := util.DWordGetLowerWord(cpuPtr.ac[0])
		lw := util.DWordGetLowerWord(cpuPtr.ac[1])
		dwd = util.DWordFromTwoWords(uw, lw)
		quot := util.DWordGetLowerWord(cpuPtr.ac[2])
		if uw > quot || quot == 0 {
			cpuPtr.carry = true
		} else {
			cpuPtr.carry = false
			cpuPtr.ac[0] = (dwd % dg.DwordT(quot)) & 0x0ffff
			cpuPtr.ac[1] = (dwd / dg.DwordT(quot)) & 0x0ffff
		}

	case "DLSH":
		dplus1 := iPtr.acd + 1
		if dplus1 == 4 {
			dplus1 = 0
		}
		dwd = dlsh(cpuPtr.ac[iPtr.acs], cpuPtr.ac[iPtr.acd], cpuPtr.ac[dplus1])
		cpuPtr.ac[iPtr.acd] = dg.DwordT(util.DWordGetUpperWord(dwd))
		cpuPtr.ac[dplus1] = dg.DwordT(util.DWordGetLowerWord(dwd))

	case "ELEF":
		cpuPtr.ac[iPtr.acd] = dg.DwordT(resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15))

	case "ESTA":
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)
		memory.WriteWord(addr, util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd]))

	case "HXL":
		dwd = cpuPtr.ac[iPtr.acd] << (uint32(iPtr.immU16) * 4)
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	case "HXR":
		dwd = cpuPtr.ac[iPtr.acd] >> (uint32(iPtr.immU16) * 4)
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	case "IOR":
		wd = util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd]) | util.DWordGetLowerWord(cpuPtr.ac[iPtr.acs])
		cpuPtr.ac[iPtr.acd] = dg.DwordT(wd)

	case "IORI":
		wd = util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd]) | iPtr.immWord
		cpuPtr.ac[iPtr.acd] = dg.DwordT(wd)

	case "LDB":
		byt = memory.ReadByteEclipseBA(util.DWordGetLowerWord(cpuPtr.ac[iPtr.acs]))
		cpuPtr.ac[iPtr.acd] = dg.DwordT(byt)

	case "LSH":
		cpuPtr.ac[iPtr.acd] = lsh(cpuPtr.ac[iPtr.acs], cpuPtr.ac[iPtr.acd])

	case "MUL": // unsigned 16-bit multiply with add: (AC1 * AC2) + AC0 => AC0(h) and AC1(l)
		ac0 := util.DWordGetLowerWord(cpuPtr.ac[0])
		ac1 := util.DWordGetLowerWord(cpuPtr.ac[1])
		ac2 := util.DWordGetLowerWord(cpuPtr.ac[2])
		dwd := (dg.DwordT(ac1) * dg.DwordT(ac2)) + dg.DwordT(ac0)
		cpuPtr.ac[0] = dg.DwordT(util.DWordGetUpperWord(dwd))
		cpuPtr.ac[1] = dg.DwordT(util.DWordGetLowerWord(dwd))

	case "SBI": // unsigned
		wd = util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if iPtr.immU16 < 1 || iPtr.immU16 > 4 {
			log.Fatal("Invalid immediate value in SBI")
		}
		wd -= dg.WordT(iPtr.immU16)
		cpuPtr.ac[iPtr.acd] = dg.DwordT(wd)

	case "STB":
		hiLo := util.TestDWbit(cpuPtr.ac[iPtr.acs], 31)
		addr = dg.PhysAddrT(util.DWordGetLowerWord(cpuPtr.ac[iPtr.acs])) >> 1
		byt = dg.ByteT(cpuPtr.ac[iPtr.acd])
		memory.WriteByte(addr, hiLo, byt)

	case "XCH":
		dwd = cpuPtr.ac[iPtr.acs]
		cpuPtr.ac[iPtr.acs] = cpuPtr.ac[iPtr.acd] & 0x0ffff
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	default:
		log.Fatalf("ERROR: ECLIPSE_OP instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg.PhysAddrT(iPtr.instrLength)
	return true
}

func dlsh(acS, acDh, acDl dg.DwordT) dg.DwordT {
	var shft = int8(acS)
	var dwd = util.DWordFromTwoWords(util.DWordGetLowerWord(acDh), util.DWordGetLowerWord(acDl))
	if shft != 0 {
		if shft < -31 || shft > 31 {
			dwd = 0
		} else {
			if shft > 0 {
				dwd >>= uint8(shft)
			} else {
				shft *= -1
				dwd >>= uint8(shft)
			}
		}
	}
	return dwd
}

func lsh(acS, acD dg.DwordT) dg.DwordT {
	var shft = int8(acS)
	var wd = util.DWordGetLowerWord(acD)
	if shft == 0 {
		wd = util.DWordGetLowerWord(acD) // do nothing
	} else {
		if shft < -15 || shft > 15 {
			wd = 0 // 16+ bit shift clears word
		} else {
			if shft > 0 {
				wd >>= uint8(shft)
			} else {
				shft *= -1
				wd >>= uint8(shft)
			}
		}
	}
	return dg.DwordT(wd)
}
