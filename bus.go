// bus.go
package main

import (
	"fmt"
	"log"
	//"os"
)

// I/O reset func
type ResetFunc func()

// DOx func
type DataOutFunc func(*Cpu, *DecodedInstr, byte)

// DIx func
type DataInFunc func(*Cpu, *DecodedInstr, byte)

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
	log.Printf("INFO: Device 0%o added to bus\n", devNum)
}

func busSetDataInFunc(devNum int, fn DataInFunc) {
	d[devNum].dataInFunc = fn
	log.Printf("INFO: Bus Data In function set for dev #0%o\n", devNum)
}

func busDataIn(cpuPtr *Cpu, iPtr *DecodedInstr, abc byte) {
	log.Printf("DEBUG: Bus Data In function called for dev #0%o\n", iPtr.ioDev)
	if d[iPtr.ioDev].dataInFunc == nil {
		log.Fatal("ERROR: busDataIn called with no function set")
	}
	d[iPtr.ioDev].dataInFunc(cpuPtr, iPtr, abc)
	log.Printf("INFO: Bus Data In function called for dev #0%o\n", iPtr.ioDev)
}

func busSetDataOutFunc(devNum int, fn DataOutFunc) {
	d[devNum].dataOutFunc = fn
	log.Printf("INFO: Bus Data Out function set for dev #0%o\n", devNum)
}

func busDataOut(cpuPtr *Cpu, iPtr *DecodedInstr, abc byte) {
	if d[iPtr.ioDev].dataOutFunc == nil {
		log.Fatal("ERROR: busDataOut called with no function set")
	}
	d[iPtr.ioDev].dataOutFunc(cpuPtr, iPtr, abc)
	log.Printf("INFO: Bus Data Out function called for dev #0%o\n", iPtr.ioDev)
}

func busSetResetFunc(devNum int, resetFn ResetFunc) {
	d[devNum].resetFunc = resetFn
	log.Printf("INFO: Bus reset function set for dev #0%o\n", devNum)
}

func busResetDevice(devNum int) {
	if d[devNum].ioDevice {
		d[devNum].resetFunc()
	} else {
		log.Fatalf("ERROR: Attepmt to reset non-I/O device #0%o\n", devNum)
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
	log.Printf("... Busy flag set to %d for device #0%o\n", boolToInt(f), devNum)
}

func busSetDone(devNum int, f bool) {
	d[devNum].done = f
	log.Printf("... Done flag set to %d for device #0%o\n", boolToInt(f), devNum)
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

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func busGetPrintableDevList() string {
	lst := fmt.Sprintf(" #  Mnem   PMB  I/O Busy Done Status%c", ASCII_NL)
	var line string
	for dev := range d {
		if d[dev].mnemonic != "" {
			line = fmt.Sprintf("%#3o %-6s %2d. %3d %4d %4d  ",
				dev, d[dev].mnemonic, d[dev].priorityMaskBit,
				boolToInt(d[dev].ioDevice), boolToInt(d[dev].busy), boolToInt(d[dev].done))
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
