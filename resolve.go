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
	"mvemg/logging"
)

func resolve16bitEclipseAddr(cpuPtr *CPU, ind byte, mode string, disp int16) DgPhysAddrT {

	var (
		eff     DgPhysAddrT
		intEff  int32
		indAddr DgWordT
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
		indAddr = memReadWord(DgPhysAddrT(intEff))
		for testWbit(indAddr, 0) {
			indAddr = memReadWord(DgPhysAddrT(indAddr))
		}
		intEff = int32(indAddr)
	}

	// mask off to Eclipse range
	eff = DgPhysAddrT(intEff) & 0x7fff

	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... resolve16bitEclipseAddr got: %d., returning %d.\n", disp, eff)
	}
	return eff
}

// This is the same as resolve16bitEclipseAddr, but without the range masking at the end
func resolve16bitEagleAddr(cpuPtr *CPU, ind byte, mode string, disp int16) DgPhysAddrT {

	var (
		eff     DgPhysAddrT
		intEff  int32
		indAddr DgDwordT
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
		indAddr = memReadDWord(DgPhysAddrT(intEff))
		for testDWbit(indAddr, 0) {
			indAddr = memReadDWord(DgPhysAddrT(indAddr))
		}
		intEff = int32(indAddr)
	}

	eff = DgPhysAddrT(intEff)

	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... resolve16bitEagleAddr got: %d., returning %d.\n", disp, eff)
	}
	return eff
}

// Resolve32bitByteAddr returns the word address and low-byte flag for a given 32-bit byte address
func resolve32bitByteAddr(byteAddr DgDwordT) (wordAddr DgPhysAddrT, loByte bool) {
	wa := DgPhysAddrT(byteAddr) >> 1
	lb := testDWbit(byteAddr, 31)
	return wa, lb
}

func resolve32bitEffAddr(cpuPtr *CPU, ind byte, mode string, disp int32) DgPhysAddrT {

	eff := DgPhysAddrT(disp)

	// handle addressing mode...
	switch mode {
	case "Absolute":
		// nothing to do
	case "PC":
		eff += DgPhysAddrT(cpuPtr.pc)
	case "AC2":
		eff += DgPhysAddrT(cpuPtr.ac[2])
	case "AC3":
		eff += DgPhysAddrT(cpuPtr.ac[3])
	}

	// handle indirection
	if ind == '@' { // down the rabbit hole...
		indAddr := memReadDWord(eff)
		for testDWbit(indAddr, 0) {
			indAddr = memReadDWord(DgPhysAddrT(indAddr))
		}
		eff = DgPhysAddrT(indAddr)
	}

	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... resolve32bitEffAddr got: %d., returning %d.\n", disp, eff)
	}
	return eff
}

// resolveEclipseBitAddr as per page 10-8 of Pop
// Used by BTO, BTZ, SNB, SZB, SZBO
func resolveEclipseBitAddr(cpuPtr *CPU, iPtr *decodedInstrT) (wordAddr DgPhysAddrT, bitNum uint) {
	// TODO handle segments and indirection
	if iPtr.acd == iPtr.acs {
		wordAddr = 0
	} else {
		if testDWbit(cpuPtr.ac[iPtr.acs], 0) {
			log.Fatal("ERROR: Indirect 16-bit BIT pointers not yet supported")
		}
		wordAddr = DgPhysAddrT(cpuPtr.ac[iPtr.acs]) & 0x7fff // mask off lower 15 bits
	}
	offset := (DgPhysAddrT(cpuPtr.ac[iPtr.acd]) & 0x0000fff0) >> 4
	wordAddr += offset // add unsigned offset
	bitNum = uint(cpuPtr.ac[iPtr.acd] & 0x000f)
	return wordAddr, bitNum
}
