// eagleMemRef.go

// Copyright (C) 2017  Steve Merrony

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
	"log"
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
	"mvemg/util"
)

func eagleMemRef(cpuPtr *CPUT, iPtr *decodedInstrT) bool {
	var (
		addr, offset       dg.PhysAddrT
		byt                dg.ByteT
		wd                 dg.WordT
		dwd                dg.DwordT
		i16a, i16b         int16
		i32                int32
		lowByte, ok        bool
		wordAddr           dg.PhysAddrT
		immMode2Word       immMode2WordT
		oneAccMode2Word    oneAccMode2WordT
		oneAccModeInd2Word oneAccModeInd2WordT
		oneAccModeInd3Word oneAccModeInd3WordT
		twoAcc1Word        twoAcc1WordT
	)

	switch iPtr.mnemonic {

	case "LNLDA":
		oneAccModeInd3Word = iPtr.variant.(oneAccModeInd3WordT)
		addr = resolve32bitEffAddr(cpuPtr, oneAccModeInd3Word.ind, oneAccModeInd3Word.mode, oneAccModeInd3Word.disp31)
		cpuPtr.ac[oneAccModeInd3Word.acd] = util.SexWordToDWord(memory.ReadWord(addr))

	case "LNSTA":
		oneAccModeInd3Word = iPtr.variant.(oneAccModeInd3WordT)
		addr = resolve32bitEffAddr(cpuPtr, oneAccModeInd3Word.ind, oneAccModeInd3Word.mode, oneAccModeInd3Word.disp31)
		wd = util.DWordGetLowerWord(cpuPtr.ac[oneAccModeInd3Word.acd])
		memory.WriteWord(addr, wd)

	case "LWLDA":
		oneAccModeInd3Word = iPtr.variant.(oneAccModeInd3WordT)
		addr = resolve32bitEffAddr(cpuPtr, oneAccModeInd3Word.ind, oneAccModeInd3Word.mode, oneAccModeInd3Word.disp31)
		cpuPtr.ac[oneAccModeInd3Word.acd] = memory.ReadDWord(addr)

	case "LWSTA":
		oneAccModeInd3Word = iPtr.variant.(oneAccModeInd3WordT)
		addr = resolve32bitEffAddr(cpuPtr, oneAccModeInd3Word.ind, oneAccModeInd3Word.mode, oneAccModeInd3Word.disp31)
		dwd = cpuPtr.ac[oneAccModeInd3Word.acd]
		memory.WriteDWord(addr, dwd)

	case "WBLM":
		wblm(cpuPtr)

	case "WBTO":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		if twoAcc1Word.acs == twoAcc1Word.acd {
			addr = 0
		} else {
			addr = resolve32bitIndirectableAddr(cpuPtr.ac[twoAcc1Word.acs])
		}
		offset = dg.PhysAddrT(cpuPtr.ac[twoAcc1Word.acd]) >> 4
		bitNum := uint(cpuPtr.ac[twoAcc1Word.acd] & 0x0f)
		wd = memory.ReadWord(addr + offset)
		wd = util.SetWbit(wd, bitNum)
		memory.WriteWord(addr+offset, wd)

	case "WBTZ":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		if twoAcc1Word.acs == twoAcc1Word.acd {
			addr = 0
		} else {
			addr = resolve32bitIndirectableAddr(cpuPtr.ac[twoAcc1Word.acs])
		}
		offset = dg.PhysAddrT(cpuPtr.ac[twoAcc1Word.acd]) >> 4
		bitNum := uint(cpuPtr.ac[twoAcc1Word.acd] & 0x0f)
		wd = memory.ReadWord(addr + offset)
		wd = util.ClearWbit(wd, bitNum)
		memory.WriteWord(addr+offset, wd)

	case "WCMV":
		wcmv(cpuPtr)

	case "WCMP":
		wcmp(cpuPtr)

	case "WCST":
		wcst(cpuPtr)

	case "WCTR":
		wctr(cpuPtr)

	case "WLDB":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		wordAddr = dg.PhysAddrT(cpuPtr.ac[twoAcc1Word.acs]) >> 1
		lowByte = util.TestDWbit(cpuPtr.ac[twoAcc1Word.acs], 31)
		byt = memory.ReadByte(wordAddr, lowByte)
		cpuPtr.ac[twoAcc1Word.acd] = dg.DwordT(byt)

	case "WSTB":
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		byt = dg.ByteT(cpuPtr.ac[twoAcc1Word.acd] & 0x0ff)
		memWriteByteBA(byt, cpuPtr.ac[twoAcc1Word.acs])

	case "XLDB":
		oneAccMode2Word = iPtr.variant.(oneAccMode2WordT)
		cpuPtr.ac[oneAccMode2Word.acd] = dg.DwordT(memory.ReadByte(resolve16bitEagleAddr(cpuPtr,
			' ',
			oneAccMode2Word.mode,
			oneAccMode2Word.disp16),
			oneAccMode2Word.bitLow)) & 0x00ff

	case "XLEF":
		oneAccModeInd2Word = iPtr.variant.(oneAccModeInd2WordT)
		cpuPtr.ac[oneAccModeInd2Word.acd] = dg.DwordT(resolve16bitEagleAddr(cpuPtr, oneAccModeInd2Word.ind, oneAccModeInd2Word.mode, oneAccModeInd2Word.disp15))

	case "XLEFB":
		oneAccMode2Word = iPtr.variant.(oneAccMode2WordT)
		loBit := oneAccMode2Word.disp16 & 1
		addr = resolve16bitEagleAddr(cpuPtr, 0, oneAccMode2Word.mode, oneAccMode2Word.disp16/2)
		addr <<= 1
		if loBit == 1 {
			addr++
		}
		cpuPtr.ac[oneAccMode2Word.acd] = dg.DwordT(addr)

	case "XNADD", "XNSUB":
		oneAccModeInd2Word = iPtr.variant.(oneAccModeInd2WordT)
		addr = resolve16bitEagleAddr(cpuPtr, oneAccModeInd2Word.ind, oneAccModeInd2Word.mode, oneAccModeInd2Word.disp15)
		i16b = int16(memory.ReadWord(addr))
		i16a = int16(util.DWordGetLowerWord(cpuPtr.ac[oneAccModeInd2Word.acd]))
		if iPtr.mnemonic == "XNADD" {
			i16a += i16b
		} else {
			i16a -= i16b
		}
		cpuPtr.ac[oneAccModeInd2Word.acd] = util.SexWordToDWord(dg.WordT(i16a))

	case "XNLDA":
		oneAccModeInd2Word = iPtr.variant.(oneAccModeInd2WordT)
		addr = resolve16bitEagleAddr(cpuPtr, oneAccModeInd2Word.ind, oneAccModeInd2Word.mode, oneAccModeInd2Word.disp15)
		wd, ok = memory.ReadWordTrap(addr)
		if !ok {
			return false
		}
		// FIXME check this...
		cpuPtr.ac[oneAccModeInd2Word.acd] = util.SexWordToDWord(wd)

	case "XSTB":
		oneAccMode2Word = iPtr.variant.(oneAccMode2WordT)
		byt = dg.ByteT(cpuPtr.ac[oneAccMode2Word.acd])
		memory.WriteByte(resolve16bitEagleAddr(cpuPtr,
			' ',
			oneAccMode2Word.mode,
			oneAccMode2Word.disp16),
			oneAccMode2Word.bitLow,
			byt)

	case "XNSTA":
		oneAccModeInd2Word = iPtr.variant.(oneAccModeInd2WordT)
		addr = resolve16bitEagleAddr(cpuPtr, oneAccModeInd2Word.ind, oneAccModeInd2Word.mode, oneAccModeInd2Word.disp15)
		wd = util.DWordGetLowerWord(cpuPtr.ac[oneAccModeInd2Word.acd])
		memory.WriteWord(addr, wd)

	case "XWADI":
		// add 1-4 to signed 32-bit acc
		immMode2Word = iPtr.variant.(immMode2WordT)
		addr = resolve16bitEagleAddr(cpuPtr, immMode2Word.ind, immMode2Word.mode, immMode2Word.disp15)
		i32 = int32(memory.ReadDWord(addr)) + int32(immMode2Word.immU16)
		// FIXME handle Carry and OVeRflow
		memory.WriteDWord(addr, dg.DwordT(i32))

	case "XWLDA":
		oneAccModeInd2Word = iPtr.variant.(oneAccModeInd2WordT)
		addr = resolve16bitEagleAddr(cpuPtr, oneAccModeInd2Word.ind, oneAccModeInd2Word.mode, oneAccModeInd2Word.disp15)
		dwd = memory.ReadDWord(addr)
		cpuPtr.ac[oneAccModeInd2Word.acd] = dwd

	case "XWSTA":
		oneAccModeInd2Word = iPtr.variant.(oneAccModeInd2WordT)
		addr = resolve16bitEagleAddr(cpuPtr, oneAccModeInd2Word.ind, oneAccModeInd2Word.mode, oneAccModeInd2Word.disp15)
		dwd = cpuPtr.ac[oneAccModeInd2Word.acd]
		memory.WriteDWord(addr, dwd)

	default:
		log.Fatalf("ERROR: EAGLE_MEMREF instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg.PhysAddrT(iPtr.instrLength)
	return true
}

func readByteBA(ba dg.DwordT) dg.ByteT {
	wordAddr, lowByte := resolve32bitByteAddr(ba)
	return memory.ReadByte(wordAddr, lowByte)
}

// memWriteByte writes the supplied byte to the address derived from the given byte addr
func memWriteByteBA(b dg.ByteT, ba dg.DwordT) {
	wordAddr, lowByte := resolve32bitByteAddr(ba)
	memory.WriteByte(wordAddr, lowByte, b)
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "DEBUG: memWriteByte wrote %c to word addr: %d\n", b, wordAddr)
	}
}

