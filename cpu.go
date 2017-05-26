// cpu.go
package main

import (
	"fmt"
	"log"
)

// TODO sbrBits is currently an abstraction of the Segment Base Registers - may need to represent physically
// via a 32-bit DWord in the future
type sbrBits struct {
	v, len, lef, io bool
	physAddr        uint32 // 19 bits used
}

type Cpu struct {
	// representations of physical attributes
	pc                      dg_phys_addr
	ac                      [4]dg_dword
	carry, atu, ion, pfflag bool
	sbr                     [8]sbrBits

	// emulator internals
	instrCount uint64
	consoleEsc bool
}

var (
	cpu Cpu
)

func cpuInit() {
	busAddDevice(DEV_CPU, "CPU", CPU_PMB, true, false, false)
}

func cpuPrintableStatus() string {
	res := fmt.Sprintf("%c      AC0       AC1       AC2       AC3        PC CRY ATU%c", ASCII_NL, ASCII_NL)
	res += fmt.Sprintf("%9d %9d %9d %9d %9d", cpu.ac[0], cpu.ac[1], cpu.ac[2], cpu.ac[3], cpu.pc)
	res += fmt.Sprintf("  %d   %d", boolToInt(cpu.carry), boolToInt(cpu.atu))
	return res
}

func cpuCompactPrintableStatus() string {
	res := fmt.Sprintf("AC0: %d AC1: %d AC2: %d AC3: %d CRY: %d PC: %d",
		cpu.ac[0], cpu.ac[1], cpu.ac[2], cpu.ac[3], boolToInt(cpu.carry), cpu.pc)
	return res
}

// Execute a single instruction
// A false return means failure, the VM should stop
func cpuExecute(iPtr *DecodedInstr) bool {
	rc := false
	switch iPtr.instrType {
	case NOVA_MEMREF:
		rc = novaMemRef(&cpu, iPtr)
	case NOVA_OP:
		rc = novaOp(&cpu, iPtr)
	case NOVA_IO:
		rc = novaIO(&cpu, iPtr)
	case NOVA_PC:
		rc = novaPC(&cpu, iPtr)
		//	case ECLIPSE_MEMREF:
		//		rc = eclipseMemRef(&cpu, iPtr)
	case ECLIPSE_OP:
		rc = eclipseOp(&cpu, iPtr)
		//	case ECLIPSE_PC:
		//		rc = eclipsePC(&cpu, iPtr)
		//	case ECLIPSE_STACK:
		//		rc = eclipseStack(&cpu, iPtr)
	case EAGLE_IO:
		rc = eagleIO(&cpu, iPtr)
	case EAGLE_OP:
		rc = eagleOp(&cpu, iPtr)
	case EAGLE_MEMREF:
		rc = eagleMemRef(&cpu, iPtr)
	case EAGLE_PC:
		rc = eaglePC(&cpu, iPtr)
		//	case EAGLE_STACK:
		//		rc = eagleStack(&cpu, iPtr)
	default:
		log.Fatalln("ERROR: Unimplemented instruction type in cpuExecute()")
	}
	cpu.instrCount++
	return rc
}
