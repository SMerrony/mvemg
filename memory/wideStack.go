// wideStack.go

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
	"mvemg/dg"
	"mvemg/logging"
)

const (
	// Some Page Zero special locations...

	WFP_LOC  = 020
	WSP_LOC  = 022
	WSL_LOC  = 024
	WSB_LOC  = 026
	WPFH_LOC = 030
	CBP_LOC  = 032
)

// PUSH a doubleword onto the Wide Stack
func WsPush(seg dg.PhysAddrT, data dg.DwordT) {
	// TODO segment handling
	// TODO overflow/underflow handling - either here or in instruction?
	wsp := ReadDWord(WSP_LOC) + 2
	WriteDWord(WSP_LOC, wsp)
	WriteDWord(dg.PhysAddrT(wsp), data)
	logging.DebugPrint(logging.DebugLog, "... memory.WsPush pushed %8d onto the Wide Stack at location: %d\n", data, wsp)
}

// POP a word off the Wide Stack
func WsPop(seg dg.PhysAddrT) dg.DwordT {
	// TODO segment handling
	// TODO overflow/underflow handling - either here or in instruction?
	wsp := ReadDWord(WSP_LOC)
	dword := ReadDWord(dg.PhysAddrT(wsp))
	WriteDWord(WSP_LOC, wsp-2)
	logging.DebugPrint(logging.DebugLog, "... memory.WsPop  popped %8d off  the Wide Stack at location: %d\n", dword, wsp)
	return dword
}

// AdvanceWSP increases the WSP by the given amount of DWords
func AdvanceWSP(dwdCnt uint) {
	wsp := ReadDWord(WSP_LOC) + dg.DwordT(dwdCnt*2)
	WriteDWord(WSP_LOC, wsp)
	logging.DebugPrint(logging.DebugLog, "... WSP advanced by %d DWords to %d\n", dwdCnt, wsp)
}