func copyByte(srcBA, destBA dg.DwordT) {
	memWriteByteBA(readByteBA(srcBA), destBA)
}

func wblm(cpuPtr *CPUT) {
	/* AC0 - unused, AC1 - no. wds to move (if neg then descending order), AC2 - src, AC3 - dest */
	numWds := int32(cpuPtr.ac[1])
	var order int32 = 1
	if numWds < 0 {
		order = -1
	}
	if numWds == 0 {
		log.Println("INFO: WBLM called with AC1 == 0, not moving anything")
		return
	}
	src := dg.PhysAddrT(cpuPtr.ac[2])
	dest := dg.PhysAddrT(cpuPtr.ac[3])
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "DEBUG: WBLM moving %d words from %d to %d\n", numWds, src, dest)
	}
	for numWds != 0 {
		memory.WriteWord(dest, memory.ReadWord(src))
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
	//cpuPtr.ac[2] = dg_dword(dest) // TODO confirm this
	//cpuPtr.ac[3] = dg_dword(dest)
	// TESTING..
	cpuPtr.ac[2] = dg.DwordT(src + 1) // TODO confirm this
	cpuPtr.ac[3] = dg.DwordT(dest + 1)
}

func wcmv(cpuPtr *CPUT) {
	// ACO destCount, AC1 srcCount, AC2 dest byte ptr, AC3 src byte ptr
	var destAscend, srcAscend bool
	destCount := int32(cpuPtr.ac[0])
	if destCount == 0 {
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, ".... WCMV called with AC0 == 0, not moving anything\n")
		}
		return
	}
	destAscend = (destCount > 0)
	srcCount := int32(cpuPtr.ac[1])
	srcAscend = (srcCount > 0)
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, ".... WCMV moving %d chars from %d to %d\n",
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

