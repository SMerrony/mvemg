// mvemg project resolve_test.go

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

	"mvemg/memory"
)

func TestResolve16bitEclipseAddr(t *testing.T) {
	cpuPtr := cpuInit(nil)
	cpuPtr.pc = 11
	memory.WriteWord(10, 9)
	memory.WriteWord(11, 10)
	memory.WriteWord(12, 12)

	r := resolve16bitEclipseAddr(cpuPtr, ' ', "Absolute", 11)
	if r != 11 {
		t.Error("Expected 11, got ", r)
	}

	r = resolve16bitEclipseAddr(cpuPtr, ' ', "PC", 1)
	if r != 12 {
		t.Error("Expected 12, got ", r)
	}

	r = resolve16bitEclipseAddr(cpuPtr, ' ', "PC", -1)
	if r != 10 {
		t.Error("Expected 10, got ", r)
	}

	r = resolve16bitEclipseAddr(cpuPtr, '@', "PC", -1)
	if r != 9 {
		t.Error("Expected 9, got ", r)
	}

	cpuPtr.ac[2] = 12
	r = resolve16bitEclipseAddr(cpuPtr, '@', "AC2", -1)
	if r != 10 {
		t.Error("Expected 10, got ", r)
	}
}
