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
	"mvemg/logging"
)

func eclipseOp(cpuPtr *CPU, iPtr *decodedInstrT) bool {
	var (
		addr   DgPhysAddrT
		byt    DgByteT
		wd     DgWordT
		dwd    DgDwordT
		bitNum uint
	)

	switch iPtr.mnemonic {

	case "ADI": // 16-bit unsigned Add Immediate
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		wd += DgWordT(iPtr.immU16) // unsigned arithmetic does wraparound in Go
		cpuPtr.ac[iPtr.acd] = DgDwordT(wd)

	case "BTO":
		// TODO Handle segment and indirection...
		addr, bitNum = resolveEclipseBitAddr(cpuPtr, iPtr)
		wd = memReadWord(addr)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... BTO Addr: %d, Bit: %d, Before: %s\n",
				addr, bitNum, wordToBinStr(wd))
		}
		wd = SetWbit(wd, bitNum)
		memWriteWord(addr, wd)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... BTO                     Result: %s\n", wordToBinStr(wd))
		}

	case "BTZ":
		// TODO Handle segment and indirection...
		addr, bitNum = resolveEclipseBitAddr(cpuPtr, iPtr)
		wd = memReadWord(addr)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... BTZ Addr: %d, Bit: %d, Before: %s\n", addr, bitNum, wordToBinStr(wd))
		}
		wd = ClearWbit(wd, bitNum)
		memWriteWord(addr, wd)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... BTZ                     Result: %s\n",
				wordToBinStr(wd))
		}

	case "DIV": // unsigned divide
		uw := dwordGetLowerWord(cpuPtr.ac[0])
		lw := dwordGetLowerWord(cpuPtr.ac[1])
		dwd = dwordFromTwoWords(uw, lw)
		quot := dwordGetLowerWord(cpuPtr.ac[2])
		if uw > quot || quot == 0 {
			cpuPtr.carry = true
		} else {
			cpuPtr.carry = false
			cpuPtr.ac[0] = (dwd % DgDwordT(quot)) & 0x0ffff
			cpuPtr.ac[1] = (dwd / DgDwordT(quot)) & 0x0ffff
		}

	case "DLSH":
		dplus1 := iPtr.acd + 1
		if dplus1 == 4 {
			dplus1 = 0
		}
		dwd = dlsh(cpuPtr.ac[iPtr.acs], cpuPtr.ac[iPtr.acd], cpuPtr.ac[dplus1])
		cpuPtr.ac[iPtr.acd] = DgDwordT(dwordGetUpperWord(dwd))
		cpuPtr.ac[dplus1] = DgDwordT(dwordGetLowerWord(dwd))

	case "ELEF":
		cpuPtr.ac[iPtr.acd] = DgDwordT(resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15))

	case "ESTA":
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)
		memWriteWord(addr, dwordGetLowerWord(cpuPtr.ac[iPtr.acd]))

	case "HXL":
		dwd = cpuPtr.ac[iPtr.acd] << (uint32(iPtr.immU16) * 4)
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	case "HXR":
		dwd = cpuPtr.ac[iPtr.acd] >> (uint32(iPtr.immU16) * 4)
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	case "IOR":
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd]) | dwordGetLowerWord(cpuPtr.ac[iPtr.acs])
		cpuPtr.ac[iPtr.acd] = DgDwordT(wd)

	case "IORI":
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd]) | iPtr.immWord
		cpuPtr.ac[iPtr.acd] = DgDwordT(wd)

	case "LDB":
		byt = memReadByteEclipseBA(dwordGetLowerWord(cpuPtr.ac[iPtr.acs]))
		cpuPtr.ac[iPtr.acd] = DgDwordT(byt)

	case "LSH":
		cpuPtr.ac[iPtr.acd] = lsh(cpuPtr.ac[iPtr.acs], cpuPtr.ac[iPtr.acd])

	case "MUL": // unsigned 16-bit multiply with add: (AC1 * AC2) + AC0 => AC0(h) and AC1(l)
		ac0 := dwordGetLowerWord(cpuPtr.ac[0])
		ac1 := dwordGetLowerWord(cpuPtr.ac[1])
		ac2 := dwordGetLowerWord(cpuPtr.ac[2])
		dwd := (DgDwordT(ac1) * DgDwordT(ac2)) + DgDwordT(ac0)
		cpuPtr.ac[0] = DgDwordT(dwordGetUpperWord(dwd))
		cpuPtr.ac[1] = DgDwordT(dwordGetLowerWord(dwd))

	case "SBI": // unsigned
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if iPtr.immU16 < 1 || iPtr.immU16 > 4 {
			log.Fatal("Invalid immediate value in SBI")
		}
		wd -= DgWordT(iPtr.immU16)
		cpuPtr.ac[iPtr.acd] = DgDwordT(wd)

	case "STB":
		hiLo := testDWbit(cpuPtr.ac[iPtr.acs], 31)
		addr = DgPhysAddrT(dwordGetLowerWord(cpuPtr.ac[iPtr.acs])) >> 1
		byt = DgByteT(cpuPtr.ac[iPtr.acd])
		memWriteByte(addr, hiLo, byt)

	case "XCH":
		dwd = cpuPtr.ac[iPtr.acs]
		cpuPtr.ac[iPtr.acs] = cpuPtr.ac[iPtr.acd] & 0x0ffff
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	default:
		log.Fatalf("ERROR: ECLIPSE_OP instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += DgPhysAddrT(iPtr.instrLength)
	return true
}

func dlsh(acS, acDh, acDl DgDwordT) DgDwordT {
	var shft = int8(acS)
	var dwd = dwordFromTwoWords(dwordGetLowerWord(acDh), dwordGetLowerWord(acDl))
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

func lsh(acS, acD DgDwordT) DgDwordT {
	var shft = int8(acS)
	var wd = dwordGetLowerWord(acD)
	if shft == 0 {
		wd = dwordGetLowerWord(acD) // do nothing
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
	return DgDwordT(wd)
}
