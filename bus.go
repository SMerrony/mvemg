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
)

type (
	// I/O reset func
	ResetFunc func()

	// DOx func
	DataOutFunc func(*CPU, *decodedInstrT, byte)

	// DIx func
	DataInFunc func(*CPU, *decodedInstrT, byte)
)

type device struct {
	mnemonic        string
	priorityMaskBit int
	resetFunc       ResetFunc
	dataOutFunc     DataOutFunc
	dataInFunc      DataInFunc
	simAttached     bool
	ioDevice        bool
	bootable        bool
	busy            bool
	done            bool
}

type devices [MAX_DEVICES]device

var d devices

func busInit() {
	for dev := range d {
		d[dev].mnemonic = ""
		d[dev].priorityMaskBit = 0
		d[dev].dataInFunc = nil
		d[dev].dataOutFunc = nil
		d[dev].simAttached = false
		d[dev].ioDevice = false
		d[dev].bootable = false
		d[dev].busy = false
		d[dev].done = false
	}
	bmcdchInit()
}

func busAddDevice(devNum int, mnem string, pmb int, att bool, io bool, boot bool) {
	if devNum >= MAX_DEVICES {
		log.Fatalf("ERROR: Attempt to add device with too-high device number: 0%o", devNum)
	}
	d[devNum].mnemonic = mnem
	d[devNum].priorityMaskBit = pmb
	d[devNum].simAttached = att
	d[devNum].ioDevice = io
	d[devNum].bootable = boot
	logging.DebugPrint(logging.DebugLog, "INFO: Device 0%o added to bus\n", devNum)
}

func busSetDataInFunc(devNum int, fn DataInFunc) {
	d[devNum].dataInFunc = fn
	logging.DebugPrint(logging.DebugLog, "INFO: Bus Data In function set for dev #0%o\n", devNum)
}

func busDataIn(cpuPtr *CPU, iPtr *decodedInstrT, abc byte) {
	//logging.DebugPrint(logging.DEBUG_LOG, "DEBUG: Bus Data In function called for dev #0%o\n", iPtr.ioDev)
	if d[iPtr.ioDev].dataInFunc == nil {
		log.Fatal("ERROR: busDataIn called with no function set")
	}
	d[iPtr.ioDev].dataInFunc(cpuPtr, iPtr, abc)
	// logging.DebugPrint(logging.DEBUG_LOG, "INFO: Bus Data In function called for dev #0%o\n", iPtr.ioDev)
}

func busSetDataOutFunc(devNum int, fn DataOutFunc) {
	d[devNum].dataOutFunc = fn
	logging.DebugPrint(logging.DebugLog, "INFO: Bus Data Out function set for dev #0%o\n", devNum)
}

func busDataOut(cpuPtr *CPU, iPtr *decodedInstrT, abc byte) {
	if d[iPtr.ioDev].dataOutFunc == nil {
		log.Fatal("ERROR: busDataOut called with no function set")
	}
	d[iPtr.ioDev].dataOutFunc(cpuPtr, iPtr, abc)
	logging.DebugPrint(logging.DebugLog, "INFO: Bus Data Out function called for dev #0%o\n", iPtr.ioDev)
}

func busSetResetFunc(devNum int, resetFn ResetFunc) {
	d[devNum].resetFunc = resetFn
	logging.DebugPrint(logging.DebugLog, "INFO: Bus reset function set for dev #0%o\n", devNum)
}

func busResetDevice(devNum int) {
	if d[devNum].ioDevice {
		d[devNum].resetFunc()
	} else {
		log.Fatalf("ERROR: Attempt to reset non-I/O device #0%o\n", devNum)
	}
}

func busResetAllIODevices() {
	for dev := range d {
		if d[dev].ioDevice {
			busResetDevice(dev)
		}
	}
}

func busSetAttached(devNum int) {
	d[devNum].simAttached = true
}
func busSetDetached(devNum int) {
	d[devNum].simAttached = false
}
func busIsAttached(devNum int) bool {
	return d[devNum].simAttached
}

func busSetBusy(devNum int, f bool) {
	d[devNum].busy = f
	logging.DebugPrint(logging.DebugLog, "... Busy flag set to %d for device #0%o\n", BoolToInt(f), devNum)
}

func busSetDone(devNum int, f bool) {
	d[devNum].done = f
	logging.DebugPrint(logging.DebugLog, "... Done flag set to %d for device #0%o\n", BoolToInt(f), devNum)
}

func busGetBusy(devNum int) bool {
	return d[devNum].busy
}

func busGetDone(devNum int) bool {
	return d[devNum].done
}

func busIsBootable(devNum int) bool {
	return d[devNum].bootable
}

func busIsIODevice(devNum int) bool {
	return d[devNum].ioDevice
}

func busGetPrintableDevList() string {
	lst := fmt.Sprintf(" #  Mnem   PMB  I/O Busy Done Status%c", ASCII_NL)
	var line string
	for dev := range d {
		if d[dev].mnemonic != "" {
			line = fmt.Sprintf("%#3o %-6s %2d. %3d %4d %4d  ",
				dev, d[dev].mnemonic, d[dev].priorityMaskBit,
				BoolToInt(d[dev].ioDevice), BoolToInt(d[dev].busy), BoolToInt(d[dev].done))
			if d[dev].simAttached {
				line += "Attached"
			} else {
				line += "Not Attached"
			}
			if d[dev].bootable {
				line += ", Bootable"
			}
			line += "\012"
			lst += line
		}
	}
	return lst
}
