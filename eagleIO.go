// eagleIO.go

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
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
	"mvemg/util"
)

func eagleIO(cpuPtr *CPUT, iPtr *decodedInstrT) bool {

	var (
		cmd, word, dataWord dg.WordT
		dwd                 dg.DwordT
		ok                  bool
		mapRegAddr          int
		rw                  bool
		wAddr               dg.PhysAddrT
		oneAcc1Word         oneAcc1WordT
		twoAcc1Word         twoAcc1WordT
		twoAccImm2Word      twoAccImm2WordT
	)

	switch iPtr.mnemonic {

	case "CIO":
		// TODO handle I/O channel
		twoAcc1Word = iPtr.variant.(twoAcc1WordT)
		word = util.DWordGetLowerWord(cpuPtr.ac[twoAcc1Word.acs])
		mapRegAddr = int(word & 0x0fff)
		rw = util.TestWbit(word, 0)
		if rw { // write command
			dataWord = util.DWordGetLowerWord(cpuPtr.ac[twoAcc1Word.acd])
			memory.BmcdchWriteReg(mapRegAddr, dataWord)
		} else { // read command
			dataWord = memory.BmcdchReadReg(mapRegAddr)
			cpuPtr.ac[twoAcc1Word.acd] = dg.DwordT(dataWord)
		}

	case "CIOI":
		// TODO handle I/O channel
		twoAccImm2Word = iPtr.variant.(twoAccImm2WordT)
		if twoAccImm2Word.acs == twoAccImm2Word.acd {
			cmd = twoAccImm2Word.immWord
		} else {
			cmd = twoAccImm2Word.immWord | util.DWordGetLowerWord(cpuPtr.ac[twoAccImm2Word.acs])
		}
		mapRegAddr = int(cmd & 0x0fff)
		rw = util.TestWbit(cmd, 0)
		if rw { // write command
			dataWord = util.DWordGetLowerWord(cpuPtr.ac[twoAccImm2Word.acd])
			memory.BmcdchWriteReg(mapRegAddr, dataWord)
		} else { // read command
			dataWord = memory.BmcdchReadReg(mapRegAddr)
			cpuPtr.ac[twoAccImm2Word.acd] = dg.DwordT(dataWord)
		}

	case "ECLID": // seems to be the same as LCPID
		dwd = cpuModelNo << 16
		dwd |= ucodeRev << 8
		dwd |= memory.MemSizeLCPID
		cpuPtr.ac[0] = dwd

	case "INTDS":
		return intds(cpuPtr)

	case "INTEN":
		return inten(cpuPtr)

	case "LCPID": // seems to be the same as ECLID
		dwd = cpuModelNo << 16
		dwd |= ucodeRev << 8
		dwd |= memory.MemSizeLCPID
		cpuPtr.ac[0] = dwd

		// MSKO is handled via DOB n,CPU

	case "NCLID":
		cpuPtr.ac[0] = cpuModelNo
		cpuPtr.ac[1] = ucodeRev
		cpuPtr.ac[2] = memory.MemSizeLCPID // TODO Check this

	case "READS":
		oneAcc1Word = iPtr.variant.(oneAcc1WordT)
		return reads(cpuPtr, oneAcc1Word.acd)

	case "WLMP":
		if cpuPtr.ac[1] == 0 {
			mapRegAddr = int(cpuPtr.ac[0] & 0x7ff)
			wAddr = dg.PhysAddrT(cpuPtr.ac[2])
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "WLMP called with AC1 = 0 - MapRegAddr was %d, 1st DWord was %d\n",
					mapRegAddr, memory.ReadDWord(wAddr))
			}
			// memory.BmcdchWriteSlot(mapRegAddr, memory.ReadDWord(wAddr))
			// cpuPtr.ac[0]++
			// cpuPtr.ac[2] += 2
		} else {
			for {
				dwd, ok = memory.ReadDWordTrap(dg.PhysAddrT(cpuPtr.ac[2]))
				if !ok {
					log.Fatalf("ERROR: Memory access failed at PC: %d\n", cpuPtr.pc)
				}
				memory.BmcdchWriteSlot(int(cpuPtr.ac[0]&0x07ff), dwd)
				if debugLogging {
					logging.DebugPrint(logging.DebugLog, "WLMP writing slot %d\n", 1+(cpuPtr.ac[0]&0x7ff))
				}
				cpuPtr.ac[2] += 2
				cpuPtr.ac[0]++
				cpuPtr.ac[1]--
				if cpuPtr.ac[1] == 0 {
					break
				}
			}
		}

	default:
		log.Fatalf("ERROR: EAGLE_IO instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg.PhysAddrT(iPtr.instrLength)
	return true
}
