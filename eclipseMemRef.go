// eclipseMemRef.go
package main

import (
	"fmt"
	"log"
)

func eclipseMemRef(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	var (
		addr dg_phys_addr
	)

	switch iPtr.mnemonic {

	case "BLM":
		/* AC0 - unused, AC1 - no. wds to move, AC2 - src, AC3 - dest */
		numWds := dwordGetLowerWord(cpuPtr.ac[1])
		if numWds == 0 {
			debugPrint(SYSTEM_LOG, "BLM called with AC1 == 0, not moving anything\n")
			break
		}
		src := dwordGetLowerWord(cpuPtr.ac[2])
		dest := dwordGetLowerWord(cpuPtr.ac[3])
		debugPrint(SYSTEM_LOG, fmt.Sprintf("BLM moving %d words from %d to %d\n", numWds, src, dest))
		for numWds != 0 {
			memWriteWord(dg_phys_addr(dest), memReadWord(dg_phys_addr(src)))
			numWds--
			src++
			dest++
		}
		cpuPtr.ac[1] = 0
		cpuPtr.ac[2] = dg_dword(src) // TODO confirm this is right, doc ambiguous
		cpuPtr.ac[3] = dg_dword(dest)

	case "ELDA":
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		cpuPtr.ac[iPtr.acd] = dg_dword(addr) & 0x0ffff

	default:
		log.Printf("ERROR: EAGLE_MEMREF instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
