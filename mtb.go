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
	MTB_CMD_COUNT       = 11

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

	mtb Mtb

	mtbLog *log.Logger
)

func mtbInit() bool {
	lf, err := os.OpenFile(MTB_LOG_FILE, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalln("Failed to open MTB log file ", err.Error())
	}
	mtbLog = log.New(lf, "", log.Ldate|log.Ltime)
	simht.simhTapeInit()
	busAddDevice(DEV_MTB, "MTB", MTB_PMB, false, true, true)
	mtb.imageAttached = false

	mtb.statusReg1 = SR1_HI_DENSITY | SR1_9TRACK | SR1_UNIT_READY
	mtb.statusReg2 = SR2_PE_MODE

	busSetResetFunc(DEV_MTB, mtbReset)
	busSetDataInFunc(DEV_MTB, mtbDataIn)
	busSetDataOutFunc(DEV_MTB, mtbDataOut)

	mtbLog.Println("MTB Initialised")
	return true
}

// Reset the MTB to startup state
func mtbReset() {
	simht.simhTapeRewind(0)
	mtb.statusReg1 = SR1_HI_DENSITY | SR1_9TRACK | SR1_BOT | SR1_UNIT_READY
	mtb.statusReg2 = SR2_PE_MODE
	mtbLog.Println("MTB Reset")
}

// Attach a SimH tape image file to the emulated tape drive
func mtbAttach(tNum int, imgName string) bool {
	mtbLog.Printf("mtbAttach called on unit #%d with image file: %s\n", tNum, imgName)
	if simht.simhTapeAttach(tNum, imgName) {
		mtb.simhTapeNum = tNum
		mtb.imageAttached = true
		mtb.statusReg1 = SR1_HI_DENSITY | SR1_9TRACK | SR1_BOT | SR1_UNIT_READY
		mtb.statusReg2 = SR2_PE_MODE
		busSetAttached(DEV_MTB)
		return true
	}
	return false
}

// Scan the attached SimH tape image to ensure it makes sense
// (This is just a pass-through to the equivalent function in simhTape)
func mtbScanImage(tNum int) string {
	return simht.simhTapeScanImage(0)
}

/* This function fakes the ROM/SCP boot-from-tape routine.
Rather than copying a ROM and executing that, we simply mimic its basic actions..
* Load file 0 from tape (1 x 2k block)
* Put the loaded code at physical location 0
* ...
*/
func mtbLoadTBoot(mem Memory) {
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

// This is called from Bus to implement DIx from the MTB device
func mtbDataIn(cpuPtr *Cpu, iPtr *DecodedInstr, abc byte) {

	switch iPtr.f {
	case 'S':
		busSetBusy(DEV_TTO, true)
		busSetDone(DEV_TTO, false)
	case 'C':
		busSetBusy(DEV_TTO, false)
		busSetDone(DEV_TTO, false)
	}

	switch abc {
	case 'A': /* Read status register 1 - see p.IV-18 of Peripherals guide */
		cpuPtr.ac[iPtr.acd] = dg_dword(mtb.statusReg1)
		mtbLog.Printf("DIA - Read Status Reg 1 %s to AC%d, PC: %d\n",
			wordToBinStr(mtb.statusReg1), iPtr.acd, cpuPtr.pc)

	case 'B': /* Read memory addr register 1 - see p.IV-19 of Peripherals guide */
		cpuPtr.ac[iPtr.acd] = dg_dword(mtb.memAddrReg)
		mtbLog.Printf("DIB - Read Mem Addr Reg 1 <%d> to AC%d, PC: %dn",
			mtb.memAddrReg, iPtr.acd, cpuPtr.pc)

	case 'C': /* Read status register 2 - see p.IV-18 of Peripherals guide */
		cpuPtr.ac[iPtr.acd] = dg_dword(mtb.statusReg2)
		mtbLog.Printf("DIC - Read Status Reg 2 %s to AC%d, PC: %d\n",
			wordToBinStr(mtb.statusReg2), iPtr.acd, cpuPtr.pc)
	}

	if iPtr.f == 'S' {
		mtbDoCommand() // TODO Can this be a goroutine?
	}
}

// This is called from Bus to implement DOx from the MTB device
func mtbDataOut(cpuPtr *Cpu, iPtr *DecodedInstr, abc byte) {

	switch iPtr.f {
	case 'S':
		busSetBusy(DEV_TTO, true)
		busSetDone(DEV_TTO, false)
	case 'C':
		busSetBusy(DEV_TTO, false)
		busSetDone(DEV_TTO, false)
	}

	ac16 := dwordGetLowerWord(cpuPtr.ac[iPtr.acd])

	switch abc {
	case 'A': // Specify Command and Drive - p.IV-17
		// which command?
		for c := 0; c < MTB_CMD_COUNT; c++ {
			if (ac16 & COMMAND_MASK) == commandSet[c] {
				mtb.currentCmd = c
				break
			}
		}
		mtbLog.Printf("DOA - Specify Command and Drive - internal cmd #: %d, PC: %d\n",
			mtb.currentCmd, cpuPtr.pc)

	case 'B':
		mtb.memAddrReg = ac16
		mtbLog.Printf("DOB - Write Memory Address Register from AC%d, Value: %d, PC: %d\n",
			iPtr.acd, ac16, cpuPtr.pc)

	case 'C':
		mtb.negWordCntReg = ac16
		mtbLog.Printf("DOC - Set (neg) Word Count to %d, PC: %d\n",
			mtb.negWordCntReg, cpuPtr.pc)
	}

	if iPtr.f == 'S' {
		mtbDoCommand() // TODO Can this be a goroutine?
	}
}
