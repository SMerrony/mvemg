// eclipseMemRef.go

// Copyright (C) 2017,2019  Steve Merrony

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package main

import (
	"fmt"
	"log"

	"github.com/SMerrony/dgemug/dg"
	"github.com/SMerrony/dgemug/logging"
	"github.com/SMerrony/dgemug/memory"
)

func eclipseMemRef(cpuPtr *CPUT, iPtr *decodedInstrT) bool {

	switch iPtr.ix {

	case instrBLM:
		/* AC0 - unused, AC1 - no. wds to move, AC2 - src, AC3 - dest */
		numWds := memory.DwordGetLowerWord(cpuPtr.ac[1])
		if numWds == 0 || numWds > 32768 {
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "BLM called with AC1 out-of-bounds, not moving anything\n")
			}
			break
		}
		src := memory.DwordGetLowerWord(cpuPtr.ac[2])
		dest := memory.DwordGetLowerWord(cpuPtr.ac[3])
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
		cpuPtr.ac[2] = dg.DwordT(src) // TODO confirm this is right, doc ambiguous
		cpuPtr.ac[3] = dg.DwordT(dest)

	case instrCMP:
		cmp(cpuPtr)

	case instrCMV:
		cmv(cpuPtr)

	case instrELDA:
		oneAccModeInt2Word := iPtr.variant.(oneAccModeInd2WordT)
		addr := resolve16bitEffAddr(cpuPtr, oneAccModeInt2Word.ind, oneAccModeInt2Word.mode, oneAccModeInt2Word.disp15, iPtr.dispOffset)
		cpuPtr.ac[oneAccModeInt2Word.acd] = dg.DwordT(memory.ReadWord(addr)) & 0x0ffff

	default:
		log.Printf("ERROR: ECLIPSE_MEMREF instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg.PhysAddrT(iPtr.instrLength)
	return true
}

func cmp(cpuPtr *CPUT) {
	var str1len, str2len int16
	str2len = int16(memory.DwordGetLowerWord(cpuPtr.ac[0]))
	str1len = int16(memory.DwordGetLowerWord(cpuPtr.ac[1]))
	if str1len == 0 && str2len == 0 {
		cpuPtr.ac[1] = 0
		return
	}
	str1bp := memory.DwordGetLowerWord(cpuPtr.ac[3])
	str2bp := memory.DwordGetLowerWord(cpuPtr.ac[2])
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
	destCount := int16(memory.DwordGetLowerWord(cpuPtr.ac[0]))
	if destCount == 0 {
		log.Println("INFO: CMV called with AC0 == 0, not moving anything")
		cpuPtr.carry = false
		return
	}
	destAscend = (destCount > 0)
	srcCount := int16(memory.DwordGetLowerWord(cpuPtr.ac[3]))
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
