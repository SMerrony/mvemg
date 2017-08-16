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

import (
	"log"
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
)

func eagleStack(cpuPtr *CPU, iPtr *decodedInstrT) bool {

	var (
		firstAc, lastAc, thisAc int
		acsUp                   = [8]int{0, 1, 2, 3, 0, 1, 2, 3}
		tmpDwd                  dg.DwordT
	)

	switch iPtr.mnemonic {

	case "LDAFP":
		cpuPtr.ac[iPtr.acd] = memory.ReadDWord(memory.WfpLoc)

	case "LDASB":
		cpuPtr.ac[iPtr.acd] = memory.ReadDWord(memory.WsbLoc)

	case "LDASL":
		cpuPtr.ac[iPtr.acd] = memory.ReadDWord(memory.WslLoc)

	case "LDASP":
		cpuPtr.ac[iPtr.acd] = memory.ReadDWord(memory.WspLoc)

	case "LPEF":
		memory.WsPush(0, dg.DwordT(resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp31)))

	case "STAFP":
		// FIXME handle segments
		memory.WriteDWord(memory.WfpLoc, cpuPtr.ac[iPtr.acd])

	case "STASB":
		// FIXME handle segments
		memory.WriteDWord(memory.WsbLoc, cpuPtr.ac[iPtr.acd])

	case "STASL":
		// FIXME handle segments
		memory.WriteDWord(memory.WslLoc, cpuPtr.ac[iPtr.acd])

	case "STASP":
		// FIXME handle segments
		memory.WriteDWord(memory.WspLoc, cpuPtr.ac[iPtr.acd])
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "... STASP set WSP to %d\n", cpuPtr.ac[iPtr.acd])
		}

	case "STATS":
		// FIXME handle segments
		memory.WriteDWord(dg.PhysAddrT(memory.ReadDWord(memory.WslLoc)), cpuPtr.ac[iPtr.acd])

	case "WMSP":
		tmpDwd = cpuPtr.ac[iPtr.acd] << 1
		tmpDwd += memory.ReadDWord(memory.WspLoc)
		memory.WriteDWord(memory.WspLoc, tmpDwd)

	case "WPOP":
		firstAc = iPtr.acs
		lastAc = iPtr.acd
		if lastAc > firstAc {
			firstAc += 4
		}
		for thisAc = firstAc; thisAc >= lastAc; thisAc-- {
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "... wide popping AC%d\n", acsUp[thisAc])
			}
			cpuPtr.ac[acsUp[thisAc]] = memory.WsPop(0)
		}

	case "WPSH":
		firstAc = iPtr.acs
		lastAc = iPtr.acd
		if lastAc < firstAc {
			lastAc += 4
		}
		for thisAc = firstAc; thisAc <= lastAc; thisAc++ {
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "... wide pushing AC%d\n", acsUp[thisAc])
			}
			memory.WsPush(0, cpuPtr.ac[acsUp[thisAc]])
		}

	// N.B. WRTN is in eaglePC

	case "WSAVR":
		wsav(cpuPtr, iPtr)
		cpu.ovk = false

	case "WSAVS":
		wsav(cpuPtr, iPtr)
		cpu.ovk = true

	case "XPEF":
		// FIXME segment handling, check for overflow
		memory.WsPush(0, dg.DwordT(resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp15)))

	default:
		log.Fatalf("ERROR: EAGLE_STACK instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg.PhysAddrT(iPtr.instrLength)
	return true
}

// wsav is common to WSAVR and WSAVS
func wsav(cpuPtr *CPU, iPtr *decodedInstrT) {
	wfpSav := memory.ReadDWord(memory.WfpLoc)
	memory.WsPush(0, cpuPtr.ac[0]) // 1
	memory.WsPush(0, cpuPtr.ac[1]) // 2
	memory.WsPush(0, cpuPtr.ac[2]) // 3
	memory.WsPush(0, wfpSav)       // 4
	dwd := cpuPtr.ac[3] & 0x7fffffff
	if cpuPtr.carry {
		dwd |= 0x80000000
	}
	memory.WsPush(0, dwd) // 5
	dwdCnt := uint(iPtr.immU16)
	if dwdCnt > 0 {
		// for d := 0; d < dwdCnt; d++ {
		// 	memory.WsPush(0, 0)
		// }
		memory.AdvanceWSP(dwdCnt)
	}
	memory.WriteDWord(memory.WfpLoc, memory.ReadDWord(memory.WspLoc))
	cpuPtr.ac[3] = memory.ReadDWord(memory.WspLoc)
}
