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

func TestADDI(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	iPtr.mnemonic = "ADDI"
	iPtr.immS16 = 3
	iPtr.acd = 0
	cpuPtr.ac[0] = 0xffff // -1
	if !eagleOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute ADDI")
	}
	if cpuPtr.ac[0] != 2 {
		t.Errorf("Expected %x, got %x", 2, cpuPtr.ac[0])
	}

	iPtr.immS16 = -3
	iPtr.acd = 0
	cpuPtr.ac[0] = 0xffff // -1
	if !eagleOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute ADDI")
	}
	if cpuPtr.ac[0] != 0xfffc {
		t.Errorf("Expected %x, got %x", -4, cpuPtr.ac[0])
	}
}

func TestWNEG(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	iPtr.mnemonic = "WNEG"
	iPtr.acs = 0
	iPtr.acd = 1
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
