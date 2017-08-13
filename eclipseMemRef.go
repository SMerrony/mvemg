// eclipseMemRef.go
package main

import (
	"fmt"
	"log"
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
	"mvemg/util"
)

func eclipseMemRef(cpuPtr *CPU, iPtr *decodedInstrT) bool {
	var (
		addr dg.PhysAddrT
	)

	switch iPtr.mnemonic {

	case "BLM":
		/* AC0 - unused, AC1 - no. wds to move, AC2 - src, AC3 - dest */
		numWds := util.DWordGetLowerWord(cpuPtr.ac[1])
		if numWds == 0 {
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "BLM called with AC1 == 0, not moving anything\n")
			}
			break
		}
		src := util.DWordGetLowerWord(cpuPtr.ac[2])
		dest := util.DWordGetLowerWord(cpuPtr.ac[3])
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, fmt.Sprintf("BLM moving %d words from %d to %d\n", numWds, src, dest))
		}
		for numWds != 0 {
			memory.WriteWord(dg.PhysAddrT(dest), memory.ReadWord(dg.PhysAddrT(src)))
			numWds--
			src++
			dest++
		}
		cpuPtr.ac[1] = 0
		cpuPtr.ac[2] = dg.DwordT(src + 1) // TODO confirm this is right, doc ambiguous
		cpuPtr.ac[3] = dg.DwordT(dest + 1)

	case "CMP":
		str2len := util.DWordGetLowerWord(cpuPtr.ac[0])
		str1len := util.DWordGetLowerWord(cpuPtr.ac[1])
		str1bp := util.DWordGetLowerWord(cpuPtr.ac[3])
		str2bp := util.DWordGetLowerWord(cpuPtr.ac[2])
		var byte1, byte2 dg.ByteT
		res := 0
		for {
			if str1len != 0 {
				byte1 = memory.ReadByteEclipseBA(str1bp)
			} else {
				byte1 = ' '
			}
			if str2len != 0 {
				byte2 = memory.ReadByteEclipseBA(str2bp)
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
		cpuPtr.ac[0] = dg.DwordT(str2len)
		cpuPtr.ac[1] = dg.DwordT(res)
		cpuPtr.ac[2] = dg.DwordT(str2bp)
		cpuPtr.ac[3] = dg.DwordT(str1bp)

	case "ELDA":
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)
		cpuPtr.ac[iPtr.acd] = dg.DwordT(memory.ReadWord(addr)) & 0x0ffff

	default:
		log.Fatalf("ERROR: ECLIPSE_MEMREF instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg.PhysAddrT(iPtr.instrLength)
	return true
}
