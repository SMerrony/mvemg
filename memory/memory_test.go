// mvemg project memory_test.go

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

package memory

import "testing"

func Testutil.DWordFromTwoWords(t *testing.T) {
	var hidg.WordT = 0x1122
	var lodg.WordT = 0x3344
	r := util.DWordFromTwoWords(hi, lo)
	if r != 0x11223344 {
		t.Error("Expected 287454020, got ", r)
	}
}

func Testutil.GetWbits(t *testing.T) {
	var wdg.WordT = 0xb38f
	r := util.GetWbits(w, 5, 3)
	if r != 3 {
		t.Error("Expected 3, got ", r)
	}

}

func Testutil.SetWbit(t *testing.T) {
	var wdg.WordT = 0x0001
	r := util.SetWbit(w, 1)
	if r != 0x4001 {
		t.Error("Expected 16385, got ", r)
	}
	// repeat - should have no effect
	r = util.SetWbit(w, 1)
	if r != 0x4001 {
		t.Error("Expected 16385, got ", r)
	}
}

func Testutil.ClearWbit(t *testing.T) {
	var wdg.WordT = 0x4001
	r := util.ClearWbit(w, 1)
	if r != 1 {
		t.Error("Expected 1, got ", r)
	}
	// repeat - should have no effect
	r = util.ClearWbit(w, 1)
	if r != 1 {
		t.Error("Expected 16385, got ", r)
	}
}
func Testutil.GetDWbits(t *testing.T) {
	var wdg.DwordT = 0xb38f00ff
	r := util.GetDWbits(w, 15, 2)
	if r != 2 {
		t.Error("Expected 2, got ", r)
	}

}

func TestMemWriteReadWord(t *testing.T) {
	var wdg.WordT
	WriteWord(78, 99)
	w = memory.ram[78]
	if w != 99 {
		t.Error("Expected 99, got ", w)
	}

	w = memory.ReadWord(78)
	if w != 99 {
		t.Error("Expected 99, got ", w)
	}
}

func TestWriteReadByte(t *testing.T) {
	var wdg.WordT
	var bdg.ByteT
	WriteByte(73, false, 0x58)
	w = memory.ram[73]
	if w != 0x5800 {
		t.Error("Expected 0x5800, got ", w)
	}
	WriteByte(74, true, 0x58)
	w = memory.ram[74]
	if w != 0x58 {
		t.Error("Expected 0x58, got ", w)
	}

	WriteWord(73, 0x11dd)
	b = ReadByte(73, true)
	if b != 0xdd {
		t.Error("Expected 0xDD, got ", b)
	}
	b = ReadByte(73, false)
	if b != 0x11 {
		t.Error("Expected 0x11, got ", b)
	}
}

func TestWriteReadDWord(t *testing.T) {
	var dwddg.DwordT
	WriteDWord(68, 0x11223344)
	w := memory.ram[68]
	if w != 0x1122 {
		t.Error("Expected 0x1122, got ", w)
	}
	w = memory.ram[69]
	if w != 0x3344 {
		t.Error("Expected 0x3344, got ", w)
	}
	dwd = ReadDWord(68)
	if dwd != 0x11223344 {
		t.Error("Expected 0x11223344, got", dwd)
	}
}

func Testutil.SexWordToDWord(t *testing.T) {
	var wddg.WordT
	var dwddg.DwordT

	wd = 79
	dwd = util.SexWordToDWord(wd)
	if dwd != 79 {
		t.Error("Expected 79, got ", dwd)
	}

	wd = 0xfff4
	dwd = util.SexWordToDWord(wd)
	if dwd != 0xfffffff4 {
		t.Error("Expected -12, got ", dwd)
	}

}

func Testutil.SwapBytes(t *testing.T) {
	var wd, sdg.WordT

	wd = 0x1234
	s = util.SwapBytes(wd)
	if s != 0x3412 {
		t.Error("Expected 13330., got ", s)
	}

	wd = util.SwapBytes(s)
	if wd != 0x1234 {
		t.Error("Expected 4660., got ", wd)
	}
}
