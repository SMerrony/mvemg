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

func TestHXL(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	iPtr.mnemonic = "HXL"
	iPtr.acd = 0
	cpuPtr.ac[0] = 0x0123
	iPtr.immU16 = 2
	expd := dg_dword(0x2300)
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute HXL")
	}
	if cpuPtr.ac[0] != expd {
		t.Errorf("Expected %x, got %x", expd, cpuPtr.ac[0])
	}

	cpuPtr.ac[0] = 0x0123
	iPtr.immU16 = 4
	expd = dg_dword(0x0)
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
	iPtr.mnemonic = "HXR"
	iPtr.acd = 0
	cpuPtr.ac[0] = 0x0123
	iPtr.immU16 = 2
	expd := dg_dword(0x0001)
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute HXL")
	}
	if cpuPtr.ac[0] != expd {
		t.Errorf("Expected %x, got %x", expd, cpuPtr.ac[0])
	}

	cpuPtr.ac[0] = 0x0123
	iPtr.immU16 = 4
	expd = dg_dword(0x0)
	if !eclipseOp(cpuPtr, &iPtr) {
		t.Error("Failed to execute HXL")
	}
	if cpuPtr.ac[0] != expd {
		t.Errorf("Expected %x, got %x", expd, cpuPtr.ac[0])
	}
}
