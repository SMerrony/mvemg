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

////////////////////////////////////////////////////////////////
// N.B. Be sure to use Double-Word memory references here... //
//////////////////////////////////////////////////////////////

package main

import "log"

func eagleStack(cpuPtr *CPU, iPtr *decodedInstrT) bool {

	switch iPtr.mnemonic {

	case "LDAFP":
		cpuPtr.ac[iPtr.acd] = memReadDWord(WFP_LOC)

	case "LDASB":
		cpuPtr.ac[iPtr.acd] = memReadDWord(WSB_LOC)

	case "LDASL":
		cpuPtr.ac[iPtr.acd] = memReadDWord(WSL_LOC)

	case "LDASP":
		cpuPtr.ac[iPtr.acd] = memReadDWord(WSP_LOC)

	case "STAFP":
		// FIXME handle segments
		memWriteDWord(WFP_LOC, cpuPtr.ac[iPtr.acd])

	case "STASB":
		// FIXME handle segments
		memWriteDWord(WSB_LOC, cpuPtr.ac[iPtr.acd])

	case "STASL":
		// FIXME handle segments
		memWriteDWord(WSL_LOC, cpuPtr.ac[iPtr.acd])

	case "STASP":
		// FIXME handle segments
		memWriteDWord(WSP_LOC, cpuPtr.ac[iPtr.acd])

	case "STATS":
		// FIXME handle segments
		memWriteDWord(dg_phys_addr(memReadDWord(WSL_LOC)), cpuPtr.ac[iPtr.acd])

	case "WSAVR":
		wsav(cpuPtr, iPtr)
		cpu.ovk = false

	case "WSAVS":
		wsav(cpuPtr, iPtr)
		cpu.ovk = true

	default:
		log.Fatalf("ERROR: EAGLE_STACK instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}

// wsav is common to WSAVR and WSAVS
func wsav(cpuPtr *CPU, iPtr *decodedInstrT) {
	wfpSav := memReadDWord(WFP_LOC)
	wsPush(0, cpuPtr.ac[0]) // 1
	wsPush(0, cpuPtr.ac[1]) // 2
	wsPush(0, cpuPtr.ac[2]) // 3
	wsPush(0, wfpSav)       // 4
	dwd := cpuPtr.ac[3] & 0x7fffffff
	if cpuPtr.carry {
		dwd |= 0x80000000
	}
	wsPush(0, dwd) // 5
	dwdCnt := int(iPtr.immU16)
	if dwdCnt > 0 {
		for d := 0; d < dwdCnt; d++ {
			wsPush(0, 0)
		}
	}
	memWriteDWord(WFP_LOC, memReadDWord(WSP_LOC))
	cpuPtr.ac[3] = memReadDWord(WSP_LOC)
}
