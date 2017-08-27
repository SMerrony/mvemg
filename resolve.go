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
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
	"mvemg/util"
)

func resolve16bitEclipseAddr(cpuPtr *CPUT, ind byte, mode string, disp int16) dg.PhysAddrT {

	var (
		eff     dg.PhysAddrT
		intEff  int32
		indAddr dg.WordT
	)

	// handle addressing mode...
	switch mode {
	case "Absolute":
		intEff = int32(disp)
	case "PC":
		intEff = int32(cpuPtr.pc) + int32(disp)
	case "AC2":
		intEff = int32(cpuPtr.ac[2]) + int32(disp)
	case "AC3":
		intEff = int32(cpuPtr.ac[3]) + int32(disp)
	}

	// handle indirection
	if ind == '@' { // down the rabbit hole...
		indAddr = memory.ReadWord(dg.PhysAddrT(intEff))
		for util.TestWbit(indAddr, 0) {
			indAddr = memory.ReadWord(dg.PhysAddrT(indAddr))
		}
		intEff = int32(indAddr)
	}

	// mask off to Eclipse range
	eff = dg.PhysAddrT(intEff) & 0x7fff

	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... resolve16bitEclipseAddr got: %d., returning %d.\n", disp, eff)
	}
	return eff
}

// This is the same as resolve16bitEclipseAddr, but without the range masking at the end
func resolve16bitEagleAddr(cpuPtr *CPUT, ind byte, mode string, disp int16) dg.PhysAddrT {

	var (
		eff     dg.PhysAddrT
		intEff  int32
		indAddr dg.DwordT
	)

	// handle addressing mode...
	switch mode {
	case "Absolute":
		intEff = int32(disp)
	case "PC":
		intEff = int32(cpuPtr.pc) + int32(disp)
	case "AC2":
		intEff = int32(cpuPtr.ac[2]) + int32(disp)
	case "AC3":
		intEff = int32(cpuPtr.ac[3]) + int32(disp)
	}

	// handle indirection
	if ind == '@' { // down the rabbit hole...
		indAddr = memory.ReadDWord(dg.PhysAddrT(intEff))
		for util.TestDWbit(indAddr, 0) {
			indAddr = memory.ReadDWord(dg.PhysAddrT(indAddr))
		}
		intEff = int32(indAddr)
	}

	eff = dg.PhysAddrT(intEff)

	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... resolve16bitEagleAddr got: %d., returning %d.\n", disp, eff)
	}
	return eff
}

// Resolve32bitByteAddr returns the word address and low-byte flag for a given 32-bit byte address
func resolve32bitByteAddr(byteAddr dg.DwordT) (wordAddr dg.PhysAddrT, loByte bool) {
	wa := dg.PhysAddrT(byteAddr) >> 1
	lb := util.TestDWbit(byteAddr, 31)
	return wa, lb
}

func resolve32bitEffAddr(cpuPtr *CPUT, ind byte, mode string, disp int32) dg.PhysAddrT {

	eff := dg.PhysAddrT(disp)

	// handle addressing mode...
	switch mode {
	case "Absolute":
		// nothing to do
	case "PC":
		eff += dg.PhysAddrT(cpuPtr.pc)
	case "AC2":
		eff += dg.PhysAddrT(cpuPtr.ac[2])
	case "AC3":
		eff += dg.PhysAddrT(cpuPtr.ac[3])
	}

	// handle indirection
	if ind == '@' { // down the rabbit hole...
		indAddr := memory.ReadDWord(eff)
		for util.TestDWbit(indAddr, 0) {
			indAddr = memory.ReadDWord(dg.PhysAddrT(indAddr))
		}
		eff = dg.PhysAddrT(indAddr)
	}

	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... resolve32bitEffAddr got: %d., returning %d.\n", disp, eff)
	}
	return eff
}

// resolveEclipseBitAddr as per page 10-8 of Pop
// Used by BTO, BTZ, SNB, SZB, SZBO
func resolveEclipseBitAddr(cpuPtr *CPUT, twoAcc1Word *twoAcc1WordT) (wordAddr dg.PhysAddrT, bitNum uint) {
	// TODO handle segments and indirection
	if twoAcc1Word.acd == twoAcc1Word.acs {
		wordAddr = 0
	} else {
		if util.TestDWbit(cpuPtr.ac[twoAcc1Word.acs], 0) {
			log.Fatal("ERROR: Indirect 16-bit BIT pointers not yet supported")
		}
		wordAddr = dg.PhysAddrT(cpuPtr.ac[twoAcc1Word.acs]) & 0x7fff // mask off lower 15 bits
	}
	offset := dg.PhysAddrT(cpuPtr.ac[twoAcc1Word.acd]) >> 4
	wordAddr += offset // add unsigned offset
	bitNum = uint(cpuPtr.ac[twoAcc1Word.acd] & 0x000f)
	return wordAddr, bitNum
}
