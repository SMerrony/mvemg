// memory.go

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
	"fmt"
	"log"
	"mvemg/logging"
	"runtime/debug"
)

const (
	MEM_SIZE_WORDS = 8388608
	MEM_SIZE_LCPID = 0x3F

	// Some Page Zero special locations...

	WFP_LOC  = 020
	WSP_LOC  = 022
	WSL_LOC  = 024
	WSB_LOC  = 026
	WPFH_LOC = 030
	CBP_LOC  = 032

	NSP_LOC  = 040 // 32. Narrow Stack Pointer
	NFP_LOC  = 041
	NSL_LOC  = 042
	NSFA_LOC = 043
)

type memoryT struct {
	ram                 [MEM_SIZE_WORDS]DgWordT
	atuEnabled          bool
	pushCount, popCount int
}

var memory memoryT

func memInit() {
	// zero ram?
	memory.atuEnabled = false
	bmcdchInit()
	logging.DebugPrint(logging.DebugLog, "INFO: Initialised %d words of main memory\n", MEM_SIZE_WORDS)
}

// read a byte from memory using word address and low-byte flag (true => lower (rightmost) byte)
func memReadByte(wordAddr DgPhysAddrT, loByte bool) DgByteT {
	var res DgByteT
	wd := memReadWord(wordAddr)
	if loByte {
		res = DgByteT(wd & 0xff)
	} else {
		res = DgByteT(wd >> 8)
	}
	return res
}

func memReadByteEclipseBA(byteAddr16 DgWordT) DgByteT {
	var (
		hiLo bool
		addr DgPhysAddrT
	)
	hiLo = testWbit(byteAddr16, 15) // determine which byte to get
	addr = DgPhysAddrT(byteAddr16) >> 1
	return memReadByte(addr, hiLo)
}

func memWriteByte(wordAddr DgPhysAddrT, loByte bool, b DgByteT) {
	// if wordAddr == 2891 {
	// 	debug.PrintStack()
	// }
	wd := memory.ram[wordAddr]
	if loByte {
		wd = (wd & 0xff00) | DgWordT(b)
	} else {
		wd = DgWordT(b)<<8 | (wd & 0x00ff)
	}
	memWriteWord(wordAddr, wd)
}

func memReadByteBA(ba DgDwordT) DgByteT {
	wordAddr, lowByte := resolve32bitByteAddr(ba)
	return memReadByte(wordAddr, lowByte)
}

// MemWriteByte writes the supplied byte to the address derived from the given byte addr
func memWriteByteBA(b DgByteT, ba DgDwordT) {
	wordAddr, lowByte := resolve32bitByteAddr(ba)
	memWriteByte(wordAddr, lowByte, b)
}

func memCopyByte(srcBA, destBA DgDwordT) {
	memWriteByteBA(memReadByteBA(srcBA), destBA)
}

func debugCatcher() {
	debug.PrintStack()
}

func memReadWord(wordAddr DgPhysAddrT) DgWordT {

	if wordAddr >= MEM_SIZE_WORDS {
		log.Fatalf("ERROR: Attempt to read word beyond end of physical memory using address: %d", wordAddr)
	}
	return memory.ram[wordAddr]
}

func memReadWordDchChan(addr DgPhysAddrT) DgWordT {
	pAddr := addr
	if getDchMode() {
		pAddr, _ = getBmcDchMapAddr(addr)
	}
	logging.DebugPrint(logging.MapLog, "memReadWordBmcChan got addr: %d, read from addr: %d\n", addr, pAddr)
	return memReadWord(pAddr)
}

func memReadWordBmcChan(addr DgPhysAddrT) DgWordT {
	var pAddr DgPhysAddrT
	decodedAddr := decodeBmcAddr(addr)
	if decodedAddr.isLogical {
		pAddr, _ = getBmcDchMapAddr(addr) // FIXME
	} else {
		pAddr = decodedAddr.ca
	}
	logging.DebugPrint(logging.MapLog, "memWriteReadBmcChan got addr: %d, wrote to addr: %d\n", addr, pAddr)
	return memReadWord(pAddr)
}