func getDirection(ac dg.DwordT) int32 {
	if int32(ac) > 0 {
		return 1
	}
	if int32(ac) < 0 {
		return -1
	}
	return 0
}

func wcmp(cpuPtr *CPUT) {
	// AC0 String2 length and dir (bwd if -ve)
	// AC1 String1 length and dir (bwd if -ve)
	// AC2 Byte Pointer to first byte of String2 to be compared
	// AC3 Byte Pointer to first byte of String1 to be compared
	str2dir := getDirection(cpuPtr.ac[0])
	str1dir := getDirection(cpuPtr.ac[1])
	var str1char, str2char dg.ByteT
	if str1dir == 0 && str2dir == 0 {
		return
	}
	for cpuPtr.ac[1] != 0 && cpuPtr.ac[0] != 0 {
		// read the two bytes to compare, substitute with a space if one string has run out
		if cpuPtr.ac[1] != 0 {
			str1char = readByteBA(cpuPtr.ac[3])
		} else {
			str1char = ' '
		}
		if cpuPtr.ac[0] != 0 {
			str2char = readByteBA(cpuPtr.ac[2])
		} else {
			str2char = ' '
		}
		// compare
		if str1char < str2char {
			cpuPtr.ac[1] = 0xFFFFFFFF
			return
		}
		if str1char > str2char {
			cpuPtr.ac[1] = 1
			return
		}
		// they were equal, so adjust remaining lengths, move pointers, and loop round
		if cpuPtr.ac[0] != 0 {
			cpuPtr.ac[0] = dg.DwordT(int32(cpuPtr.ac[0]) + str2dir)
		}
		if cpuPtr.ac[1] != 0 {
			cpuPtr.ac[1] = dg.DwordT(int32(cpuPtr.ac[1]) + str1dir)
		}
		cpuPtr.ac[2] = dg.DwordT(int32(cpuPtr.ac[2]) + str2dir)
		cpuPtr.ac[3] = dg.DwordT(int32(cpuPtr.ac[3]) + str1dir)
	}
}

