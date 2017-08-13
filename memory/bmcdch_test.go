// bmcdch_test.go

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

import (
	"testing"
)

func TestBmcdchReset(t *testing.T) {
	var wddg.WordT
	bmcdchInit()
	wd = regs[IOCHAN_DEF_REG]
	if wd != IOCDR_1 {
		t.Error("Got ", wd)
	}
}

func TestWriteReadMapSlot(t *testing.T) {
	var dwd1, dwd2dg.DwordT
	dwd1 = 0x11223344
	bmcdchWriteSlot(17, dwd1)
	dwd2 = bmcdchReadSlot(17)
	if dwd2 != 0x11223344 {
		t.Error("Expected 0x11223344, got ", dwd2)
	}

}

func TestBmcDchMapAddr(t *testing.T) {
	var addr1, addr2, pagedg.PhysAddrT
	bmcdchWriteSlot(0, 0)
	addr1 = 1
	addr2, page = getBmcDchMapAddr(addr1)
	if addr2 != 1 {
		t.Error("Expected 1, got ", addr2, page)
	}
	bmcdchWriteSlot(0, 3)
	addr1 = 1
	addr2, page = getBmcDchMapAddr(addr1)
	// 3 << 10 is 3072
	if addr2 != 3073 {
		t.Error("Expected 3073, got ", addr2, page)
	}
}
