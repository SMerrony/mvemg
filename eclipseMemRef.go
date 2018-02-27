// eclipseMemRef.go
package main

import (
	"fmt"
	"log"
	"mvemg/logging"

	"github.com/SMerrony/dgemug/util"

	"github.com/SMerrony/dgemug/memory"

	"github.com/SMerrony/dgemug"
)

func eclipseMemRef(cpuPtr *CPUT, iPtr *decodedInstrT) bool {
	var (
		addr               dg.PhysAddrT
		oneAccModeInt2Word oneAccModeInd2WordT
	)

	switch iPtr.ix {

	case instrBLM:
		/* AC0 - unused, AC1 - no. wds to move, AC2 - src, AC3 - dest */
		numWds := util.DwordGetLowerWord(cpuPtr.ac[1])
		if numWds == 0 {
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "BLM called with AC1 == 0, not moving anything\n")
			}
			break
		}
		src := util.DwordGetLowerWord(cpuPtr.ac[2])
		dest := util.DwordGetLowerWord(cpuPtr.ac[3])
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

	case instrCMP:
		cmp(cpuPtr)

	case instrCMV:
		cmv(cpuPtr)

	case instrELDA:
		oneAccModeInt2Word = iPtr.variant.(oneAccModeInd2WordT)
		addr = resolve16bitEclipseAddr(cpuPtr, oneAccModeInt2Word.ind, oneAccModeInt2Word.mode, oneAccModeInt2Word.disp15)
		cpuPtr.ac[oneAccModeInt2Word.acd] = dg.DwordT(memory.ReadWord(addr)) & 0x0ffff

	default:
		log.Fatalf("ERROR: ECLIPSE_MEMREF instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg.PhysAddrT(iPtr.instrLength)
	return true
}

func cmp(cpuPtr *CPUT) {
	var str1len, str2len int16
	str2len = int16(util.DwordGetLowerWord(cpuPtr.ac[0]))
	str1len = int16(util.DwordGetLowerWord(cpuPtr.ac[1]))
	str1bp := util.DwordGetLowerWord(cpuPtr.ac[3])
	str2bp := util.DwordGetLowerWord(cpuPtr.ac[2])
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
}

func cmv(cpuPtr *CPUT) {
	// ACO destCount, AC1 srcCount, AC2 dest byte ptr, AC3 src byte ptr
	var destAscend, srcAscend bool
	destCount := int16(cpuPtr.ac[0] & 0x0000ffff)
	if destCount == 0 {
		log.Println("INFO: CMV called with AC0 == 0, not moving anything")
		return
	}
	destAscend = (destCount > 0)
	srcCount := int16(cpuPtr.ac[1] & 0x0000ffff)
	srcAscend = (srcCount > 0)
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "DEBUG: CMV moving %d chars from %d to %d\n",
			srcCount, cpuPtr.ac[3], cpuPtr.ac[2])
	}
	// set carry if length of src is greater than length of dest
	if cpuPtr.ac[1] > cpuPtr.ac[2] {
		cpuPtr.carry = true
	}
	// 1st move srcCount bytes
	for {
		copyByte(cpuPtr.ac[3], cpuPtr.ac[2])
		if srcAscend {
			cpuPtr.ac[3]++
			srcCount--
		} else {
			cpuPtr.ac[3]--
			srcCount++
		}
		if destAscend {
			cpuPtr.ac[2]++
			destCount--
		} else {
			cpuPtr.ac[2]--
			destCount++
		}
		if srcCount == 0 || destCount == 0 {
			break
		}
	}
	// now fill any excess bytes with ASCII spaces
	if destCount != 0 {
		for {
			memWriteByteBA(asciiSPC, cpuPtr.ac[2])
			if destAscend {
				cpuPtr.ac[2]++
				destCount--
			} else {
				cpuPtr.ac[2]--
				destCount++
			}
			if destCount == 0 {
				break
			}
		}
	}
	cpuPtr.ac[0] = 0
	cpuPtr.ac[1] = dg.DwordT(srcCount)
}
