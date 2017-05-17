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

func (c *Cpu) cpuInit() {
	bus.busAddDevice(DEV_CPU, "CPU", CPU_PMB, true, false, false)
}

func (c *Cpu) cpuPrintableStatus() string {
	res := fmt.Sprintf("%c      AC0       AC1       AC2       AC3        PC CRY ATU%c", ASCII_NL, ASCII_NL)
	res += fmt.Sprintf("%9d %9d %9d %9d %9d", c.ac[0], c.ac[1], c.ac[2], c.ac[3], c.pc)
	res += fmt.Sprintf("  %d   %d", boolToInt(c.carry), boolToInt(c.atu))
	return res
}

func (c *Cpu) cpuCompactPrintableStatus() string {
	res := fmt.Sprintf("AC0: %d AC1: %d AC2: %d AC3: %d CRY: %d PC: %d",
		c.ac[0], c.ac[1], c.ac[2], c.ac[3], boolToInt(c.carry), c.pc)
	return res
}

// Execute a single instruction
// A false return means failure, the VM should stop
func (cpuPtr *Cpu) cpuExecute(iPtr *DecodedInstr) bool {
	rc := false
	switch iPtr.instrType {
	case NOVA_MEMREF:
		rc = novaMemRef(cpuPtr, iPtr)
	case NOVA_OP:
		rc = novaOp(cpuPtr, iPtr)
	case NOVA_IO:
		rc = novaIO(cpuPtr, iPtr)
		//	case NOVA_PC:
		//		rc = novaPC(cpuPtr, iPtr)
		//	case ECLIPSE_MEMREF:
		//		rc = eclipseMemRef(cpuPtr, iPtr)
		//	case ECLIPSE_OP:
		//		rc = eclipseOp(cpuPtr, iPtr)
		//	case ECLIPSE_PC:
		//		rc = eclipsePC(cpuPtr, iPtr)
		//	case ECLIPSE_STACK:
		//		rc = eclipseStack(cpuPtr, iPtr)
	case EAGLE_OP:
		rc = eagleOp(cpuPtr, iPtr)
	case EAGLE_MEMREF:
		rc = eagleMemRef(cpuPtr, iPtr)
	case EAGLE_PC:
		rc = eaglePC(cpuPtr, iPtr)
		//	case EAGLE_IO:
		//		rc = eagleIO(cpuPtr, iPtr)
		//	case EAGLE_STACK:
		//		rc = eagleStack(cpuPtr, iPtr)
	default:
		log.Fatalln("ERROR: Unimplemented instruction type in cpuExecute()")
	}
	cpuPtr.instrCount++
	return rc
}
