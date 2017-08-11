// eagleIO.go
package main

import (
	"log"
	"mvemg/logging"
)

func eagleIO(cpuPtr *CPU, iPtr *decodedInstrT) bool {

	var (
		cmd, word, dataWord DgWordT
		dwd                 DgDwordT
		mapRegAddr          int
		rw                  bool
		wAddr               DgPhysAddrT
	)

	switch iPtr.mnemonic {

	case "CIO":
		// TODO handle I/O channel
		word = dwordGetLowerWord(cpuPtr.ac[iPtr.acs])
		mapRegAddr = int(word & 0x0fff)
		rw = testWbit(word, 0)
		if rw { // write command
			dataWord = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
			bmcdchWriteReg(mapRegAddr, dataWord)
		} else { // read command
			dataWord = bmcdchReadReg(mapRegAddr)
			cpuPtr.ac[iPtr.acd] = DgDwordT(dataWord)
		}

	case "CIOI":
		// TODO handle I/O channel
		if iPtr.acs == iPtr.acd {
			cmd = iPtr.immWord
		} else {
			cmd = iPtr.immWord | dwordGetLowerWord(cpuPtr.ac[iPtr.acs])
		}
		mapRegAddr = int(cmd & 0x0fff)
		rw = testWbit(cmd, 0)
		if rw { // write command
			dataWord = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
			bmcdchWriteReg(mapRegAddr, dataWord)
		} else { // read command
			dataWord = bmcdchReadReg(mapRegAddr)
			cpuPtr.ac[iPtr.acd] = DgDwordT(dataWord)
		}

	case "INTDS":
		cpu.ion = false

	case "INTEN":
		log.Fatal("ERROR: INTEN not yet supported")

	case "LCPID":
		dwd = CPU_MODEL_NO << 16
		dwd |= UCODE_REV << 8
		dwd |= MEM_SIZE_LCPID
		cpuPtr.ac[iPtr.acd] = dwd

	case "NCLID":
		cpuPtr.ac[0] = CPU_MODEL_NO
		cpuPtr.ac[1] = UCODE_REV
		cpuPtr.ac[2] = MEM_SIZE_LCPID // TODO Check this

	case "WLMP":
		if cpuPtr.ac[1] == 0 {
			mapRegAddr = int(cpuPtr.ac[0] & 0x7ff)
			wAddr = DgPhysAddrT(cpuPtr.ac[2])
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "WLMP called with AC1 = 0 - MapRegAddr was %d, 1st DWord was %d\n",
					mapRegAddr, memReadDWord(wAddr))
			}
			bmcdchWriteSlot(mapRegAddr, memReadDWord(wAddr))
			cpuPtr.ac[0]++
			cpuPtr.ac[2] += 2
		} else {
			for {
				bmcdchWriteSlot(int(cpuPtr.ac[0]&0x07ff), memReadDWord(DgPhysAddrT(cpuPtr.ac[2])))
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

	cpuPtr.pc += DgPhysAddrT(iPtr.instrLength)
	return true
}
