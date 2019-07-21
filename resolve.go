// resolve.go

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

	"github.com/SMerrony/dgemug/logging"

	"github.com/SMerrony/dgemug/memory"

	"github.com/SMerrony/dgemug/dg"
)

const (
	physMask16 = 0x7fff
	physMask32 = 0x7fffffff
)

func resolve16bitEclipseAddr(cpuPtr *CPUT, ind byte, mode int, disp int16) dg.PhysAddrT {

	var (
		eff     dg.PhysAddrT
		intEff  int32
		indAddr dg.WordT
	)

	// handle addressing mode...
	switch mode {
	case absoluteMode:
		intEff = int32(disp)
	case pcMode:
		intEff = int32(cpuPtr.pc) + int32(disp)
	case ac2Mode:
		intEff = int32(cpuPtr.ac[2]) + int32(disp)
	case ac3Mode:
		intEff = int32(cpuPtr.ac[3]) + int32(disp)
	}

	// handle indirection
	if ind == '@' { // down the rabbit hole...
		indAddr = memory.ReadWord(dg.PhysAddrT(intEff))
		for memory.TestWbit(indAddr, 0) {
			indAddr = memory.ReadWord(dg.PhysAddrT(indAddr))
		}
		intEff = int32(indAddr)
	}

	// Handle wrap-around
	if intEff < 0 {
		intEff += 65536
	}
	if intEff > 65535 {
		intEff -= 65536
	}

	// mask off to Eclipse range
	eff = dg.PhysAddrT(intEff) & 0x7fff

	// if debugLogging {
	// 	logging.DebugPrint(logging.DebugLog, "... resolve16bitEclipseAddr got: %#o, returning %#o\n", disp, eff)
	// }
	return eff
}

// This is the same as resolve16bitEclipseAddr, but without the range masking at the end
func resolve16bitEagleAddr(cpuPtr *CPUT, ind byte, mode int, disp int16) dg.PhysAddrT {

	var (
		eff     dg.PhysAddrT
		intEff  int32
		indAddr dg.WordT
		ok      bool
	)

	// handle addressing mode...
	switch mode {
	case absoluteMode:
		intEff = int32(disp)
	case pcMode:
		intEff = int32(cpuPtr.pc) + int32(disp)
	case ac2Mode:
		intEff = int32(cpuPtr.ac[2]) + int32(disp)
	case ac3Mode:
		intEff = int32(cpuPtr.ac[3]) + int32(disp)
	}

	// handle indirection
	if ind == '@' { // down the rabbit hole...
		indAddr = memory.ReadWord(dg.PhysAddrT(intEff))
		for memory.TestWbit(indAddr, 0) {
			indAddr, ok = memory.ReadWordTrap(dg.PhysAddrT(indAddr & physMask16))
			if !ok {
				log.Fatalf("ERROR: PC=%d", cpuPtr.pc)
			}
		}
		intEff = int32(indAddr)
		// res, ok := memory.ReadWordTrap(dg.PhysAddrT(indAddr))
		// if !ok {
		// 	log.Fatalf("ERROR: PC=%d", cpuPtr.pc)
		// }
		// intEff = int32(res)
	}

	eff = dg.PhysAddrT(intEff)

	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... resolve16bitEagleAddr got: %#o, returning %#o\n", disp, eff)
	}
	return eff
}

// Resolve32bitByteAddr returns the word address and low-byte flag for a given 32-bit byte address
func resolve32bitByteAddr(byteAddr dg.DwordT) (wordAddr dg.PhysAddrT, loByte bool) {
	wordAddr = dg.PhysAddrT(byteAddr) >> 1
	loByte = memory.TestDwbit(byteAddr, 31)
	return wordAddr, loByte
}

func resolve32bitEffAddr(cpuPtr *CPUT, ind byte, mode int, disp int32) dg.PhysAddrT {

	eff := dg.PhysAddrT(disp)

	// handle addressing mode...
	switch mode {
	case absoluteMode:
		// nothing to do
	case pcMode:
		eff += cpuPtr.pc
	case ac2Mode:
		eff += dg.PhysAddrT(cpuPtr.ac[2])
	case ac3Mode:
		eff += dg.PhysAddrT(cpuPtr.ac[3])
	}

	// handle indirection
	if ind == '@' { // down the rabbit hole...
		indAddr := memory.ReadDWord(eff)
		for memory.TestDwbit(indAddr, 0) {
			indAddr = memory.ReadDWord(dg.PhysAddrT(indAddr & physMask32))
		}
		eff = dg.PhysAddrT(indAddr)
	}

	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... resolve32bitEffAddr got: %#o, returning %#o\n", disp, eff)
	}
	return eff
}

func resolve32bitIndirectableAddr(iAddr dg.DwordT) dg.PhysAddrT {
	eff := iAddr
	// handle indirection
	for memory.TestDwbit(eff, 0) {
		eff = memory.ReadDWord(dg.PhysAddrT(eff & physMask32))
	}
	return dg.PhysAddrT(eff)
}

// resolveEclipseBitAddr as per page 10-8 of Pop
// Used by BTO, BTZ, SNB, SZB, SZBO
func resolveEclipseBitAddr(cpuPtr *CPUT, twoAcc1Word *twoAcc1WordT) (wordAddr dg.PhysAddrT, bitNum uint) {
	// TODO handle segments and indirection
	if twoAcc1Word.acd == twoAcc1Word.acs {
		wordAddr = 0
	} else {
		if memory.TestDwbit(cpuPtr.ac[twoAcc1Word.acs], 0) {
			log.Fatal("ERROR: Indirect 16-bit BIT pointers not yet supported")
		}
		wordAddr = dg.PhysAddrT(cpuPtr.ac[twoAcc1Word.acs]) & physMask16 // mask off lower 15 bits
	}
	offset := dg.PhysAddrT(cpuPtr.ac[twoAcc1Word.acd]) >> 4
	wordAddr += offset // add unsigned offset
	bitNum = uint(cpuPtr.ac[twoAcc1Word.acd] & 0x000f)
	return wordAddr, bitNum
}

// resolveEagleeBitAddr as per page 1-17 of Pop
// Used by eg. WSZB
func resolveEagleBitAddr(cpuPtr *CPUT, twoAcc1Word *twoAcc1WordT) (wordAddr dg.PhysAddrT, bitNum uint) {
	// TODO handle segments and indirection
	if twoAcc1Word.acd == twoAcc1Word.acs {
		wordAddr = 0
	} else {
		if memory.TestDwbit(cpuPtr.ac[twoAcc1Word.acs], 0) {
			log.Fatal("ERROR: Indirect 32-bit BIT pointers not yet supported")
		}
		wordAddr = dg.PhysAddrT(cpuPtr.ac[twoAcc1Word.acs])
	}
	offset := dg.PhysAddrT(cpuPtr.ac[twoAcc1Word.acd]) >> 4
	wordAddr += offset // add unsigned offset
	bitNum = uint(cpuPtr.ac[twoAcc1Word.acd] & 0x000f)
	return wordAddr, bitNum
}
