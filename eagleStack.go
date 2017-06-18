package main

import "log"

func eagleStack(cpuPtr *Cpu, iPtr *DecodedInstr) bool {

	var wfpSav, dwd dg_dword

	switch iPtr.mnemonic {

	case "STAFP":
		// FIXME handle segments
		memWriteDWord(WFP_LOC, cpuPtr.ac[iPtr.acd])

	case "STASB":
		// FIXME handle segments
		memWriteDWord(WSB_LOC, cpuPtr.ac[iPtr.acd])

	case "STASL":
		// FIXME handle segments
		memWriteDWord(WSL_LOC, cpuPtr.ac[iPtr.acd])

	case "STASP":
		// FIXME handle segments
		memWriteDWord(WSP_LOC, cpuPtr.ac[iPtr.acd])

	case "STATS":
		// FIXME handle segments
		memWriteDWord(dg_phys_addr(memReadDWord(WSL_LOC)), cpuPtr.ac[iPtr.acd])

	case "WSAVR":
		wfpSav = memReadDWord(WFP_LOC)
		wsPush(0, cpuPtr.ac[0]) // 1
		wsPush(0, cpuPtr.ac[1]) // 2
		wsPush(0, cpuPtr.ac[2]) // 3
		wsPush(0, wfpSav)       // 4
		dwd = cpuPtr.ac[3] & 0x7fffffff
		if cpuPtr.carry {
			dwd |= 0x80000000
		}
		wsPush(0, dwd) // 5
		dwdCnt := int(iPtr.imm16b)
		if dwdCnt > 0 {
			for d := 0; d < dwdCnt; d++ {
				wsPush(0, 0)
			}
		}
		cpu.ovk = false
		memWriteDWord(WFP_LOC, memReadDWord(WSP_LOC))
		cpuPtr.ac[3] = memReadDWord(WSP_LOC)

	default:
		log.Fatalf("ERROR: EAGLE_STACK instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
