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
