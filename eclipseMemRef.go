// eclipseMemRef.go
package main

import (
	"fmt"
	"log"
	"mvemg/logging"
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
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "BLM called with AC1 == 0, not moving anything\n")
			}
			break
		}
		src := dwordGetLowerWord(cpuPtr.ac[2])
		dest := dwordGetLowerWord(cpuPtr.ac[3])
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, fmt.Sprintf("BLM moving %d words from %d to %d\n", numWds, src, dest))
		}
		for numWds != 0 {
			memWriteWord(dg_phys_addr(dest), memReadWord(dg_phys_addr(src)))
			numWds--
			src++
			dest++
		}
		cpuPtr.ac[1] = 0
		cpuPtr.ac[2] = dg_dword(src + 1) // TODO confirm this is right, doc ambiguous
		cpuPtr.ac[3] = dg_dword(dest + 1)

	case "CMP":
		str2len := dwordGetLowerWord(cpuPtr.ac[0])
		str1len := dwordGetLowerWord(cpuPtr.ac[1])
		str1bp := dwordGetLowerWord(cpuPtr.ac[3])
		str2bp := dwordGetLowerWord(cpuPtr.ac[2])
		var byte1, byte2 dg_byte
		res := 0
		for {
			if str1len != 0 {
				byte1 = memReadByteEclipseBA(str1bp)
			} else {
				byte1 = ' '
			}
			if str2len != 0 {
				byte2 = memReadByteEclipseBA(str2bp)
			} else {
				byte2 = ' '
			}
			if byte1 > byte2 {
				res = 1
				break
			}
			if byte1 < byte2 {
				res = -1
				break
			}
			if str1len > 0 {
				str1bp++
				str1len--
			}
			if str1len < 0 {
				str1bp--
				str1len++
			}
			if str2len > 0 {
				str2bp++
				str2len--
			}
			if str2len < 0 {
				str2bp--
				str2len++
			}
			if str1len == 0 && str2len == 0 {
				break
			}
		}
		cpuPtr.ac[0] = dg_dword(str2len)
		cpuPtr.ac[1] = dg_dword(res)
		cpuPtr.ac[2] = dg_dword(str2bp)
		cpuPtr.ac[3] = dg_dword(str1bp)

	case "ELDA":
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		cpuPtr.ac[iPtr.acd] = dg_dword(memReadWord(addr)) & 0x0ffff

	default:
		log.Fatalf("ERROR: ECLIPSE_MEMREF instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