// memWriteWord - ALL memory-writing should ultimately go through this function
// N.B. minor exceptions may be made for nsPush() and nsPop()
func memWriteWord(wordAddr DgPhysAddrT, datum DgWordT) {
	// if wordAddr == 2891 {
	// 	debugCatcher()
	// }
	if wordAddr >= MEM_SIZE_WORDS {
		log.Fatalf("ERROR: Attempt to write word beyond end of physical memory using address: %d", wordAddr)
	}
	memory.ram[wordAddr] = datum
}

func memWriteWordDchChan(addr DgPhysAddrT, data DgWordT) DgPhysAddrT {
	pAddr := addr

	if getDchMode() {
		pAddr, _ = getBmcDchMapAddr(addr)
	}
	memWriteWord(pAddr, data)
	logging.DebugPrint(logging.MapLog, "memWriteWordDchChan got addr: %d, wrote to addr: %d\n", addr, pAddr)
	return pAddr
}

func memWriteWordBmcChan(addr DgPhysAddrT, data DgWordT) DgPhysAddrT {
	var pAddr DgPhysAddrT
	decodedAddr := decodeBmcAddr(addr)
	if decodedAddr.isLogical {
		pAddr, _ = getBmcDchMapAddr(addr) // FIXME
	} else {
		pAddr = decodedAddr.ca
	}
	memWriteWord(pAddr, data)
	logging.DebugPrint(logging.MapLog, "memWriteWordBmcChan got addr: %d, wrote to addr: %d\n", addr, pAddr)
	return pAddr
}

func memReadDWord(wordAddr DgPhysAddrT) DgDwordT {
	if wordAddr >= MEM_SIZE_WORDS {
		log.Fatalf("ERROR: Attempt to read doubleword beyond end of physical memory using address: %d", wordAddr)
	}
	var dword DgDwordT
	//dword = dg_dword(memory.ram[wordAddr]) << 16
	//dword = dword | dg_dword(memory.ram[wordAddr+1])
	dword = dwordFromTwoWords(memory.ram[wordAddr], memory.ram[wordAddr+1])
	return dword
}

func memWriteDWord(wordAddr DgPhysAddrT, dwd DgDwordT) {
	if wordAddr >= MEM_SIZE_WORDS {
		log.Fatalf("ERROR: Attempt to write doubleword beyond end of physical memory using address: %d", wordAddr)
	}
	memWriteWord(wordAddr, dwordGetUpperWord(dwd))
	memWriteWord(wordAddr+1, dwordGetLowerWord(dwd))
}

// PUSH a word onto the Narrow Stack
func nsPush(seg DgPhysAddrT, data DgWordT) {
	// TODO segment handling
	// TODO overflow/underflow handling - either here or in instruction?
	memory.ram[NSP_LOC]++ // we allow this direct write to a fixed location for performance
	addr := DgPhysAddrT(memory.ram[NSP_LOC])
	memWriteWord(addr, data)
	logging.DebugPrint(logging.DebugLog, "... nsPush pushed %8d onto the Narrow Stack at location: %d\n", data, addr)
}

// POP a word off the Narrow Stack
func nsPop(seg DgPhysAddrT) DgWordT {
	// TODO segment handling
	// TODO overflow/underflow handling - either here or in instruction?
	addr := DgPhysAddrT(memory.ram[NSP_LOC])
	data := memReadWord(addr)
	logging.DebugPrint(logging.DebugLog, "... nsPop  popped %8d off  the Narrow Stack at location: %d\n", data, addr)
	memory.ram[NSP_LOC]-- // we allow this direct write to a fixed location for performance
	return data
}

// PUSH a doubleword onto the Wide Stack
func wsPush(seg DgPhysAddrT, data DgDwordT) {
	// TODO segment handling
	// TODO overflow/underflow handling - either here or in instruction?
	wsp := memReadDWord(WSP_LOC) + 2
	memWriteDWord(WSP_LOC, wsp)
	memWriteDWord(DgPhysAddrT(wsp), data)
	logging.DebugPrint(logging.DebugLog, "... wsPush pushed %8d onto the Wide Stack at location: %d\n", data, wsp)
}

