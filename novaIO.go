// novaIO.go
package main

import (
	"log"
)

func novaIO(cpuPtr *Cpu, iPtr *DecodedInstr) bool {

	// a couple of special cases we need to catch
	// First: DICC 0,077 => I/O Reset
	if iPtr.mnemonic == "DIC" && iPtr.f == 'C' && iPtr.acd == 0 && iPtr.ioDev == DEV_CPU {
		log.Printf("INFO: I/O Reset due to DICC 0,CPU instruction\n")
		busResetAllIODevices()
		cpuPtr.pc++
		return true
	}
	// Second: DOC 0-3,077 => Halt
	if iPtr.mnemonic == "DOC" && iPtr.ioDev == DEV_CPU {
		log.Printf("INFO: CPU Halting due to DOC %d,CPU instruction\n", iPtr.acs)
		// do not advance PC
		return false
	}

	switch iPtr.mnemonic {

	case "DIA", "DIB", "DIC", "DOA", "DOB", "DOC":
		if busIsAttached(iPtr.ioDev) && busIsIODevice(iPtr.ioDev) {
			var abc byte
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
				busDataIn(cpuPtr, iPtr, abc)
			case "DOA", "DOB", "DOC":
				busDataOut(cpuPtr, iPtr, abc)
			}
		} else {
			log.Printf("WARN: I/O attempted to unattached or non-I/O capable device 0#%o\n", iPtr.ioDev)
			if iPtr.ioDev != 2 {
				return false // TODO Exception for ?MMU?
			}
		}

	case "IORST":
		busResetAllIODevices()
		cpuPtr.ion = false
		// TODO More to do for SMP support - HaHa!

	case "PRTSEL":
		log.Printf("INFO: PRTSEL AC0: %d, PC: %d\n", cpuPtr.ac[0], cpuPtr.pc)
		// only handle the query mode, setting is a no-op on this 'single-channel' machine
		if dwordGetLowerWord(cpuPtr.ac[0]) == 0xffff {
			// return default I/O channel if -1 passed in
			cpuPtr.ac[0] = 0
		}

	case "SKP":
		busy := busGetBusy(iPtr.ioDev)
		done := busGetDone(iPtr.ioDev)
		switch iPtr.t {
		case "BN":
			if busy {
				cpuPtr.pc++
			}
		case "BZ":
			if !busy {
				cpuPtr.pc++
			}
		case "DN":
			if done {
				cpuPtr.pc++
			}
		case "DZ":
			if !done {
				cpuPtr.pc++
			}
		}

	default:
		log.Printf("ERROR: NOVA_IO instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc++
	return true
}
