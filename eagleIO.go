// eagleIO.go
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
		mapRegAddr          int
		rw                  bool
		wAddr               dg.PhysAddrT
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

	case "INTDS":
		cpu.ion = false

	case "INTEN":
		log.Fatal("ERROR: INTEN not yet supported")

	case "LCPID":
		dwd = CPU_MODEL_NO << 16
		dwd |= UCODE_REV << 8
		dwd |= memory.MemSizeLCPID
		cpuPtr.ac[0] = dwd

	case "NCLID":
		cpuPtr.ac[0] = CPU_MODEL_NO
		cpuPtr.ac[1] = UCODE_REV
		cpuPtr.ac[2] = memory.MemSizeLCPID // TODO Check this

	case "WLMP":
		if cpuPtr.ac[1] == 0 {
			mapRegAddr = int(cpuPtr.ac[0] & 0x7ff)
			wAddr = dg.PhysAddrT(cpuPtr.ac[2])
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "WLMP called with AC1 = 0 - MapRegAddr was %d, 1st DWord was %d\n",
					mapRegAddr, memory.ReadDWord(wAddr))
			}
			memory.BmcdchWriteSlot(mapRegAddr, memory.ReadDWord(wAddr))
			cpuPtr.ac[0]++
			cpuPtr.ac[2] += 2
		} else {
			for {
				memory.BmcdchWriteSlot(int(cpuPtr.ac[0]&0x07ff), memory.ReadDWord(dg.PhysAddrT(cpuPtr.ac[2])))
				if debugLogging {
					logging.DebugPrint(logging.DebugLog, "WLMP writing slot %d\n", 1+(cpuPtr.ac[0]&0x7ff))
				}
				cpuPtr.ac[2] += 2
				cpuPtr.ac[0]++
				cpuPtr.ac[1]--
				if cpuPtr.ac[1] <= 0 {
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
