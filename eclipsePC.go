// eclipsePC.go

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

func eclipsePC(cpuPtr *CPUT, iPtr *decodedInstrT) bool {

	switch iPtr.ix {

	case instrCLM: // signed compare to limits
		var (
			l, h int16
			inc  dg.PhysAddrT
		)
		twoAcc1Word := iPtr.variant.(twoAcc1WordT)
		acs := int16(memory.DwordGetLowerWord(cpuPtr.ac[twoAcc1Word.acs]))
		if twoAcc1Word.acs == twoAcc1Word.acd {
			l = int16(memory.ReadWord(cpuPtr.pc + 1))
			h = int16(memory.ReadWord(cpuPtr.pc + 2))
			if acs < l || acs > h {
				inc = 3
			} else {
				inc = 4
			}
		} else {
			l = int16(memory.ReadWord(dg.PhysAddrT(memory.DwordGetLowerWord(cpuPtr.ac[twoAcc1Word.acd]))))
			h = int16(memory.ReadWord(dg.PhysAddrT(memory.DwordGetLowerWord(cpuPtr.ac[twoAcc1Word.acd]) + 1)))
			if acs < l || acs > h {
				inc = 1
			} else {
				inc = 2
			}
		}
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "CLM compared %d with limits %d and %d, moving PC by %d\n", acs, l, h, inc)
		}
		cpuPtr.pc += inc

	case instrDSPA:
		oneAccModeInt2Word := iPtr.variant.(oneAccModeInd2WordT)
		tableStart := resolve16bitEffAddr(cpuPtr, oneAccModeInt2Word.ind, oneAccModeInt2Word.mode, oneAccModeInt2Word.disp15, iPtr.dispOffset)
		offset := memory.DwordGetLowerWord(cpuPtr.ac[oneAccModeInt2Word.acd])
		lowLimit := memory.ReadWord(tableStart - 2)
		hiLimit := memory.ReadWord(tableStart - 1)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "DSPA called with table at %d, offset %d, lo %d hi %d\n",
				tableStart, offset, lowLimit, hiLimit)
		}
		if offset < lowLimit || offset > hiLimit {
			log.Fatalf("ERROR: DPSA called with out of bounds offset %d", offset)
		}
		entry := tableStart - dg.PhysAddrT(lowLimit) + dg.PhysAddrT(offset)
		addr := dg.PhysAddrT(memory.ReadWord(entry))
		if addr == 0xffffffff {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc = addr
		}

	case instrEISZ:
		noAccModeInd2Word := iPtr.variant.(noAccModeInd2WordT)
		addr := resolve15bitDisplacement(cpuPtr, noAccModeInd2Word.ind, noAccModeInd2Word.mode, noAccModeInd2Word.disp15, iPtr.dispOffset) & 0x7fff
		wd := memory.ReadWord(addr)
		wd++
		memory.WriteWord(addr, wd)
		if wd == 0 {
			cpuPtr.pc += 3
		} else {
			cpuPtr.pc += 2
		}

	case instrEJMP:
		noAccModeInd2Word := iPtr.variant.(noAccModeInd2WordT)
		addr := resolve15bitDisplacement(cpuPtr, noAccModeInd2Word.ind, noAccModeInd2Word.mode, noAccModeInd2Word.disp15, iPtr.dispOffset)
		cpuPtr.pc = addr & 0x7fff

	case instrEJSR:
		noAccModeInd2Word := iPtr.variant.(noAccModeInd2WordT)
		cpuPtr.ac[3] = dg.DwordT(cpuPtr.pc) + 2
		addr := resolve15bitDisplacement(cpuPtr, noAccModeInd2Word.ind, noAccModeInd2Word.mode, noAccModeInd2Word.disp15, iPtr.dispOffset)
		cpuPtr.pc = addr & 0x7fff

	case instrFNS:
		cpuPtr.pc++

	case instrSGT: //16-bit signed numbers
		twoAcc1Word := iPtr.variant.(twoAcc1WordT)
		acs := int16(memory.DwordGetLowerWord(cpuPtr.ac[twoAcc1Word.acs]))
		acd := int16(memory.DwordGetLowerWord(cpuPtr.ac[twoAcc1Word.acd]))
		if acs > acd {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc++
		}

	case instrSNB:
		twoAcc1Word := iPtr.variant.(twoAcc1WordT)
		addr, bit := resolveEclipseBitAddr(cpuPtr, &twoAcc1Word)
		wd := memory.ReadWord(addr)
		if memory.TestWbit(wd, int(bit)) {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc++
		}
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "SNB: Wd Addr: %d., word: %0X, bit #: %d\n", addr, wd, bit)
		}

	case instrSZB:
		twoAcc1Word := iPtr.variant.(twoAcc1WordT)
		addr, bit := resolveEclipseBitAddr(cpuPtr, &twoAcc1Word)
		wd := memory.ReadWord(addr)
		if !memory.TestWbit(wd, int(bit)) {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc++
		}
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "SZB: Wd Addr: %d., word: %0X, bit #: %d\n", addr, wd, bit)
		}

	default:
		log.Fatalf("ERROR: ECLIPSE_PC instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	return true
}
