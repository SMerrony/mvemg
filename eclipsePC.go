// eclipsePC.go
package main

import (
	"log"
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
	"mvemg/util"
)

func eclipsePC(cpuPtr *CPUT, iPtr *decodedInstrT) bool {
	var (
		addr, inc          dg.PhysAddrT
		acd, acs, h, l     int16
		wd                 dg.WordT
		bit                uint
		noAccModeInd2Word  noAccModeInd2WordT
		oneAccModeInt2Word oneAccModeInd2WordT
		twoAcc1Word        twoAcc1WordT
	)

	switch iPtr.mnemonic {

	case "CLM": // signed compare to limits
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		acs = int16(util.DWordGetLowerWord(cpuPtr.ac[twoAcc1Word.acs]))
		if twoAcc1Word.acs == twoAcc1Word.acd {
			l = int16(memory.ReadWord(cpuPtr.pc + 1))
			h = int16(memory.ReadWord(cpuPtr.pc + 2))
			if acs < l || acs > h {
				inc = 3
			} else {
				inc = 4
			}
		} else {
			l = int16(memory.ReadWord(dg.PhysAddrT(util.DWordGetLowerWord(cpuPtr.ac[twoAcc1Word.acd]))))
			h = int16(memory.ReadWord(dg.PhysAddrT(util.DWordGetLowerWord(cpuPtr.ac[twoAcc1Word.acd]) + 1)))
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

	case "DSPA":
		oneAccModeInt2Word = iPtr.variant.(oneAccModeInd2WordT)
		tableStart := resolve16bitEclipseAddr(cpuPtr, oneAccModeInt2Word.ind, oneAccModeInt2Word.mode, oneAccModeInt2Word.disp15)
		offset := util.DWordGetLowerWord(cpuPtr.ac[oneAccModeInt2Word.acd])
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
		addr = dg.PhysAddrT(memory.ReadWord(entry))
		if addr == 0xffffffff {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc = addr
		}

	case "EISZ":
		noAccModeInd2Word = iPtr.variant.(noAccModeInd2WordT)
		addr = resolve16bitEclipseAddr(cpuPtr, noAccModeInd2Word.ind, noAccModeInd2Word.mode, noAccModeInd2Word.disp15)
		wd = memory.ReadWord(addr)
		wd++
		memory.WriteWord(addr, wd)
		if wd == 0 {
			cpuPtr.pc += 3
		} else {
			cpuPtr.pc += 2
		}

	case "EJMP":
		noAccModeInd2Word = iPtr.variant.(noAccModeInd2WordT)
		addr = resolve16bitEclipseAddr(cpuPtr, noAccModeInd2Word.ind, noAccModeInd2Word.mode, noAccModeInd2Word.disp15)
		cpuPtr.pc = addr

	case "EJSR":
		noAccModeInd2Word = iPtr.variant.(noAccModeInd2WordT)
		cpuPtr.ac[3] = dg.DwordT(cpuPtr.pc) + 2
		addr = resolve16bitEclipseAddr(cpuPtr, noAccModeInd2Word.ind, noAccModeInd2Word.mode, noAccModeInd2Word.disp15)
		cpuPtr.pc = addr

	case "SGT": //16-bit signed numbers
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		acs = int16(util.DWordGetLowerWord(cpuPtr.ac[twoAcc1Word.acs]))
		acd = int16(util.DWordGetLowerWord(cpuPtr.ac[twoAcc1Word.acd]))
		if acs > acd {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc++
		}

	case "SNB":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		addr, bit = resolveEclipseBitAddr(cpuPtr, &twoAcc1Word)
		wd := memory.ReadWord(addr)
		if util.TestWbit(wd, int(bit)) {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc++
		}
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "SNB: Wd Addr: %d., word: %0X, bit #: %d\n", addr, wd, bit)
		}

	case "SZB":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		addr, bit = resolveEclipseBitAddr(cpuPtr, &twoAcc1Word)
		wd := memory.ReadWord(addr)
		if !util.TestWbit(wd, int(bit)) {
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