// POP a word off the Wide Stack
func wsPop(seg DgPhysAddrT) DgDwordT {
	// TODO segment handling
	// TODO overflow/underflow handling - either here or in instruction?
	wsp := memReadDWord(WSP_LOC)
	dword := memReadDWord(DgPhysAddrT(wsp))
	memWriteDWord(WSP_LOC, wsp-2)
	logging.DebugPrint(logging.DebugLog, "... wsPop  popped %8d off  the Wide Stack at location: %d\n", dword, wsp)
	return dword
}

// AdvanceWSP increases the WSP by the given amount of DWords
func AdvanceWSP(dwdCnt uint) {
	wsp := memReadDWord(WSP_LOC) + DgDwordT(dwdCnt*2)
	memWriteDWord(WSP_LOC, wsp)
	logging.DebugPrint(logging.DebugLog, "... WSP advanced by %d DWords to %d\n", dwdCnt, wsp)
}

// utility functions

// BoolToInt converts a bool to 1 or 0
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// BoolToYN converts a bool to Y or N
func BoolToYN(b bool) byte {
	if b {
		return 'Y'
	}
	return 'N'
}

// BoolToOnOff converts a bool to "On" or "Off"
func BoolToOnOff(b bool) string {
	if b {
		return "On"
	}
	return "Off"
}

// BoolToOZ converts a boolean to a O(ne) or Z(ero) byte
func boolToOZ(b bool) byte {
	if b {
		return 'O'
	}
	return 'Z'
}

// dwordGetLowerWord gets the DG-lower word of a doubleword
// Called VERY often, hopefully inlined!
func dwordGetLowerWord(dwd DgDwordT) DgWordT {
	return DgWordT(dwd) // & 0x0000ffff mask unneccessary
}

func dwordGetUpperWord(dwd DgDwordT) DgWordT {
	return DgWordT(dwd >> 16)
}

func dwordFromTwoWords(hw DgWordT, lw DgWordT) DgDwordT {
	return DgDwordT(hw)<<16 | DgDwordT(lw)
}

// in the DG world, the first (leftmost) bit is numbered zero...
// extract nbits from value starting at leftBit
func getWbits(value DgWordT, leftBit int, nbits int) DgWordT {
	var res DgWordT
	rightBit := leftBit + nbits
	for b := leftBit; b < rightBit; b++ {
		res = res << 1
		if testWbit(value, b) {
			res++
		}
	}
	return res
}

// SetWbit sets a single bit in a DG word
func SetWbit(word DgWordT, bitNum uint) DgWordT {
	return word | 1<<(15-bitNum)
}

// ClearWbit clears a single bit in a DG word
func ClearWbit(word DgWordT, bitNum uint) DgWordT {
	return word ^ 1<<(15-bitNum)
}

// in the DG world, the first (leftmost) bit is numbered zero...
// extract nbits from value starting at leftBit
func getDWbits(value DgDwordT, leftBit int, nbits int) DgDwordT {
	var res DgDwordT
	rightBit := leftBit + nbits
	for b := leftBit; b < rightBit; b++ {
		res = res << 1
		if testDWbit(value, b) {
			res++
		}
	}
	return res
}

// sign-extend a DG word to a DG DoubleWord
func sexWordToDWord(wd DgWordT) DgDwordT {
	var dwd DgDwordT
	if testWbit(wd, 0) {
		dwd = DgDwordT(wd) | 0xffff0000
	} else {
		dwd = DgDwordT(wd) & 0x0000ffff
	}
	return dwd
}

// swap over the two bytes in a dg_word
func swapBytes(wd DgWordT) DgWordT {
	var res DgWordT
	res = (wd >> 8) | ((wd & 0x00ff) << 8)
	return res
}

var bb uint8

// does word w have bit b set?
func testWbit(w DgWordT, b int) bool {
	bb = uint8(b)
	return (w & (1 << (15 - bb))) != 0
}

// does dword dw have bit b set?
func testDWbit(dw DgDwordT, b int) bool {
	bb = uint8(b)
	return ((dw & (1 << (31 - bb))) != 0)
}

// get a pretty-printable string of a word
func wordToBinStr(w DgWordT) string {
	return fmt.Sprintf("%08b %08b", w>>8, w&0x0ff)
}
