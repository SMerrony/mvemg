// ram.go

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

package memory

import (
	"log"
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/util"
)

const (
	memSizeWords = 8388608
	MemSizeLCPID = 0x3F
)

// The memoryT structure holds our representation of system RAM.
// It is not exported and should not be directly accessed other than within this package.
type memoryT struct {
	ram        [memSizeWords]dg.WordT
	atuEnabled bool
}

var memory memoryT // memory is NOT exported

// MemInit should be called at machine start
func MemInit() {
	// zero ram?
	memory.atuEnabled = false
	bmcdchInit()
	logging.DebugPrint(logging.DebugLog, "INFO: Initialised %d words of main memory\n", memSizeWords)
}

// ReadByte - read a byte from memory using word address and low-byte flag (true => lower (rightmost) byte)
func ReadByte(wordAddr dg.PhysAddrT, loByte bool) dg.ByteT {
	var res dg.ByteT
	wd := ReadWord(wordAddr)
	if loByte {
		res = dg.ByteT(wd & 0xff)
	} else {
		res = dg.ByteT(wd >> 8)
	}
	return res
}

func ReadByteEclipseBA(byteAddr16 dg.WordT) dg.ByteT {
	var (
		hiLo bool
		addr dg.PhysAddrT
	)
	hiLo = util.TestWbit(byteAddr16, 15) // determine which byte to get
	addr = dg.PhysAddrT(byteAddr16) >> 1
	return ReadByte(addr, hiLo)
}

// WriteByte takes a normal word addr, low-byte flag and datum byte
func WriteByte(wordAddr dg.PhysAddrT, loByte bool, b dg.ByteT) {
	// if wordAddr == 2891 {
	// 	debug.PrintStack()
	// }
	wd := memory.ram[wordAddr]
	if loByte {
		wd = (wd & 0xff00) | dg.WordT(b)
	} else {
		wd = dg.WordT(b)<<8 | (wd & 0x00ff)
	}
	WriteWord(wordAddr, wd)
}

// ReadWord returns the DG Word at the specified physical address
func ReadWord(wordAddr dg.PhysAddrT) dg.WordT {

	if wordAddr >= memSizeWords {
		log.Fatalf("ERROR: Attempt to read word beyond end of physical memory using address: %d", wordAddr)
	}
	return memory.ram[wordAddr]
}

// WriteWord - ALL memory-writing should ultimately go through this function
// N.B. minor exceptions may be made for memory.NsPush() and memory.NsPop()
func WriteWord(wordAddr dg.PhysAddrT, datum dg.WordT) {
	// if wordAddr == 2891 {
	// 	debugCatcher()
	// }
	if wordAddr >= memSizeWords {
		log.Fatalf("ERROR: Attempt to write word beyond end of physical memory using address: %d", wordAddr)
	}
	memory.ram[wordAddr] = datum
}

// ReadDWord returns the doubleword at the given physical address
func ReadDWord(wordAddr dg.PhysAddrT) dg.DwordT {
	if wordAddr >= memSizeWords {
		log.Fatalf("ERROR: Attempt to read doubleword beyond end of physical memory using address: %d", wordAddr)
	}
	var dword dg.DwordT
	//dword = dg_dword(memory.ram[wordAddr]) << 16
	//dword = dword | dg_dword(memory.ram[wordAddr+1])
	dword = util.DWordFromTwoWords(memory.ram[wordAddr], memory.ram[wordAddr+1])
	return dword
}

// WriteDWord writes a doubleword into memory at the given physical address
func WriteDWord(wordAddr dg.PhysAddrT, dwd dg.DwordT) {
	if wordAddr >= memSizeWords {
		log.Fatalf("ERROR: Attempt to write doubleword beyond end of physical memory using address: %d", wordAddr)
	}
	WriteWord(wordAddr, util.DWordGetUpperWord(dwd))
	WriteWord(wordAddr+1, util.DWordGetLowerWord(dwd))
}