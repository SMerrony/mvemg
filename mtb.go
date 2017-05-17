// mtb.go
package main

import (
	//"bytes"
	"log"
	"os"
)

const (
	MTB_MAX_RECORD_SIZE = 16384
	MTB_LOG_FILE        = "mtb_debug.log"

	SR1_ERROR      = 0100000
	SR1_HI_DENSITY = 04000
	SR1_EOT        = 01000
	SR1_EOF        = 0400
	SR1_BOT        = 0200
	SR1_9TRACK     = 0100
	SR1_UNIT_READY = 0001

	SR2_ERROR      = 0x8000
	SR2_DATA_ERROR = 0x0400
	SR2_EOT        = 0x0200
	SR2_EOF        = 0x0100
	SR2_PE_MODE    = 0x0001
)

type Mtb struct {
	simhTapeNum            int
	imageAttached          bool
	statusReg1, statusReg2 dg_word
	memAddrReg             dg_phys_addr
	negWordCntReg          int
	currentCmd             int
	debug                  bool
}

var (
	simht SimhTapes

	mtbLog *log.Logger
)

func (mtb *Mtb) mtbInit() bool {
	lf, err := os.OpenFile(MTB_LOG_FILE, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalln("Failed to open MTB log file ", err.Error())
	}
	mtbLog = log.New(lf, "", log.Ldate|log.Ltime)
	simht.simhTapeInit()
	bus.busAddDevice(DEV_MTB, "MTB", MTB_PMB, false, true, true)
	mtb.imageAttached = false

	mtb.statusReg1 = SR1_HI_DENSITY | SR1_9TRACK | SR1_UNIT_READY
	mtb.statusReg2 = SR2_PE_MODE

	bus.busSetResetFunc(DEV_MTB, mtb.mtbReset)

	mtbLog.Println("MTB Initialised")
	return true
}

// Reset the MTB to startup state
func (mtb *Mtb) mtbReset() {
	simht.simhTapeRewind(0)
	mtb.statusReg1 = SR1_HI_DENSITY | SR1_9TRACK | SR1_BOT | SR1_UNIT_READY
	mtb.statusReg2 = SR2_PE_MODE
	mtbLog.Println("MTB Reset")
}

// Attach a SimH tape image file to the emulated tape drive
func (mtb *Mtb) mtbAttach(tNum int, imgName string) bool {
	mtbLog.Printf("mtbAttach called on unit #%d with image file: %s\n", tNum, imgName)
	if simht.simhTapeAttach(tNum, imgName) {
		mtb.simhTapeNum = tNum
		mtb.imageAttached = true
		mtb.statusReg1 = SR1_HI_DENSITY | SR1_9TRACK | SR1_BOT | SR1_UNIT_READY
		mtb.statusReg2 = SR2_PE_MODE
		bus.busSetAttached(DEV_MTB)
		return true
	}
	return false
}

// Scan the attached SimH tape image to ensure it makes sense
// (This is just a pass-through to the equivalent function in simhTape)
func (mtb *Mtb) mtbScanImage(tNum int) string {
	return simht.simhTapeScanImage(0)
}

/* This function fakes the ROM/SCP boot-from-tape routine.
Rather than copying a ROM and executing that, we simply mimic its basic actions..
* Load file 0 from tape (1 x 2k block)
* Put the loaded code at physical location 0
* ...
*/
func (mtb *Mtb) mtbLoadTBoot(mem Memory) {
	const (
		TBTSIZ_B = 2048
		TBTSIZ_W = 1024
	)
	simht.simhTapeRewind(0)
	hdr, ok := simht.simhTapeReadRecordHeader(0)
	if !ok || hdr != TBTSIZ_B {
		log.Printf("WARN: mtbLoadTBoot called when no bootable tape image attached\n")
		return
	}
	tapeData, ok := simht.simhTapeReadRecord(0, TBTSIZ_B)
	var byte0, byte1 byte
	var word dg_word
	var wdix dg_phys_addr
	for wdix = 0; wdix < TBTSIZ_W; wdix++ {
		byte1 = tapeData[wdix*2]
		byte0 = tapeData[wdix*2+1]
		word = dg_word(byte1)<<8 | dg_word(byte0)
		memWriteWord(wdix, word)
	}
	trailer, ok := simht.simhTapeReadRecordHeader(0)
	if hdr != trailer {
		log.Printf("WARN: mtbLoadTBoot found mismatched trailer in TBOOT file\n")
	}
}
