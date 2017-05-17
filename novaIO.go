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
		bus.busResetAllIODevices()
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
	case "IORST":
		bus.busResetAllIODevices()
		cpuPtr.ion = false
		// TODO More to do for SMP support - HaHa!

	case "PRTSEL":
		log.Printf("INFO: PRTSEL AC0: %d, PC: %d\n", cpuPtr.ac[0], cpuPtr.pc)
		// only handle the query mode, setting is a no-op on this 'single-channel' machine
		if dwordGetLowerWord(cpuPtr.ac[0]) == 0xffff {
			// return default I/O channel if -1 passed in
			cpuPtr.ac[0] = 0
		}

	default:
		log.Printf("ERROR: NOVA_IO instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc++

	return true
}
