// eclipsePC.go
package main

import (
	"log"
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
	"mvemg/util"
)

func eclipsePC(cpuPtr *CPU, iPtr *decodedInstrT) bool {
	var (
		addr, inc         dg.PhysAddrT
		acd, acs, h, l    int16
		wd                dg.WordT
		bit               uint
		noAccModeInd2Word noAccModeInd2WordT
	)

	switch iPtr.mnemonic {

	case "CLM": // signed compare to limits
		acs = int16(util.DWordGetLowerWord(cpuPtr.ac[iPtr.acs]))
		if iPtr.acs == iPtr.acd {
			l = int16(memory.ReadWord(cpuPtr.pc + 1))
			h = int16(memory.ReadWord(cpuPtr.pc + 2))
			if acs < l || acs > h {
				inc = 3
			} else {
				inc = 4
			}
		} else {
			l = int16(memory.ReadWord(dg.PhysAddrT(util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd]))))
			h = int16(memory.ReadWord(dg.PhysAddrT(util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd]) + 1)))
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
		tableStart := resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)
		offset := util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])
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
		acs = int16(util.DWordGetLowerWord(cpuPtr.ac[iPtr.acs]))
		acd = int16(util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd]))
		if acs > acd {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "SNB":
		addr, bit = resolveEclipseBitAddr(cpuPtr, iPtr)
		wd := memory.ReadWord(addr)
		if util.TestWbit(wd, int(bit)) {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "SNB: Wd Addr: %d., word: %0X, bit #: %d\n", addr, wd, bit)
		}

	case "SZB":
		addr, bit = resolveEclipseBitAddr(cpuPtr, iPtr)
		wd := memory.ReadWord(addr)
		if !util.TestWbit(wd, int(bit)) {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
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
