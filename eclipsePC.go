// eclipsePC.go
package main

import (
	"log"
	"mvemg/logging"
)

func eclipsePC(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	var (
		addr, inc      dg_phys_addr
		acd, acs, h, l int16
		wd             dg_word
	//dwd dg_dword
	)

	switch iPtr.mnemonic {

	case "CLM": // signed compare to limits
		acs = int16(dwordGetLowerWord(cpuPtr.ac[iPtr.acs]))
		if iPtr.acs == iPtr.acd {
			l = int16(memReadWord(cpuPtr.pc + 1))
			h = int16(memReadWord(cpuPtr.pc + 2))
			if acs < l || acs > h {
				inc = 3
			} else {
				inc = 4
			}
		} else {
			l = int16(memReadWord(dg_phys_addr(dwordGetLowerWord(cpuPtr.ac[iPtr.acd]))))
			h = int16(memReadWord(dg_phys_addr(dwordGetLowerWord(cpuPtr.ac[iPtr.acd]) + 1)))
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
		tableStart := resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		offset := dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		lowLimit := memReadWord(tableStart - 2)
		hiLimit := memReadWord(tableStart - 1)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "DSPA called with table at %d, offset %d, lo %d hi %d\n",
				tableStart, offset, lowLimit, hiLimit)
		}
		if offset < lowLimit || offset > hiLimit {
			log.Fatalf("ERROR: DPSA called with out of bounds offset %d", offset)
		}
		entry := tableStart - dg_phys_addr(lowLimit) + dg_phys_addr(offset)
		addr = dg_phys_addr(memReadWord(entry))
		if addr == 0xffffffff {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc = addr
		}

	case "EISZ":
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		wd = memReadWord(addr)
		wd++
		memWriteWord(addr, wd)
		if wd == 0 {
			cpuPtr.pc += 3
		} else {
			cpuPtr.pc += 2
		}

	case "EJMP":
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		cpuPtr.pc = addr

	case "EJSR":
		cpuPtr.ac[3] = dg_dword(cpuPtr.pc) + 2
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		cpuPtr.pc = addr

	case "SGT": //16-bit signed numbers
		acs = int16(dwordGetLowerWord(cpuPtr.ac[iPtr.acs]))
		acd = int16(dwordGetLowerWord(cpuPtr.ac[iPtr.acd]))
		if acs > acd {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "SNB":
		// resolve an ECLIPSE bit address
		if iPtr.acd == iPtr.acs {
			addr = 0
		} else {
			eff := dg_word(cpuPtr.ac[iPtr.acs])
			for testWbit(eff, 0) {
				eff = memReadWord(dg_phys_addr(eff))
			}
			addr = dg_phys_addr(eff)
		}
		addr16 := cpuPtr.ac[iPtr.acd] >> 4
		addr += dg_phys_addr(addr16)
		wd := memReadWord(addr)
		bit := int(cpuPtr.ac[iPtr.acd] & 0x000f)
		if testWbit(wd, bit) {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "SNB: Wd Addr: %d., word: %0X, bit #: %d\n", addr, wd, bit)
		}

	case "SZB":
		// resolve an ECLIPSE bit address
		if iPtr.acd == iPtr.acs {
			addr = 0
		} else {
			eff := dg_word(cpuPtr.ac[iPtr.acs])
			for testWbit(eff, 0) {
				eff = memReadWord(dg_phys_addr(eff))
			}
			addr = dg_phys_addr(eff)
		}
		addr16 := cpuPtr.ac[iPtr.acd] >> 4
		addr += dg_phys_addr(addr16)
		wd := memReadWord(addr)
		bit := int(cpuPtr.ac[iPtr.acd] & 0x000f)
		if !testWbit(wd, bit) {
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
