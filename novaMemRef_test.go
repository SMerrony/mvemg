// novaMemRef_test.go

// Copyright (C) 2019 Steve Merrony

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

	"github.com/SMerrony/dgemug/memory"
)

func TestDSZ(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var novaNoAccEffAddr novaNoAccEffAddrT
	iPtr.ix = instrDSZ
	memory.MemInit(10000, false)
	memory.WriteWord(500, 2)
	cpuPtr.pc = 10
	novaNoAccEffAddr.disp15 = 500
	novaNoAccEffAddr.ind = ' '
	novaNoAccEffAddr.mode = absoluteMode
	iPtr.variant = novaNoAccEffAddr

	if !novaMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute DSZ")
	}
	if cpuPtr.pc != 11 {
		t.Errorf("Expected PC 11, got %d.", cpuPtr.pc)
	}
	w := memory.ReadWord(500)
	if w != 1 {
		t.Errorf("Expected loc 500 to contain 1, got: %x", w)
	}

	if !novaMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute DSZ")
	}
	if cpuPtr.pc != 13 {
		t.Errorf("Expected PC 13, got %d.", cpuPtr.pc)
	}
	w = memory.ReadWord(500)
	if w != 0 {
		t.Errorf("Expected loc 500 to contain 0, got: %x", w)
	}
}

func TestISZ(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var novaNoAccEffAddr novaNoAccEffAddrT
	iPtr.ix = instrISZ
	memory.MemInit(10000, false)
	memory.WriteWord(500, 0xfffe)
	cpuPtr.pc = 10
	novaNoAccEffAddr.disp15 = 500
	novaNoAccEffAddr.ind = ' '
	novaNoAccEffAddr.mode = absoluteMode
	iPtr.variant = novaNoAccEffAddr

	if !novaMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute ISZ")
	}
	if cpuPtr.pc != 11 {
		t.Errorf("Expected PC 11, got %d.", cpuPtr.pc)
	}
	w := memory.ReadWord(500)
	if w != 0xffff {
		t.Errorf("Expected loc 500 to contain 0xffff, got: %x", w)
	}

	if !novaMemRef(cpuPtr, &iPtr) {
		t.Error("Failed to execute ISZ")
	}
	if cpuPtr.pc != 13 {
		t.Errorf("Expected PC 13, got %d.", cpuPtr.pc)
	}
	w = memory.ReadWord(500)
	if w != 0 {
		t.Errorf("Expected loc 500 to contain 0, got: %x", w)
	}
}
