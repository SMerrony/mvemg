// mvemg project ram_test.go

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
// THE SOFTWARE.package memory

package memory

import (
	"mvemg/dg"
	"testing"
)

func TestWriteReadByte(t *testing.T) {
	var w dg.WordT
	var b dg.ByteT
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

func TestWriteReadWord(t *testing.T) {
	var w dg.WordT
	WriteWord(78, 99)
	w = memory.ram[78]
	if w != 99 {
		t.Error("Expected 99, got ", w)
	}

	w = ReadWord(78)
	if w != 99 {
		t.Error("Expected 99, got ", w)
	}
}
func TestWriteReadDWord(t *testing.T) {
	var dwd dg.DwordT
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