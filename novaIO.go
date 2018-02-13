// novaIO.go

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
	"mvemg/util"
)

func novaIO(cpuPtr *CPUT, iPtr *decodedInstrT) bool {

	var (
		abc        byte
		busy, done bool
		ioFlagsDev ioFlagsDevT
		ioTestDev  ioTestDevT
		novaDataIo novaDataIoT
		//oneAcc1Word oneAcc1WordT
	)

	// The Eclipse LEF instruction is handled funkily...
	if cpuPtr.atu && cpuPtr.sbr[util.GetSegment(cpuPtr.pc)].lef {
		iPtr.mnemonic = "LEF"
		log.Fatalf("ERROR: LEF not yet implemented, location %d\n", cpuPtr.pc)
	}

	switch iPtr.mnemonic {

	case "DIA", "DIB", "DIC", "DOA", "DOB", "DOC":
		novaDataIo = iPtr.variant.(novaDataIoT)

		// catch CPU I/O instructions
		if novaDataIo.ioDev == devCPU {
			switch iPtr.mnemonic {
			case "DIA": // READS
				logging.DebugPrint(logging.DebugLog, "INFO: Interpreting DIA n,CPU as READS n instruction\n")
				return reads(cpuPtr, novaDataIo.acd)
			case "DIB": // INTA
				log.Fatalf("ERROR: DIB n,CPU (INTA )not yet implemented, location %d\n", cpuPtr.pc)
			case "DIC": // IORST
				logging.DebugPrint(logging.DebugLog, "INFO: I/O Reset due to DIC 0,CPU instruction\n")
				return iorst(cpuPtr)
			case "DOB": // MKSO
				logging.DebugPrint(logging.DebugLog, "INFO: Handling DOB %d, CPU instruction as MSKO\n", novaDataIo.acd)
				novaDataIo = iPtr.variant.(novaDataIoT)
				return msko(cpuPtr, novaDataIo.acd)
			case "DOC": // HALT
				logging.DebugPrint(logging.DebugLog, "INFO: CPU Halting due to DOC %d,CPU (HALT) instruction\n", novaDataIo.acd)
				return halt()
			}
		}

		if busIsAttached(novaDataIo.ioDev) && busIsIODevice(novaDataIo.ioDev) {
			switch iPtr.mnemonic {
			case "DOA", "DIA":
				abc = 'A'
			case "DOB", "DIB":
				abc = 'B'
			case "DOC", "DIC":
				abc = 'C'
			}
			switch iPtr.mnemonic {
			case "DIA", "DIB", "DIC":
				busDataIn(cpuPtr, &novaDataIo, abc)
			case "DOA", "DOB", "DOC":
				busDataOut(cpuPtr, &novaDataIo, abc)
			}
		} else {
			logging.DebugPrint(logging.DebugLog, "WARN: I/O attempted to unattached or non-I/O capable device 0#%o\n", novaDataIo.ioDev)
			if novaDataIo.ioDev != 2 {
				logging.DebugLogsDump()
				log.Fatal("Illegal I/O device crash") // TODO Exception for ?MMU?
			}
		}

	case "HALT":
		logging.DebugPrint(logging.DebugLog, "INFO: CPU Halting due to HALT instruction\n")
		return halt()

	case "IORST":
		// oneAcc1Word = iPtr.variant.(oneAcc1WordT) // <== this is just an assertion really
		busResetAllIODevices()
		cpuPtr.ion = false
		// TODO More to do for SMP support - HaHa!

	case "NIO":
		ioFlagsDev = iPtr.variant.(ioFlagsDevT)

		if ioFlagsDev.ioDev == devCPU {
			switch ioFlagsDev.f {
			case 'C': // INTDS
				return intds(cpuPtr)
			case 'S': // INTEN
				return inten(cpuPtr)
			}

		}
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "Sending NIO to device #%d.\n", ioFlagsDev.ioDev)
		}
		var novaDataIo novaDataIoT
		novaDataIo.f = ioFlagsDev.f
		novaDataIo.ioDev = ioFlagsDev.ioDev
		busDataOut(cpuPtr, &novaDataIo, 'N') // DUMMY FLAG

	case "PRTSEL":
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "INFO: PRTSEL AC0: %d, PC: %d\n", cpuPtr.ac[0], cpuPtr.pc)
		}
		// only handle the query mode, setting is a no-op on this 'single-channel' machine
		if util.DWordGetLowerWord(cpuPtr.ac[0]) == 0xffff {
			// return default I/O channel if -1 passed in
			cpuPtr.ac[0] = 0
		}

	case "SKP":
		ioTestDev = iPtr.variant.(ioTestDevT)
		if ioTestDev.ioDev == devCPU {
			busy = cpuPtr.ion
			done = cpuPtr.pfflag
		} else {
			busy = busGetBusy(ioTestDev.ioDev)
			done = busGetDone(ioTestDev.ioDev)
		}
		done = busGetDone(ioTestDev.ioDev)
		switch ioTestDev.t {
		case "BN":
			if busy {
				cpuPtr.pc++
				// if debugLogging {
				// 	logging.DebugPrint(logging.DebugLog, "... skipping\n")
				// }
			}
		case "BZ":
			if !busy {
				cpuPtr.pc++
				// if debugLogging {
				// 	logging.DebugPrint(logging.DebugLog, "... skipping\n")
				// }
			}
		case "DN":
			if done {
				cpuPtr.pc++
				// if debugLogging {
				// 	logging.DebugPrint(logging.DebugLog, "... skipping\n")
				// }
			}
		case "DZ":
			if !done {
				cpuPtr.pc++
				// if debugLogging {
				// 	logging.DebugPrint(logging.DebugLog, "... skipping\n")
				// }
			}
		}

	default:
		log.Fatalf("ERROR: NOVA_IO instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc++
	return true
}

func halt() bool {
	// do not advance PC
	return false // stop processing
}

func intds(cpuPtr *CPUT) bool {
	cpuPtr.ion = false
	cpuPtr.pc++
	return true
}

func inten(cpuPtr *CPUT) bool {
	cpuPtr.ion = true
	cpuPtr.pc++
	return true
}

func iorst(cpuPtr *CPUT) bool {
	busResetAllIODevices()
	cpuPtr.pc++
	return true
}

func msko(cpuPtr *CPUT, destAc int) bool {
	cpuPtr.mask = util.DWordGetLowerWord(cpuPtr.ac[destAc])
	cpuPtr.pc++
	return true
}

func reads(cpuPtr *CPUT, destAc int) bool {
	// load the AC with the contents of the dummy CPU register 'SR'
	cpuPtr.ac[destAc] = dg.DwordT(cpuPtr.sr)
	cpuPtr.pc++
	return true
}
