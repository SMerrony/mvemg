// bus.go

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
	"mvemg/logging"
	"sync"

	"github.com/SMerrony/dgemug/util"
)

type (
	// ResetFunc stores an I/O reset func pointer
	ResetFunc func()

	// DataOutFunc stores a DOx func pointer
	DataOutFunc func(*CPUT, *novaDataIoT, byte)

	// DataInFunc stores a DIx func pointer
	DataInFunc func(*CPUT, *novaDataIoT, byte)
)

// type pioMsgT struct {
// 	cpuPtr *CPUT
// 	iPtr   *decodedInstrT
// 	IO     byte // 'I' or 'O'
// 	abc    byte // 'A', 'B', or 'C'
// }

type device struct {
	devMu           sync.RWMutex
	mnemonic        string
	priorityMaskBit int
	resetFunc       ResetFunc
	dataOutFunc     DataOutFunc
	dataInFunc      DataInFunc
	simAttached     bool
	simImageName    string
	ioDevice        bool
	bootable        bool
	busy            bool
	done            bool
}

type devices [devMax]device

var d devices

func busInit() {
	for dev := range d {
		d[dev].devMu.Lock()
		d[dev].mnemonic = ""
		d[dev].priorityMaskBit = 0
		d[dev].dataInFunc = nil
		d[dev].dataOutFunc = nil
		d[dev].simAttached = false
		d[dev].ioDevice = false
		d[dev].bootable = false
		d[dev].busy = false
		d[dev].done = false
		d[dev].devMu.Unlock()
	}
}

func busAddDevice(devNum int, mnem string, pmb int, att bool, io bool, boot bool) {
	if devNum >= devMax {
		log.Fatalf("ERROR: Attempt to add device with too-high device number: 0%o", devNum)
	}
	d[devNum].devMu.Lock()
	d[devNum].mnemonic = mnem
	d[devNum].priorityMaskBit = pmb
	d[devNum].simAttached = att
	d[devNum].ioDevice = io
	d[devNum].bootable = boot
	logging.DebugPrint(logging.DebugLog, "INFO: Device 0%o added to bus\n", devNum)
	d[devNum].devMu.Unlock()
}

func busSetDataInFunc(devNum int, fn DataInFunc) {
	d[devNum].devMu.Lock()
	d[devNum].dataInFunc = fn
	logging.DebugPrint(logging.DebugLog, "INFO: Bus Data In function set for dev #0%o (%d.)\n",
		devNum, devNum)
	d[devNum].devMu.Unlock()
}

func busDataIn(cpuPtr *CPUT, iPtr *novaDataIoT, abc byte) {
	//var pio pioMsgT
	//logging.DebugPrint(logging.DEBUG_LOG, "DEBUG: Bus Data In function called for dev #0%o\n", iPtr.ioDev)
	d[iPtr.ioDev].devMu.RLock()
	if d[iPtr.ioDev].dataInFunc == nil {
		log.Fatalf("ERROR: busDataIn called for device %d with no function set", iPtr.ioDev)
	}
	d[iPtr.ioDev].devMu.RUnlock()
	if d[iPtr.ioDev].dataInFunc != nil {
		d[iPtr.ioDev].dataInFunc(cpuPtr, iPtr, abc)
	}
	// else {
	// 	pio.cpuPtr = cpuPtr
	// 	pio.iPtr = iPtr
	// 	pio.IO = 'I'
	// 	pio.abc = abc
	// 	d[iPtr.ioDev].pioChan <- pio
	// 	_ = <-d[iPtr.ioDev].pioDoneChan
	// }

	// logging.DebugPrint(logging.DEBUG_LOG, "INFO: Bus Data In function called for dev #0%o\n", iPtr.ioDev)
}

func busSetDataOutFunc(devNum int, fn DataOutFunc) {
	d[devNum].devMu.Lock()
	d[devNum].dataOutFunc = fn
	d[devNum].devMu.Unlock()
	logging.DebugPrint(logging.DebugLog, "INFO: Bus Data Out function set for dev #0%o (%d.)\n",
		devNum, devNum)
}

