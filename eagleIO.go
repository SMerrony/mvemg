// eagleIO.go
package main

import (
	"log"
)

func eagleIO(cpuPtr *Cpu, iPtr *DecodedInstr) bool {

	var (
		dwd dg_dword
	)

	switch iPtr.mnemonic {

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

	default:
		log.Printf("ERROR: EAGLE_IO instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
