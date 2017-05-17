// bus.go
package main

import (
	"fmt"
	"log"
	//"os"
)

type ResetFunc func()
type DataOutFunc func() // TODO expand when parm types defined in CPU/Instruction
type DataInFunc func()

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

type Devices [MAX_DEVICES]device

func (d *Devices) busInit() {
	for dev := range d {
		d[dev].mnemonic = ""
		d[dev].priorityMaskBit = 0
		d[dev].simAttached = false
		d[dev].ioDevice = false
		d[dev].bootable = false
		d[dev].busy = false
		d[dev].done = false
	}
}

func (d *Devices) busAddDevice(devNum int, mnem string, pmb int, att bool, io bool, boot bool) {
	if devNum >= MAX_DEVICES {
		log.Fatalf("ERROR: Attempt to add device with too-high device number: %o", devNum)
	}
	d[devNum].mnemonic = mnem
	d[devNum].priorityMaskBit = pmb
	d[devNum].simAttached = att
	d[devNum].ioDevice = io
	d[devNum].bootable = boot
	log.Printf("INFO: Device %o added to bus\n", devNum)
}

func (d *Devices) busSetResetFunc(devNum int, resetFn ResetFunc) {
	d[devNum].resetFunc = resetFn
	log.Printf("Bus reset function set for dev #%d\n", devNum)
}

func (d *Devices) busResetDevice(devNum int) {
	if d[devNum].ioDevice {
		d[devNum].resetFunc()
	} else {
		log.Fatalf("ERROR - Attepmt to reset non-I/O device #%o\n", devNum)
	}
}

func (d *Devices) busResetAllIODevices() {
	for dev := range d {
		if d[dev].ioDevice {
			d.busResetDevice(dev)
		}
	}
}

func (d *Devices) busSetAttached(devNum int) {
	d[devNum].simAttached = true
}
func (d *Devices) busSetDetached(devNum int) {
	d[devNum].simAttached = false
}
func (d *Devices) busIsAttached(devNum int) bool {
	return d[devNum].simAttached
}

func (d *Devices) busIsBootable(devNum int) bool {
	return d[devNum].bootable
}

func boolToInt(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

func (d *Devices) busGetPrintableDevList() string {
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
