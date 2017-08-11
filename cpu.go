// cpu.go
package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// TODO sbrBits is currently an abstraction of the Segment Base Registers - may need to represent physically
// via a 32-bit DWord in the future
type sbrBits struct {
	v, len, lef, io bool
	physAddr        uint32 // 19 bits used
}

// CPU holds the current state of a CPU
type CPU struct {
	cpuMu sync.RWMutex
	// representations of physical attributes
	pc                           DgPhysAddrT
	ac                           [4]DgDwordT
	carry, atu, ion, pfflag, ovk bool
	sbr                          [8]sbrBits

	// emulator internals
	instrCount uint64
	scpIO      bool
}

// cpuStatT defines the data we will send to the statusCollector monitor
type cpuStatT struct {
	pc                      DgPhysAddrT
	ac                      [4]DgDwordT
	carry, atu, ion, pfflag bool
	instrCount              uint64
}

const cpuStatPeriodMs = 500 // 125 // i.e. we send stats every 1/8th of a second

var (
	cpu CPU
)

func cpuInit(statsChan chan cpuStatT) *CPU {
	busAddDevice(DEV_CPU, "CPU", CPU_PMB, true, false, false)
	go cpuStatSender(statsChan)
	return &cpu
}

func cpuPrintableStatus() string {
	cpu.cpuMu.RLock()
	res := fmt.Sprintf("%c      AC0       AC1       AC2       AC3        PC CRY ATU%c", ASCII_NL, ASCII_NL)
	res += fmt.Sprintf("%9d %9d %9d %9d %9d", cpu.ac[0], cpu.ac[1], cpu.ac[2], cpu.ac[3], cpu.pc)
	res += fmt.Sprintf("  %d   %d", BoolToInt(cpu.carry), BoolToInt(cpu.atu))
	cpu.cpuMu.RUnlock()
	return res
}

func cpuCompactPrintableStatus() string {
	cpu.cpuMu.RLock()
	res := fmt.Sprintf("AC0: %d AC1: %d AC2: %d AC3: %d CRY: %d PC: %d",
		cpu.ac[0], cpu.ac[1], cpu.ac[2], cpu.ac[3], BoolToInt(cpu.carry), cpu.pc)
	cpu.cpuMu.RUnlock()
	return res
}

// Execute a single instruction
// A false return means failure, the VM should stop
func cpuExecute(iPtr *decodedInstrT) bool {
	rc := false
	cpu.cpuMu.Lock()
	switch iPtr.instrType {
	case NOVA_MEMREF:
		rc = novaMemRef(&cpu, iPtr)
	case NOVA_OP:
		rc = novaOp(&cpu, iPtr)
	case NOVA_IO:
		rc = novaIO(&cpu, iPtr)
	case NOVA_PC:
		rc = novaPC(&cpu, iPtr)
	case ECLIPSE_MEMREF:
		rc = eclipseMemRef(&cpu, iPtr)
	case ECLIPSE_OP:
		rc = eclipseOp(&cpu, iPtr)
	case ECLIPSE_PC:
		rc = eclipsePC(&cpu, iPtr)
	case ECLIPSE_STACK:
		rc = eclipseStack(&cpu, iPtr)
	case EAGLE_IO:
		rc = eagleIO(&cpu, iPtr)
	case EAGLE_OP:
		rc = eagleOp(&cpu, iPtr)
	case EAGLE_MEMREF:
		rc = eagleMemRef(&cpu, iPtr)
	case EAGLE_PC:
		rc = eaglePC(&cpu, iPtr)
	case EAGLE_STACK:
		rc = eagleStack(&cpu, iPtr)
	default:
		log.Fatalln("ERROR: Unimplemented instruction type in cpuExecute()")
	}
	cpu.instrCount++
	cpu.cpuMu.Unlock()
	return rc
}

func cpuStatSender(sChan chan cpuStatT) {
	var stats cpuStatT
	for {
		cpu.cpuMu.RLock()
		stats.pc = cpu.pc
		stats.ac[0] = cpu.ac[0]
		stats.ac[1] = cpu.ac[1]
		stats.ac[2] = cpu.ac[1]
		stats.ac[3] = cpu.ac[3]
		stats.instrCount = cpu.instrCount
		cpu.cpuMu.RUnlock()
		select {
		case sChan <- stats:
		default:
		}
		time.Sleep(time.Millisecond * cpuStatPeriodMs)
	}
}
