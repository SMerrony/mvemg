// mvemg project eagleOp_test.go

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

import "testing"

func TestNADD(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var twoAcc1Word twoAcc1WordT
	iPtr.ix = instrNADD
	twoAcc1Word.acs = 0
	twoAcc1Word.acd = 1
	// test neg + neg
	cpuPtr.ac[0] = 0xffff // -1
	cpuPtr.ac[1] = 0xffff // -1
	iPtr.variant = twoAcc1Word
	if !eagleOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute NADD")
	}
	if cpuPtr.ac[1] != 0xfffffffe {
		t.Errorf("Expected %x, got %x", 0xfffffffe, cpuPtr.ac[1])
	}

	// test neg + pos
	cpuPtr.ac[0] = 0x0001 // -1
	cpuPtr.ac[1] = 0xffff // -1

	if !eagleOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute NADD")
	}
	if cpuPtr.ac[1] != 0 {
		t.Errorf("Expected %x, got %x", 0, cpuPtr.ac[1])
	}
}

func TestNSUB(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var twoAcc1Word twoAcc1WordT
	iPtr.ix = instrNSUB
	twoAcc1Word.acs = 0
	twoAcc1Word.acd = 1
	iPtr.variant = twoAcc1Word
	// test neg - neg
	cpuPtr.ac[0] = 0xffff // -1
	cpuPtr.ac[1] = 0xffff // -1

	if !eagleOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute NSUB")
	}
	if cpuPtr.ac[1] != 0 {
		t.Errorf("Expected %x, got %x", 0, cpuPtr.ac[1])
	}

	// test neg - pos
	cpuPtr.ac[0] = 0x0001 // 1
	cpuPtr.ac[1] = 0xffff // -1

	if !eagleOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute NADD")
	}
	if cpuPtr.ac[1] != 0xfffffffe {
		t.Errorf("Expected %x, got %x", 0xfffffffe, cpuPtr.ac[1])
	}
}

func TestWANDI(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var oneAccImmDwd3Word oneAccImmDwd3WordT
	iPtr.ix = instrWANDI
	oneAccImmDwd3Word.immDword = 0x7fffffff
	oneAccImmDwd3Word.acd = 0
	iPtr.variant = oneAccImmDwd3Word
	cpuPtr.ac[0] = 0x3171
	if !eagleOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute WANDI")
	}
	if cpuPtr.ac[0] != 0x3171 {
		t.Errorf("Expected %x, got %x", 0x3171, cpuPtr.ac[0])
	}
	oneAccImmDwd3Word.immDword = 0x7fffffff
	oneAccImmDwd3Word.acd = 0
	iPtr.variant = oneAccImmDwd3Word
	cpuPtr.ac[0] = 0x20202020
	if !eagleOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute WANDI")
	}
	if cpuPtr.ac[0] != 0x20202020 {
		t.Errorf("Expected %x, got %x", 0x20202020, cpuPtr.ac[0])
	}
}

func TestWNEG(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var twoAcc1Word twoAcc1WordT
	iPtr.ix = instrWNEG
	twoAcc1Word.acs = 0
	twoAcc1Word.acd = 1
	iPtr.variant = twoAcc1Word
	cpuPtr.ac[0] = 37
	// test cpnversion to negative
	if !eagleOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute WNEG")
	}
	if cpuPtr.ac[1] != 0xffffffdb {
		t.Errorf("Expected 0xffffffdb, got %x", cpuPtr.ac[1])
	}

	// convert back to test conversion from negative
	cpuPtr.ac[0] = cpuPtr.ac[1]
	if !eagleOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute WNEG")
	}
	if cpuPtr.ac[1] != 37 {
		t.Errorf("Expected 37, got %d", cpuPtr.ac[1])
	}
}
