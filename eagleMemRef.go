// eagleMemRef.go

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
	"log"
	"mvemg/logging"
)

func eagleMemRef(cpuPtr *CPU, iPtr *DecodedInstr) bool {
	var (
		addr dg_phys_addr
		wd   dg_word
		dwd  dg_dword
		i32  int32
	)

	switch iPtr.mnemonic {

	case "LNLDA":
		addr = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		cpuPtr.ac[iPtr.acd] = sexWordToDWord(memReadWord(addr))

	case "LNSTA":
		addr = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		memWriteWord(addr, wd)

	case "LWLDA":
		addr = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		cpuPtr.ac[iPtr.acd] = memReadDWord(addr)

	case "LWSTA":
		addr = resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		dwd = cpuPtr.ac[iPtr.acd]
		memWriteDWord(addr, dwd)

	case "WBLM":
		/* AC0 - unused, AC1 - no. wds to move (if neg then descending order), AC2 - src, AC3 - dest */
		numWds := int32(cpuPtr.ac[1])
		var order int32 = 1
		if numWds < 0 {
			order = -1
		}
		if numWds == 0 {
			log.Println("INFO: WBLM called with AC1 == 0, not moving anything")
		} else {
			src := dg_phys_addr(cpuPtr.ac[2])
			dest := dg_phys_addr(cpuPtr.ac[3])
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "DEBUG: WBLM moving %d words from %d to %d\n", numWds, src, dest)
			}
			for numWds != 0 {
				memWriteWord(dest, memReadWord(src))
				numWds -= order
				if order == 1 {
					src++
					dest++
				} else {
					src--
					dest--
				}
			}
			cpuPtr.ac[1] = 0
			//cpuPtr.ac[2] = dg_dword(dest) // TODO confirm this
			//cpuPtr.ac[3] = dg_dword(dest)
			// TESTING..
			cpuPtr.ac[2] = dg_dword(src + 1) // TODO confirm this
			cpuPtr.ac[3] = dg_dword(dest + 1)
		}
	case "XLDB":
		cpuPtr.ac[iPtr.acd] = dg_dword(memReadByte(resolve16bitEagleAddr(cpuPtr, ' ', iPtr.mode, iPtr.disp), iPtr.loHiBit)) & 0x00ff

	case "XLEF":
		cpuPtr.ac[iPtr.acd] = dg_dword(resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp))

	case "XLEFB":
		loBit := iPtr.disp & 1
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp/2)
		addr <<= 1
		if loBit == 1 {
			addr++
		}
		cpuPtr.ac[iPtr.acd] = dg_dword(addr)

	case "XNLDA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		wd = memReadWord(addr)
		cpuPtr.ac[iPtr.acd] = sexWordToDWord(wd) // FIXME check this...

	case "XNSTA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		memWriteWord(addr, wd)

	case "XWADI":
		// add 1-4 to signed 32-bit acc
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		i32 = int32(memReadDWord(addr)) + iPtr.immVal
		// FIXME handle Carry and OVeRflow
		memWriteDWord(addr, dg_dword(i32))

	case "XWLDA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		dwd = memReadDWord(addr)
		cpuPtr.ac[iPtr.acd] = dwd

	case "XWSTA":
		addr = resolve16bitEagleAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		dwd = cpuPtr.ac[iPtr.acd]
		memWriteDWord(addr, dwd)

	default:
		log.Fatalf("ERROR: EAGLE_MEMREF instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
