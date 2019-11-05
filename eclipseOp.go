// mvemg project eclipseOp.go

// Copyright (C) 2017,2019  Steve Merrony

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

	"github.com/SMerrony/dgemug/dg"
	"github.com/SMerrony/dgemug/logging"
	"github.com/SMerrony/dgemug/memory"
)

func eclipseOp(cpuPtr *CPUT, iPtr *decodedInstrT) bool {

	switch iPtr.ix {

	case instrADDI:
		oneAccImm2Word := iPtr.variant.(oneAccImm2WordT)
		// signed 16-bit add immediate
		s16 := int16(memory.DwordGetLowerWord(cpuPtr.ac[oneAccImm2Word.acd]))
		s16 += oneAccImm2Word.immS16
		cpuPtr.ac[oneAccImm2Word.acd] = dg.DwordT(s16) & 0X0000FFFF

	case instrANDI:
		oneAccImmWd2Word := iPtr.variant.(oneAccImmWd2WordT)
		wd := memory.DwordGetLowerWord(cpuPtr.ac[oneAccImmWd2Word.acd])
		cpuPtr.ac[oneAccImmWd2Word.acd] = dg.DwordT(wd&oneAccImmWd2Word.immWord) & 0x0000ffff

	case instrADI: // 16-bit unsigned Add Immediate
		immOneAcc := iPtr.variant.(immOneAccT)
		wd := memory.DwordGetLowerWord(cpuPtr.ac[immOneAcc.acd])
		wd += dg.WordT(immOneAcc.immU16) // unsigned arithmetic does wraparound in Go
		cpuPtr.ac[immOneAcc.acd] = dg.DwordT(wd)

	case instrBTO:
		// TODO Handle segment and indirection...
		twoAcc1Word := iPtr.variant.(twoAcc1WordT)
		addr, bitNum := resolveEclipseBitAddr(cpuPtr, &twoAcc1Word)
		wd := memory.ReadWord(addr)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... BTO Addr: %d, Bit: %d, Before: %s\n",
				addr, bitNum, memory.WordToBinStr(wd))
		}
		memory.SetWbit(&wd, bitNum)
		memory.WriteWord(addr, wd)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... BTO                     Result: %s\n", memory.WordToBinStr(wd))
		}

	case instrBTZ:
		// TODO Handle segment and indirection...
		twoAcc1Word := iPtr.variant.(twoAcc1WordT)
		addr, bitNum := resolveEclipseBitAddr(cpuPtr, &twoAcc1Word)
		wd := memory.ReadWord(addr)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... BTZ Addr: %d, Bit: %d, Before: %s\n", addr, bitNum, memory.WordToBinStr(wd))
		}
		memory.ClearWbit(&wd, bitNum)
		memory.WriteWord(addr, wd)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... BTZ                     Result: %s\n",
				memory.WordToBinStr(wd))
		}

	case instrDHXL:
		immOneAcc := iPtr.variant.(immOneAccT)
		dplus1 := immOneAcc.acd + 1
		if dplus1 == 4 {
			dplus1 = 0
		}
		dwd := memory.DwordFromTwoWords(memory.DwordGetLowerWord(cpuPtr.ac[immOneAcc.acd]), memory.DwordGetLowerWord(cpuPtr.ac[dplus1]))
		dwd <<= (immOneAcc.immU16 * 4)
		cpuPtr.ac[immOneAcc.acd] = dg.DwordT(memory.DwordGetUpperWord(dwd))
		cpuPtr.ac[dplus1] = dg.DwordT(memory.DwordGetLowerWord(dwd))

	case instrDLSH:
		twoAcc1Word := iPtr.variant.(twoAcc1WordT)
		dplus1 := twoAcc1Word.acd + 1
		if dplus1 == 4 {
			dplus1 = 0
		}
		dwd := dlsh(cpuPtr.ac[twoAcc1Word.acs], cpuPtr.ac[twoAcc1Word.acd], cpuPtr.ac[dplus1])
		cpuPtr.ac[twoAcc1Word.acd] = dg.DwordT(memory.DwordGetUpperWord(dwd))
		cpuPtr.ac[dplus1] = dg.DwordT(memory.DwordGetLowerWord(dwd))

	case instrELEF:
		oneAccModeInt2Word := iPtr.variant.(oneAccModeInd2WordT)
		cpuPtr.ac[oneAccModeInt2Word.acd] = dg.DwordT(resolve16bitEffAddr(cpuPtr, oneAccModeInt2Word.ind, oneAccModeInt2Word.mode, oneAccModeInt2Word.disp15, iPtr.dispOffset))

	case instrESTA:
		oneAccModeInt2Word := iPtr.variant.(oneAccModeInd2WordT)
		addr := resolve16bitEffAddr(cpuPtr, oneAccModeInt2Word.ind, oneAccModeInt2Word.mode, oneAccModeInt2Word.disp15, iPtr.dispOffset)
		memory.WriteWord(addr, memory.DwordGetLowerWord(cpuPtr.ac[oneAccModeInt2Word.acd]))

	case instrHXL:
		immOneAcc := iPtr.variant.(immOneAccT)
		dwd := cpuPtr.ac[immOneAcc.acd] << (uint32(immOneAcc.immU16) * 4)
		cpuPtr.ac[immOneAcc.acd] = dwd & 0x0ffff

	case instrHXR:
		immOneAcc := iPtr.variant.(immOneAccT)
		dwd := cpuPtr.ac[immOneAcc.acd] >> (uint32(immOneAcc.immU16) * 4)
		cpuPtr.ac[immOneAcc.acd] = dwd & 0x0ffff

	case instrIOR:
		twoAcc1Word := iPtr.variant.(twoAcc1WordT)
		wd := memory.DwordGetLowerWord(cpuPtr.ac[twoAcc1Word.acd]) | memory.DwordGetLowerWord(cpuPtr.ac[twoAcc1Word.acs])
		cpuPtr.ac[twoAcc1Word.acd] = dg.DwordT(wd)

	case instrIORI:
		oneAccImmWd2Word := iPtr.variant.(oneAccImmWd2WordT)
		wd := memory.DwordGetLowerWord(cpuPtr.ac[oneAccImmWd2Word.acd]) | oneAccImmWd2Word.immWord
		cpuPtr.ac[oneAccImmWd2Word.acd] = dg.DwordT(wd)

	case instrLDB:
		twoAcc1Word := iPtr.variant.(twoAcc1WordT)
		cpuPtr.ac[twoAcc1Word.acd] = dg.DwordT(memory.ReadByteEclipseBA(memory.DwordGetLowerWord(cpuPtr.ac[twoAcc1Word.acs])))

	case instrLSH:
		twoAcc1Word := iPtr.variant.(twoAcc1WordT)
		cpuPtr.ac[twoAcc1Word.acd] = lsh(cpuPtr.ac[twoAcc1Word.acs], cpuPtr.ac[twoAcc1Word.acd])

	case instrSBI: // unsigned
		immOneAcc := iPtr.variant.(immOneAccT)
		wd := memory.DwordGetLowerWord(cpuPtr.ac[immOneAcc.acd])
		if immOneAcc.immU16 < 1 || immOneAcc.immU16 > 4 {
			log.Fatal("Invalid immediate value in SBI")
		}
		wd -= dg.WordT(immOneAcc.immU16)
		cpuPtr.ac[immOneAcc.acd] = dg.DwordT(wd)

	case instrSTB:
		twoAcc1Word := iPtr.variant.(twoAcc1WordT)
		hiLo := memory.TestDwbit(cpuPtr.ac[twoAcc1Word.acs], 31)
		addr := dg.PhysAddrT(memory.DwordGetLowerWord(cpuPtr.ac[twoAcc1Word.acs])) >> 1
		byt := dg.ByteT(cpuPtr.ac[twoAcc1Word.acd])
		memory.WriteByte(addr, hiLo, byt)

	case instrXCH:
		twoAcc1Word := iPtr.variant.(twoAcc1WordT)
		dwd := cpuPtr.ac[twoAcc1Word.acs]
		cpuPtr.ac[twoAcc1Word.acs] = cpuPtr.ac[twoAcc1Word.acd] & 0x0ffff
		cpuPtr.ac[twoAcc1Word.acd] = dwd & 0x0ffff

	default:
		log.Fatalf("ERROR: ECLIPSE_OP instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg.PhysAddrT(iPtr.instrLength)
	return true
}

func dlsh(acS, acDh, acDl dg.DwordT) dg.DwordT {
	var shft = int8(acS)
	var dwd = memory.DwordFromTwoWords(memory.DwordGetLowerWord(acDh), memory.DwordGetLowerWord(acDl))
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
	var wd = memory.DwordGetLowerWord(acD)
	if shft == 0 {
		wd = memory.DwordGetLowerWord(acD) // do nothing
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
