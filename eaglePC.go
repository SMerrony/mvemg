// eaglePC.go
package main

import (
	"log"
)

func eaglePC(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	//var addr dg_phys_addr
	var (
		wd      dg_word
		tmp32b  dg_dword
		tmpAddr dg_phys_addr
	)

	switch iPtr.mnemonic {

	case "LJMP":
		cpuPtr.pc = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)

	case "LJSR":
		cpuPtr.ac[3] = dg_dword(cpuPtr.pc) + 3
		cpuPtr.pc = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)

	case "LNISZ":
		// unsigned narrow increment and skip if zero
		tmpAddr = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		wd = memReadWord(tmpAddr) + 1
		memWriteWord(tmpAddr, wd)
		if wd == 0 {
			cpuPtr.pc += 4
		} else {
			cpuPtr.pc += 3
		}

	case "LWDSZ":
		// unsigned wide decrement and skip if zero
		tmpAddr = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		tmp32b = memReadDWord(tmpAddr) - 1
		memWriteDWord(tmpAddr, tmp32b)
		if tmp32b == 0 {
			cpuPtr.pc += 4
		} else {
			cpuPtr.pc += 3
		}

	case "WBR":
		//		if iPtr.disp > 0 {
		//			cpuPtr.pc += dg_phys_addr(iPtr.disp)
		//		} else {
		//			cpuPtr.pc -= dg_phys_addr(iPtr.disp)
		//		}
		cpuPtr.pc += dg_phys_addr(iPtr.disp)

	case "WSEQ":
		if iPtr.acd == iPtr.acs {
			tmp32b = 0
		} else {
			tmp32b = cpuPtr.ac[iPtr.acd]
		}
		if cpuPtr.ac[iPtr.acs] == tmp32b {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSEQI":
		tmp32b = sexWordToDWord(iPtr.imm16b)
		if cpuPtr.ac[iPtr.acd] == tmp32b {
			cpuPtr.pc += 3
		} else {
			cpuPtr.pc += 2
		}

	case "WSGE":
		if iPtr.acd == iPtr.acs {
			tmp32b = 0
		} else {
			tmp32b = cpuPtr.ac[iPtr.acd]
		}
		if cpuPtr.ac[iPtr.acs] >= tmp32b {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSGT":
		if iPtr.acd == iPtr.acs {
			tmp32b = 0
		} else {
			tmp32b = cpuPtr.ac[iPtr.acd]
		}
		if cpuPtr.ac[iPtr.acs] > tmp32b {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSKBO":
		if testDWbit(cpuPtr.ac[0], iPtr.bitNum) {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSKBZ":
		if !testDWbit(cpuPtr.ac[0], iPtr.bitNum) {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSLE":
		if iPtr.acd == iPtr.acs {
			tmp32b = 0
		} else {
			tmp32b = cpuPtr.ac[iPtr.acd]
		}
		if cpuPtr.ac[iPtr.acs] <= tmp32b {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSLT":
		if iPtr.acd == iPtr.acs {
			tmp32b = 0
		} else {
			tmp32b = cpuPtr.ac[iPtr.acd]
		}
		if cpuPtr.ac[iPtr.acs] < tmp32b {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "WSNE":
		if iPtr.acd == iPtr.acs {
			tmp32b = 0
		} else {
			tmp32b = cpuPtr.ac[iPtr.acd]
		}
		if cpuPtr.ac[iPtr.acs] != tmp32b {
			cpuPtr.pc += 2
		} else {
			cpuPtr.pc += 1
		}

	case "XJMP":
		cpuPtr.pc = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)

	case "XJSR":
		cpuPtr.ac[3] = dg_dword(cpuPtr.pc + 2) // TODO Check this, PoP is self-contradictory on p.11-642
		cpuPtr.pc = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)

	case "XNISZ":
		tmpAddr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		wd = memReadWord(tmpAddr)
		wd++
		memWriteWord(tmpAddr, wd)
		if wd == 0 {
			cpuPtr.pc += 3
		} else {
			cpuPtr.pc += 2
		}

	default:
		log.Printf("ERROR: EAGLE_PC instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	return true
}
