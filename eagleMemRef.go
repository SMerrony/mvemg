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

func eagleMemRef(cpuPtr *CPU, iPtr *decodedInstrT) bool {
	var (
		addr dg.PhysAddrT
		byt  dg.ByteT
		wd   dg.WordT
		dwd  dg.DwordT
		i32  int32
	)

	switch iPtr.mnemonic {

	case "LNLDA":
		addr = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp31)
		cpuPtr.ac[iPtr.acd] = util.SexWordToDWord(memory.ReadWord(addr))

	case "LNSTA":
		addr = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp31)
		wd = util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])
		memory.WriteWord(addr, wd)

	case "LWLDA":
		addr = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp31)
		cpuPtr.ac[iPtr.acd] = memory.ReadDWord(addr)

	case "LWSTA":
		addr = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp31)
		dwd = cpuPtr.ac[iPtr.acd]
		memory.WriteDWord(addr, dwd)

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

	case "WCMV": // ACO destCount, AC1 srcCount, AC2 dest byte ptr, AC3 src byte ptr
		var destAscend, srcAscend bool
		destCount := int32(cpuPtr.ac[0])
		if destCount == 0 {
			break
		}
		destAscend = (destCount > 0)
		srcCount := int32(cpuPtr.ac[1])
		srcAscend = (srcCount > 0)
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "DEBUG: WCMV moving %d chars from %d to %d\n",
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
				memWriteByteBA(ASCII_SPC, cpuPtr.ac[2])
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

	case "WSTB":
		byt = dg.ByteT(cpuPtr.ac[iPtr.acd] & 0x0ff)
		memWriteByteBA(byt, cpuPtr.ac[iPtr.acs])

	case "XLDB":
		cpuPtr.ac[iPtr.acd] = dg.DwordT(memory.ReadByte(resolve16bitEagleAddr(cpuPtr, ' ', iPtr.mode, iPtr.disp16), iPtr.bitLow)) & 0x00ff

	case "XLEF":
		cpuPtr.ac[iPtr.acd] = dg.DwordT(resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15))

	case "XLEFB":
		loBit := iPtr.disp16 & 1
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp16/2)
		addr <<= 1
		if loBit == 1 {
			addr++
		}
		cpuPtr.ac[iPtr.acd] = dg.DwordT(addr)

	case "XNLDA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)
		wd = memory.ReadWord(addr)
		cpuPtr.ac[iPtr.acd] = util.SexWordToDWord(wd) // FIXME check this...

	case "XNSTA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)
		wd = util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])
		memory.WriteWord(addr, wd)

	case "XWADI":
		// add 1-4 to signed 32-bit acc
		variant := iPtr.variant.(immMode2WordT)
		addr = resolve16bitEagleAddr(cpuPtr, variant.ind, variant.mode, variant.disp15)
		i32 = int32(memory.ReadDWord(addr)) + int32(variant.immU16)
		// FIXME handle Carry and OVeRflow
		memory.WriteDWord(addr, dg.DwordT(i32))

	case "XWLDA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)
		dwd = memory.ReadDWord(addr)
		cpuPtr.ac[iPtr.acd] = dwd

	case "XWSTA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)
		dwd = cpuPtr.ac[iPtr.acd]
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
}

func copyByte(srcBA, destBA dg.DwordT) {
	memWriteByteBA(readByteBA(srcBA), destBA)
}