func busDataOut(cpuPtr *CPUT, iPtr *novaDataIoT, abc byte) {
	//var pio pioMsgT
	d[iPtr.ioDev].devMu.Lock()
	if d[iPtr.ioDev].dataOutFunc == nil {
		logging.DebugLogsDump()
		log.Fatalf("ERROR: busDataOut called for device %d with no function set",
			iPtr.ioDev)
	}
	d[iPtr.ioDev].devMu.Unlock()
	if d[iPtr.ioDev].dataOutFunc != nil {
		d[iPtr.ioDev].dataOutFunc(cpuPtr, iPtr, abc)
		//logging.DebugPrint(logging.DebugLog, "INFO: Bus Data Out function called for dev #0%o\n", iPtr.ioDev)
	}
	// else {
	// 	pio.cpuPtr = cpuPtr
	// 	pio.iPtr = iPtr
	// 	pio.IO = 'O'
	// 	pio.abc = abc
	// 	d[iPtr.ioDev].pioChan <- pio
	// 	_ = <-d[iPtr.ioDev].pioDoneChan
	// 	//logging.DebugPrint(logging.DebugLog, "INFO: Bus Data Out sent PIO msg to dev #0%o\n", iPtr.ioDev)
	// }
}

func busSetResetFunc(devNum int, resetFn ResetFunc) {
	d[devNum].devMu.Lock()
	d[devNum].resetFunc = resetFn
	logging.DebugPrint(logging.DebugLog, "INFO: Bus reset function set for dev #0%o\n", devNum)
	d[devNum].devMu.Unlock()
}

func busResetDevice(devNum int) {
	d[devNum].devMu.Lock()
	io := d[devNum].ioDevice
	d[devNum].devMu.Unlock()
	if io {
		d[devNum].resetFunc()
	} else {
		log.Fatalf("ERROR: Attempt to reset non-I/O device #0%o\n", devNum)
	}

}

func busResetAllIODevices() {
	for dev := range d {
		d[dev].devMu.Lock()
		io := d[dev].ioDevice
		d[dev].devMu.Unlock()
		if io {
			busResetDevice(dev)
		}
	}
}

func busSetAttached(devNum int, imgName string) {
	d[devNum].devMu.Lock()
	d[devNum].simAttached = true
	d[devNum].simImageName = imgName
	d[devNum].devMu.Unlock()
}
func busSetDetached(devNum int) {
	d[devNum].devMu.Lock()
	d[devNum].simAttached = false
	d[devNum].simImageName = ""
	d[devNum].devMu.Unlock()
}
func busIsAttached(devNum int) bool {
	d[devNum].devMu.RLock()
	att := d[devNum].simAttached
	d[devNum].devMu.RUnlock()
	return att
}

func busSetBusy(devNum int, f bool) {
	d[devNum].devMu.Lock()
	d[devNum].busy = f
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... Busy flag set to %d for device #0%o\n", util.BoolToInt(f), devNum)
	}
	d[devNum].devMu.Unlock()
}

func busSetDone(devNum int, f bool) {
	d[devNum].devMu.Lock()
	d[devNum].done = f
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... Done flag set to %d for device #0%o\n", util.BoolToInt(f), devNum)
	}
	d[devNum].devMu.Unlock()
}

func busGetBusy(devNum int) bool {
	d[devNum].devMu.RLock()
	bz := d[devNum].busy
	d[devNum].devMu.RUnlock()
	return bz
}

func busGetDone(devNum int) bool {
	d[devNum].devMu.RLock()
	dn := d[devNum].done
	d[devNum].devMu.RUnlock()
	return dn
}

func busIsBootable(devNum int) bool {
	d[devNum].devMu.RLock()
	bt := d[devNum].bootable
	d[devNum].devMu.RUnlock()
	return bt
}

func busIsIODevice(devNum int) bool {
	d[devNum].devMu.RLock()
	io := d[devNum].ioDevice
	d[devNum].devMu.RUnlock()
	return io
}

func busGetPrintableDevList() string {
	lst := fmt.Sprintf(" #  Mnem   PMB  I/O Busy Done Status%c", asciiNL)
	var line string
	for dev := range d {
		d[dev].devMu.RLock()
		if d[dev].mnemonic != "" {
			line = fmt.Sprintf("%#3o %-6s %2d. %3d %4d %4d  ",
				dev, d[dev].mnemonic, d[dev].priorityMaskBit,
				util.BoolToInt(d[dev].ioDevice), util.BoolToInt(d[dev].busy), util.BoolToInt(d[dev].done))
			if d[dev].simAttached {
				line += "Attached"
				if d[dev].simImageName != "" {
					line += " to image: " + d[dev].simImageName
				}
			} else {
				line += "Not Attached"
			}
			// Commented out the below for now as it's a bit misleading...
			// if d[dev].bootable {
			// 	line += ", Bootable"
			// }
			line += "\012"
			lst += line
		}
		d[dev].devMu.RUnlock()
	}
	return lst
}
