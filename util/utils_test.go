// mvemg project util_test.go

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

package util

import (
	"mvemg/dg"
	"testing"
)

func TestDWordFromTwoWords(t *testing.T) {
	var hi dg.WordT = 0x1122
	var lo dg.WordT = 0x3344
	r := DWordFromTwoWords(hi, lo)
	if r != 0x11223344 {
		t.Error("Expected 287454020, got ", r)
	}
}

func TestGetWbits(t *testing.T) {
	var w dg.WordT = 0xb38f
	r := GetWbits(w, 5, 3)
	if r != 3 {
		t.Error("Expected 3, got ", r)
	}

}

func TestSetWbit(t *testing.T) {
	var w dg.WordT = 0x0001
	r := SetWbit(w, 1)
	if r != 0x4001 {
		t.Error("Expected 16385, got ", r)
	}
	// repeat - should have no effect
	r = SetWbit(w, 1)
	if r != 0x4001 {
		t.Error("Expected 16385, got ", r)
	}
}

func TestClearWbit(t *testing.T) {
	var w dg.WordT = 0x4001
	r := ClearWbit(w, 1)
	if r != 1 {
		t.Error("Expected 1, got ", r)
	}
	// repeat - should have no effect
	r = ClearWbit(w, 1)
	if r != 1 {
		t.Error("Expected 16385, got ", r)
	}
}
func TestGetDWbits(t *testing.T) {
	var w dg.DwordT = 0xb38f00ff
	r := GetDWbits(w, 15, 2)
	if r != 2 {
		t.Error("Expected 2, got ", r)
	}

}

func TestSexWordToDWord(t *testing.T) {
	var wd dg.WordT
	var dwd dg.DwordT

	wd = 79
	dwd = SexWordToDWord(wd)
	if dwd != 79 {
		t.Error("Expected 79, got ", dwd)
	}

	wd = 0xfff4
	dwd = SexWordToDWord(wd)
	if dwd != 0xfffffff4 {
		t.Error("Expected -12, got ", dwd)
	}

}

func TestSwapBytes(t *testing.T) {
	var wd, s dg.WordT

	wd = 0x1234
	s = SwapBytes(wd)
	if s != 0x3412 {
		t.Error("Expected 13330., got ", s)
	}

	wd = SwapBytes(s)
	if wd != 0x1234 {
		t.Error("Expected 4660., got ", wd)
	}
}
