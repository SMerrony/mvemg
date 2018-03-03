// mvemg project eagleMemRef_test.go

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

func TestXSTB(t *testing.T) {
	cpuPtr := cpuInit(nil)
	var iPtr decodedInstrT
	var oneAccMode2Word oneAccMode2WordT
	iPtr.ix = instrXSTB
	memory.MemInit(10000, false)
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
	w := memory.ReadWord(7)
	if w != 0x0044 {
		t.Errorf("Expected %d, got %d", 0x0044, w)
	}
}
