// mvemg project eagleMemRef_test.go

// Copyright (C) 2017,2019 Steve Merrony

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
	"testing"

	"github.com/SMerrony/dgemug/dg"
	"github.com/SMerrony/dgemug/memory"
)

func TestWBTZ(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var twoAcc1Word twoAcc1WordT
	iPtr.ix = instrWBTZ
	memory.MemInit(10000, false)

	// case where acs == acd
	twoAcc1Word.acs = 0
	twoAcc1Word.acd = 0
	iPtr.variant = twoAcc1Word
	var wordOffset dg.DwordT = 73 << 4
	var bitNum dg.DwordT = 3
	cpuPtr.ac[0] = wordOffset | bitNum
	memory.WriteWord(73, 0xffff)
	if !eagleMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute WBTZ")
	}
	w := memory.ReadWord(73)
	if w != 0xefff {
		t.Errorf("Expected %x, got %x", 0xefff, w)
	}

	// case where acs != acd
	twoAcc1Word.acs = 1
	twoAcc1Word.acd = 0
	iPtr.variant = twoAcc1Word
	wordOffset = 33 << 4
	bitNum = 3
	cpuPtr.ac[0] = wordOffset | bitNum
	cpuPtr.ac[1] = 40
	memory.WriteWord(73, 0xffff)
	if !eagleMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute WBTZ")
	}
	w = memory.ReadWord(73)
	if w != 0xefff {
		t.Errorf("Expected %x, got %x", 0xefff, w)
	}

	// case where acs != acd and acs is indirect
	twoAcc1Word.acs = 1
	twoAcc1Word.acd = 0
	iPtr.variant = twoAcc1Word
	wordOffset = 33 << 4
	bitNum = 3
	cpuPtr.ac[0] = wordOffset | bitNum
	// put an indirect address in ac1 pointing to 60
	cpuPtr.ac[1] = 0x80000000 | 60
	// put 40 in location 60
	memory.WriteDWord(60, 40) // DWord!!!
	memory.WriteWord(73, 0xffff)
	if !eagleMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute WBTZ")
	}
	w = memory.ReadWord(73)
	if w != 0xefff {
		t.Errorf("Expected %x, got %x", 0xefff, w)
	}
}

func TestWCMV(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	iPtr.ix = instrWCMV
	memory.MemInit(1000, false)
	memory.WriteByte(100, false, 'A')
	memory.WriteByte(100, true, 'B')
	memory.WriteByte(101, false, 'C')
	memory.WriteByte(101, true, 'D')
	memory.WriteByte(102, false, 'E')
	memory.WriteByte(102, true, 'F')
	memory.WriteByte(103, false, 'G')

	// simple, word-aligned fwd move
	destNoBytes := 7
	srcNoBytes := 7
	destBytePtr := 200 << 1
	srcBytePtr := 100 << 1
	cpuPtr.ac[0] = dg.DwordT(destNoBytes)
	cpuPtr.ac[1] = dg.DwordT(srcNoBytes)
	cpuPtr.ac[2] = dg.DwordT(destBytePtr)
	cpuPtr.ac[3] = dg.DwordT(srcBytePtr)
	if !eagleMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute WCMV")
	}
	r := memory.ReadByte(200, false)
	if r != 'A' {
		t.Errorf("Expected 'A', got '%c'", r)
	}
	r = memory.ReadByte(200, true)
	if r != 'B' {
		t.Errorf("Expected 'B', got '%c'", r)
	}
	r = memory.ReadByte(202, true)
	if r != 'F' {
		t.Errorf("Expected 'F', got '%c'", r)
	}

	// non-word-aligned fwd move
	srcBytePtr++
	cpuPtr.ac[0] = dg.DwordT(destNoBytes)
	cpuPtr.ac[1] = dg.DwordT(srcNoBytes)
	cpuPtr.ac[2] = dg.DwordT(destBytePtr)
	cpuPtr.ac[3] = dg.DwordT(srcBytePtr)
	if !eagleMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute WCMV")
	}
	r = memory.ReadByte(200, false)
	if r != 'B' {
		t.Errorf("Expected 'B', got '%c'", r)
	}
	r = memory.ReadByte(200, true)
	if r != 'C' {
		t.Errorf("Expected 'C', got '%c'", r)
	}

	// src backwards
	srcBytePtr = 100 << 1
	destNoBytes = -7
	cpuPtr.ac[0] = dg.DwordT(destNoBytes)
	cpuPtr.ac[1] = dg.DwordT(srcNoBytes)
	cpuPtr.ac[2] = dg.DwordT(destBytePtr)
	cpuPtr.ac[3] = dg.DwordT(srcBytePtr)
	if !eagleMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute WCMV")
	}
	r = memory.ReadByte(200, false)
	if r != 'A' {
		t.Errorf("Expected 'A', got '%c'", r)
	}
	r = memory.ReadByte(199, true)
	if r != 'B' {
		t.Errorf("Expected 'B', got '%c'", r)
	}
	r = memory.ReadByte(199, false)
	if r != 'C' {
		t.Errorf("Expected 'C', got '%c'", r)
	}
}

func TestWLDB(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var twoAcc1Word twoAcc1WordT
	iPtr.ix = instrWLDB
	memory.MemInit(1000, false)
	memory.WriteByte(100, false, 'A')
	memory.WriteByte(100, true, 'B')
	twoAcc1Word.acs = 1
	twoAcc1Word.acd = 2
	cpuPtr.ac[1] = 100 << 1
	iPtr.variant = twoAcc1Word
	if !eagleMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute WLDB")
	}
	if cpuPtr.ac[2] != 'A' {
		t.Errorf("Expected %d, got %d", 'A', cpuPtr.ac[2])
	}
	twoAcc1Word.acs = 1
	twoAcc1Word.acd = 2
	cpuPtr.ac[1] = 100<<1 + 1
	iPtr.variant = twoAcc1Word
	if !eagleMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute WLDB")
	}
	if cpuPtr.ac[2] != 'B' {
		t.Errorf("Expected %d, got %d", 'B', cpuPtr.ac[2])
	}
}

func TestXSTB(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var oneAccMode2Word oneAccMode2WordT
	iPtr.ix = instrXSTB
	memory.MemInit(10000, false)

	// test high (left) byte write
	memory.WriteWord(7, 0) // write 0 into Word at normal addr 7
	oneAccMode2Word.disp16 = 7
	oneAccMode2Word.mode = absoluteMode
	oneAccMode2Word.bitLow = false
	oneAccMode2Word.acd = 1
	iPtr.variant = oneAccMode2Word
	cpuPtr.ac[1] = 0x11223344
	if !eagleMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute XSTB")
	}
	w := memory.ReadWord(7)
	if w != 0x4400 {
		t.Errorf("Expected %d, got %d", 0x4400, w)
	}

	// test low (right) byte write
	memory.WriteWord(7, 0) // write 0 into Word at normal addr 7
	oneAccMode2Word.disp16 = 7
	oneAccMode2Word.mode = absoluteMode
	oneAccMode2Word.bitLow = true
	oneAccMode2Word.acd = 1
	iPtr.variant = oneAccMode2Word
	cpuPtr.ac[1] = 0x11223344
	if !eagleMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute XSTB")
	}
	w = memory.ReadWord(7)
	if w != 0x0044 {
		t.Errorf("Expected %d, got %d", 0x0044, w)
	}
}
