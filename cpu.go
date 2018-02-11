// cpu.go

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
	"fmt"
	"log"
	"mvemg/dg"
	"mvemg/util"
	"runtime"
	"sync"
	"time"
)

// TODO sbrBits is currently an abstraction of the Segment Base Registers - may need to represent physically
// via a 32-bit DWord in the future
type sbrBits struct {
	v, len, lef, io bool
	physAddr        uint32 // 19 bits used
}

// CPUT holds the current state of a CPUT
type CPUT struct {
	cpuMu sync.RWMutex
	// representations of physical attributes
	pc                      dg.PhysAddrT // 32-bit PC
	ac                      [4]dg.DwordT // 4 x 32-bit Accumulators
	mask                    dg.WordT     // interrupt mask
	psr                     dg.WordT     // Processor Status Register - see PoP A-4
	carry, atu, ion, pfflag bool         // flag bits
	sbr                     [8]sbrBits   // SBRs (see above)
	fpac                    [4]dg.QwordT // 4 x 64-bit Floating Point Accumulators
	fpsr                    dg.QwordT    // 64-bit Floating-Point Status Register
	sr                      dg.WordT     // Not sure about this... fake Switch Register

	// emulator internals
	instrCount uint64 // how many instructions executed during the current run, running at 2 MIPS this will loop round roughly every 100 million years!
	scpIO      bool   // true if console I/O is directed to the SCP
}

// cpuStatT defines the data we will send to the statusCollector monitor
type cpuStatT struct {
	pc                      dg.PhysAddrT
	ac                      [4]dg.DwordT
	carry, atu, ion, pfflag bool
	instrCount              uint64
	goVersion               string
	build                   string
	goroutineCount          int
	hostCPUCount            int
	heapSizeMB              int
}

const cpuStatPeriodMs = 500 // 125 // i.e. we send stats every 1/8th of a second

var cpu CPUT

// shamelessly borrowed from SIMH...
var eclipseAplRom = [...]dg.WordT{
	062677,  //      IORST           ;Reset all I/O
	060477,  //      READS 0         ;Read SR into AC0
	024026,  //      LDA 1,C77       ;Get dev mask
	0107400, //      AND 0,1         ;Isolate dev code
	0124000, //      COM 1,1         ;- device code - 1
	010014,  // LOOP: ISZ OP1        ;Device code to all
	010030,  //      ISZ OP2         ;I/O instructions
	010032,  //      ISZ OP3
	0125404, //      INC 1,1,SZR     ;done?
	000005,  //      JMP LOOP        ;No, increment again
	030016,  //      LDA 2,C377      ;place JMP 377 into
	050377,  //      STA 2,377       ;location 377
	060077,  // OP1: 060077          ;start device (NIOS 0)
	0101102, //      MOVL 0,0,SZC    ;Test switch 0, low speed?
	000377,  // C377: JMP 377        ;no - jmp 377 & wait
	004030,  // LOOP2: JSR GET+1     ;Get a frame
	0101065, //      MOVC 0,0,SNR    ;is it non-zero?
	000017,  //      JMP LOOP2       ;no, ignore
	004027,  // LOOP4: JSR GET       ;yes, get full word
	046026,  //      STA 1,@C77      ;store starting at 100
	//                      ;2's complement of word ct
	010100, //      ISZ 100         ;done?
	000022, //      JMP LOOP4       ;no, get another
	000077, // C77: JMP 77          ;yes location ctr and
	//                      ;jmp to last word
	0126420, // GET: SUBZ 1,1        ; clr AC1, set carry
	// OP2:
	063577,  // LOOP3: 063577        ;done? (SKPDN 0) - 1
	000030,  //      JMP LOOP3       ;no -- wait
	060477,  // OP3: 060477          ;y--read in ac0 (DIAS 0,0)
	0107363, //      ADDCS 0,1,SNC   ;add 2 frames swapped - got 2nd?
	000030,  //      JMP LOOP3       ;no go back after it
	0125300, //      MOVS 1,1        ;yes swap them
	001400,  //      JMP 0,3         ;rtn with full word
	0,       //      0               ;padding
}

func cpuInit(statsChan chan cpuStatT) *CPUT {
	busAddDevice(devCPU, "CPU", cpuPMB, true, false, false)
	go cpuStatSender(statsChan)
	return &cpu
}

func cpuPrintableStatus() string {
	cpu.cpuMu.RLock()
	res := fmt.Sprintf("%c      AC0       AC1       AC2       AC3        PC CRY ATU%c", asciiNL, asciiNL)
	res += fmt.Sprintf("%9d %9d %9d %9d %9d", cpu.ac[0], cpu.ac[1], cpu.ac[2], cpu.ac[3], cpu.pc)
	res += fmt.Sprintf("  %d   %d", util.BoolToInt(cpu.carry), util.BoolToInt(cpu.atu))
	cpu.cpuMu.RUnlock()
	return res
}

func cpuCompactPrintableStatus() string {
	cpu.cpuMu.RLock()
	res := fmt.Sprintf("AC0: %d AC1: %d AC2: %d AC3: %d CRY: %d PC: %d",
		cpu.ac[0], cpu.ac[1], cpu.ac[2], cpu.ac[3], util.BoolToInt(cpu.carry), cpu.pc)
	cpu.cpuMu.RUnlock()
	return res
}

// GetOVR is a getter for the OVR flag embedded in the PSR
func (c *CPUT) GetOVR() bool {
	return util.TestWbit(c.psr, 1)
}

// SetOVR is a setter for the OVR flag embedded in the PSR
func (c *CPUT) SetOVR(newOVR bool) {
	if newOVR {
		util.SetWbit(&c.psr, 1)
	} else {
		util.ClearWbit(&c.psr, 1)
	}
}

// GetOVK is a getter for the OVK flag embedded in the PSR
func (c *CPUT) GetOVK() bool {
	return util.TestWbit(c.psr, 0)
}

// SetOVK is a setter for the OVK flag embedded in the PSR
func (c *CPUT) SetOVK(newOVK bool) {
	if newOVK {
		util.SetWbit(&c.psr, 0)
	} else {
		util.ClearWbit(&c.psr, 0)
	}
}

// Execute a single instruction
// A false return means failure, the VM should stop
func cpuExecute(iPtr *decodedInstrT) (rc bool) {
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
		log.Println("ERROR: Unimplemented instruction type in cpuExecute()")
		rc = false
	}
	cpu.instrCount++
	cpu.cpuMu.Unlock()
	return rc
}

func cpuStatSender(sChan chan cpuStatT) {
	var stats cpuStatT
	var memStats runtime.MemStats
	stats.goVersion = runtime.Version()
	stats.hostCPUCount = runtime.NumCPU()
	for {
		cpu.cpuMu.RLock()
		stats.pc = cpu.pc
		stats.ac[0] = cpu.ac[0]
		stats.ac[1] = cpu.ac[1]
		stats.ac[2] = cpu.ac[1]
		stats.ac[3] = cpu.ac[3]
		stats.instrCount = cpu.instrCount
		cpu.cpuMu.RUnlock()
		stats.goroutineCount = runtime.NumGoroutine()
		runtime.ReadMemStats(&memStats)
		stats.heapSizeMB = int(memStats.HeapAlloc / 1048576)
		select {
		case sChan <- stats:
		default:
		}
		time.Sleep(time.Millisecond * cpuStatPeriodMs)
	}
}
