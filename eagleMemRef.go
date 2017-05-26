// eagleMemRef.go
package main

import (
	"log"
)

func eagleMemRef(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	var (
		addr dg_phys_addr
		wd   dg_word
		dwd  dg_dword
		i32  int32
	)

	switch iPtr.mnemonic {

	case "WBLM":
		/* AC0 - unused, AC1 - no. wds to move (if neg then descending order), AC2 - src, AC3 - dest */
		numWds := int32(cpuPtr.ac[1])
		var order int32 = 1
		if numWds < 0 {
			order = -1
		}
		if numWds == 0 {
			log.Println("INFO: WBLM called with AC1 == 0, not moving anything")
		} else {
			src := dg_phys_addr(cpuPtr.ac[2])
			dest := dg_phys_addr(cpuPtr.ac[3])
			log.Printf("DEBUG: WBLM moving %d words from %d to %d\n", numWds, src, dest)
			for numWds != 0 {
				memWriteWord(dest, memReadWord(src))
				numWds -= order
				if order == 1 {
					src++
					dest++
				} else {
					src--
					dest--
				}
			}
			cpuPtr.ac[1] = 0
			cpuPtr.ac[2] = dg_dword(dest) // TODO confirm this
			cpuPtr.ac[3] = dg_dword(dest)
		}
	case "XLDB":
		cpuPtr.ac[iPtr.acd] = dg_dword(memReadByte(resolve16bitEagleAddr(cpuPtr, ' ', iPtr.mode, iPtr.disp), iPtr.loHiBit)) & 0x00ff

	case "XLEF":
		cpuPtr.ac[iPtr.acd] = dg_dword(resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp))

	case "XNLDA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		wd = memReadWord(addr)
		cpuPtr.ac[iPtr.acd] = sexWordToDWord(wd) // FIXME check this...

	case "XNSTA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		memWriteWord(addr, wd)

	case "XWADI":
		// add 1-4 to signed 32-bit acc
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		i32 = int32(memReadDWord(addr)) + iPtr.immVal
		// FIXME handle Carry and OVeRflow
		memWriteDWord(addr, dg_dword(i32))

	case "XWLDA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		dwd = memReadDWord(addr)
		cpuPtr.ac[iPtr.acd] = dwd

	case "XWSTA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		dwd = cpuPtr.ac[iPtr.acd]
		memWriteDWord(addr, dwd)

	default:
		log.Printf("ERROR: EAGLE_MEMREF instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