func wcst(cpuPtr *CPUT) {
	strLenDir := int(int32(cpuPtr.ac[1]))
	if strLenDir == 0 {
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, ".... WCST called with AC1 == 0, not scanning anything\n")
		}
		return
	}
	delimTabAddr := resolve32bitIndirectableAddr(cpuPtr.ac[0])
	cpuPtr.ac[0] = dg.DwordT(delimTabAddr)
	// load the table which is 256 bits stored as 16 words
	var table [256]bool
	var tIx dg.PhysAddrT
	for tIx = 0; tIx < 16; tIx++ {
		wd := memory.ReadWord(delimTabAddr + tIx)
		for bit := 0; bit < 16; bit++ {
			if util.TestWbit(wd, bit) {
				table[(int(tIx)*16)+bit] = true
			}
		}
	}
	// table[] now contains true for any delimiter
	var dir int32 = 1
	if strLenDir < 0 {
		dir = -1
	}

	for strLenDir != 0 {
		thisChar := readByteBA(cpuPtr.ac[3])
		if table[int(thisChar)] {
			// match, so set AC1 and return
			cpuPtr.ac[1] = 0
			return
		}
		cpuPtr.ac[1] = dg.DwordT(int32(cpuPtr.ac[1]) + dir)
		cpuPtr.ac[3] = dg.DwordT(int32(cpuPtr.ac[3]) + dir)
		strLenDir += int(dir)
	}
}

func wctr(cpuPtr *CPUT) {
	// AC0 Wide Byte addr of translation table - unchanged
	// AC1 # of bytes in each string, NB. -ve => translate-and-move mode, +ve => translate-and-compare mode
	// AC2 destination string ("string2") Byte addr
	// AC3 source string ("string1") byte addr
	if cpuPtr.ac[1] == 0 {
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "INFO: WCTR called with AC1 == 0, not translating anything\n")
		}
		return
	}
	transTablePtr := dg.DwordT(resolve32bitIndirectableAddr(cpuPtr.ac[0]))
	// build an array representation of the table
	var transTable [256]dg.ByteT
	var c dg.DwordT
	for c = 0; c < 256; c++ {
		transTable[c] = readByteBA(transTablePtr + c)
	}

	for cpuPtr.ac[1] != 0 {
		srcByte := readByteBA(cpuPtr.ac[3])
		cpuPtr.ac[3]++
		transByte := transTable[int(srcByte)]
		if int32(cpuPtr.ac[1]) < 0 {
			// move mode
			memWriteByteBA(transByte, cpuPtr.ac[2])
			cpuPtr.ac[2]++
			cpuPtr.ac[1]++
		} else {
			// compare mode
			str2byte := readByteBA(cpuPtr.ac[2])
			cpuPtr.ac[2]++
			trans2byte := transTable[int(str2byte)]
			if srcByte < trans2byte {
				cpuPtr.ac[1] = 0xffffffff
				break
			}
			if srcByte > trans2byte {
				cpuPtr.ac[1] = 1
				break
			}
			cpuPtr.ac[1]--
		}
	}
}
