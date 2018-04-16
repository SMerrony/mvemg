// mvemg project eaglePC_test.go

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
	"testing"

	"github.com/SMerrony/dgemug/memory"
)

func TestXWDSZ(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var noAccModeInd2Word noAccModeInd2WordT
	iPtr.ix = instrXWDSZ
	memory.MemInit(10000, false)
	memory.WriteDWord(100, 2) // write 2 into Word at normal addr 100
	noAccModeInd2Word.disp15 = 100
	noAccModeInd2Word.ind = ' '
	noAccModeInd2Word.mode = absoluteMode
	iPtr.variant = noAccModeInd2Word
	cpuPtr.pc = 1000
	if !eaglePC(cpuPtr, &iPtr) {
		t.Error("Failed to execute XWDSZ")
	}
	// 1st time should simply decrement contents
	if cpuPtr.pc != 1002 {
		t.Errorf("Expected PC: 1002., got %d.", cpuPtr.pc)
	}
	w := memory.ReadDWord(100)
	if w != 1 {
		t.Errorf("Expected loc 100. to contain 1, got %d", w)
	}
	// 2nd time should dec and skip
	if !eaglePC(cpuPtr, &iPtr) {
		t.Error("Failed to execute XWDSZ")
	}
	if cpuPtr.pc != 1005 {
		t.Errorf("Expected PC: 1005., got %d.", cpuPtr.pc)
	}
	w = memory.ReadDWord(100)
	if w != 0 {
		t.Errorf("Expected loc 100. to contain 0, got %d", w)
	}
}