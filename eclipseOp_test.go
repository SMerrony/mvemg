// mvemg project eclipseOp_test.go

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

import "mvemg/dg"

func TestDIV(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	iPtr.ix = instrDIV
	cpuPtr.ac[0] = 0 // hi dividend
	cpuPtr.ac[1] = 6 // lo dividend
	cpuPtr.ac[2] = 2 // divisor
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute DIV")
	}
	if cpuPtr.ac[1] != 3 || cpuPtr.ac[0] != 0 || cpuPtr.ac[2] != 2 {
		t.Errorf("Expected 3, 0, 2, got: %d, %d, %d",
			cpuPtr.ac[1], cpuPtr.ac[0], cpuPtr.ac[2])
	}

	cpuPtr.ac[0] = 0 // hi dividend
	cpuPtr.ac[1] = 6 // lo dividend
	cpuPtr.ac[2] = 4 // divisor
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute DIV")
	}
	if cpuPtr.ac[1] != 1 || cpuPtr.ac[0] != 2 || cpuPtr.ac[2] != 4 {
		t.Errorf("Expected 1, 2, 4, got: %d, %d, %d",
			cpuPtr.ac[1], cpuPtr.ac[0], cpuPtr.ac[2])
	}

	cpuPtr.ac[0] = 0      // hi dividend
	cpuPtr.ac[1] = 0xf000 // lo dividend
	cpuPtr.ac[2] = 2      // divisor
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute DIV")
	}
	if cpuPtr.ac[1] != 0x7800 || cpuPtr.ac[0] != 0 || cpuPtr.ac[2] != 2 {
		t.Errorf("Expected 30720, 0, 2, got: %d, %d, %d",
			cpuPtr.ac[1], cpuPtr.ac[0], cpuPtr.ac[2])
	}

	cpuPtr.ac[0] = 1      // hi dividend
	cpuPtr.ac[1] = 0xf000 // lo dividend
	cpuPtr.ac[2] = 2      // divisor
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute DIV")
	}
	if cpuPtr.ac[1] != 0xf800 || cpuPtr.ac[0] != 0 || cpuPtr.ac[2] != 2 {
		t.Errorf("Expected 63488, 0, 2, got: %d, %d, %d",
			cpuPtr.ac[1], cpuPtr.ac[0], cpuPtr.ac[2])
	}

	cpuPtr.ac[0] = 0xf000 // hi dividends- SHOULD CAUSE EXCEPTION
	cpuPtr.ac[1] = 0xf000 // lo dividend
	cpuPtr.ac[2] = 512    // divisor
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute DIV")
	}
	if cpuPtr.ac[1] != 61440 || cpuPtr.ac[0] != 61440 || cpuPtr.ac[2] != 512 {
		t.Errorf("Expected 61440, 61440, 512, got: %d, %d, %d",
			cpuPtr.ac[1], cpuPtr.ac[0], cpuPtr.ac[2])
	}
}

func TestHXL(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var immOneAcc immOneAccT
	iPtr.ix = instrHXL
	immOneAcc.acd = 0
	cpuPtr.ac[0] = 0x0123
	immOneAcc.immU16 = 2
	iPtr.variant = immOneAcc
	expd := dg.DwordT(0x2300)
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute HXL")
	}
	if cpuPtr.ac[0] != expd {
		t.Errorf("Expected %x, got %x", expd, cpuPtr.ac[0])
	}

	cpuPtr.ac[0] = 0x0123
	immOneAcc.immU16 = 4
	expd = dg.DwordT(0x0)
	iPtr.variant = immOneAcc
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute HXL")
	}
	if cpuPtr.ac[0] != expd {
		t.Errorf("Expected %x, got %x", expd, cpuPtr.ac[0])
	}
}
func TestHXR(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var immOneAcc immOneAccT
	iPtr.ix = instrHXR
	immOneAcc.acd = 0
	cpuPtr.ac[0] = 0x0123
	immOneAcc.immU16 = 2
	iPtr.variant = immOneAcc
	expd := dg.DwordT(0x0001)
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute HXL")
	}
	if cpuPtr.ac[0] != expd {
		t.Errorf("Expected %x, got %x", expd, cpuPtr.ac[0])
	}

	cpuPtr.ac[0] = 0x0123
	immOneAcc.immU16 = 4
	iPtr.variant = immOneAcc
	expd = dg.DwordT(0x0)
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute HXL")
	}
	if cpuPtr.ac[0] != expd {
		t.Errorf("Expected %x, got %x", expd, cpuPtr.ac[0])
	}
}

func TestSBI(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var immOneAcc immOneAccT
	iPtr.ix = instrSBI
	immOneAcc.immU16 = 3
	immOneAcc.acd = 0
	iPtr.variant = immOneAcc
	cpuPtr.ac[0] = 0xffff // 65535
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute SBI")
	}
	if cpuPtr.ac[0] != 65532 {
		t.Errorf("Expected %x, got %x", 65532, cpuPtr.ac[0])
	}

	// test 'negative' wraparound
	immOneAcc.immU16 = 3
	immOneAcc.acd = 0
	iPtr.variant = immOneAcc
	cpuPtr.ac[0] = 2
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute SBI")
	}
	if cpuPtr.ac[0] != 65535 {
		t.Errorf("Expected %x, got %x", 65535, cpuPtr.ac[0])
	}
}
