// mtb.go

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
	//"bytes"
	"log"
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
	"mvemg/util"
	"os"
)

const (
	mtb_MAX_RECORD_SIZE = 16384
	mtb_EOF             = 0
	mtb_LOG_FILE        = "mtb_debug.log"
	mtb_cmd_COUNT       = 11
	mtb_cmd_MASK        = 0x00b8

	cmd_READ_BITS         = 0x0000
	cmd_REWIND_BITS       = 0x0008
	cmd_CTRL_MODE_BITS    = 0x0010
	cmd_SPACE_FWD_BITS    = 0x0018
	cmd_SPACE_REV_BITS    = 0x0020
	cmd_WRITE_BITS        = 0x0028
	cmd_WRITE_EOF_BITS    = 0x0030
	cmd_ERASE_BITS        = 0x0038
	cmd_READ_NONSTOP_BITS = 0x0080
	cmd_UNLOAD_BITS       = 0x0088
	cmd_DRIVE_MODE_BITS   = 0x0090

	cmd_READ         = 0
	cmd_REWIND       = 1
	cmd_CTRL_MODE    = 2
	cmd_SPACE_FWD    = 3
	cmd_SPACE_REV    = 4
	cmd_WRITE        = 5
	cmd_WRITE_EOF    = 6
	cmd_ERASE        = 7
	cmd_READ_NONSTOP = 8
	cmd_UNLOAD       = 9
	cmd_DRIVE_MODE   = 10

	sr1_ERROR      = 0100000
	sr1_HI_DENSITY = 04000
	sr1_EOT        = 01000
	sr1_EOF        = 0400
	sr1_BOT        = 0200
	sr1_9TRACK     = 0100
	sr1_UNIT_READY = 0001

	sr2_ERROR      = 0x8000
	sr2_DATA_ERROR = 0x0400
	sr2_EOT        = 0x0200
	sr2_EOF        = 0x0100
	sr2_PE_MODE    = 0x0001
)

// const mtbStatsPeriodMs = 500

// type mtbStatT struct {
// 	imageAttached bool

// }

type mtbT struct {
	simhTapeNum            int
	imageAttached          bool
	statusReg1, statusReg2 dg.WordT
	memAddrReg             dg.PhysAddrT
	negWordCntReg          int
	currentcmd             int
	debug                  bool
}

var (
	simht SimhTapesT

	mtb        mtbT
	commandSet [mtb_cmd_COUNT]dg.WordT

	mtbLog *log.Logger
)

func mtbInit() bool {
	commandSet[cmd_READ] = cmd_READ_BITS
	commandSet[cmd_REWIND] = cmd_REWIND_BITS
	commandSet[cmd_CTRL_MODE] = cmd_CTRL_MODE_BITS
	commandSet[cmd_SPACE_FWD] = cmd_SPACE_FWD_BITS
	commandSet[cmd_SPACE_REV] = cmd_SPACE_REV_BITS
	commandSet[cmd_WRITE] = cmd_WRITE_BITS
	commandSet[cmd_WRITE_EOF] = cmd_WRITE_EOF_BITS
	commandSet[cmd_ERASE] = cmd_ERASE_BITS
	commandSet[cmd_READ_NONSTOP] = cmd_READ_NONSTOP_BITS
	commandSet[cmd_UNLOAD] = cmd_UNLOAD_BITS
	commandSet[cmd_DRIVE_MODE] = cmd_DRIVE_MODE_BITS

	lf, err := os.OpenFile(mtb_LOG_FILE, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalln("Failed to open MTB log file ", err.Error())
	}
	mtbLog = log.New(lf, "", log.Ldate|log.Ltime)
	simht.simhTapeInit()
	busAddDevice(DEV_MTB, "MTB", MTB_PMB, false, true, true)
	mtb.imageAttached = false

	mtb.statusReg1 = sr1_HI_DENSITY | sr1_9TRACK | sr1_UNIT_READY
	mtb.statusReg2 = sr2_PE_MODE

	busSetResetFunc(DEV_MTB, mtbReset)
	busSetDataInFunc(DEV_MTB, mtbDataIn)
	busSetDataOutFunc(DEV_MTB, mtbDataOut)

	mtbLog.Println("MTB Initialised")
	return true
}

// Reset the MTB to startup state
func mtbReset() {
	simht.simhTapeRewind(0)
	mtb.statusReg1 = sr1_HI_DENSITY | sr1_9TRACK | sr1_BOT | sr1_UNIT_READY
	mtb.statusReg2 = sr2_PE_MODE
	mtbLog.Println("MTB Reset")
}

// Attach a SimH tape image file to the emulated tape drive
func mtbAttach(tNum int, imgName string) bool {
	mtbLog.Printf("mtbAttach called on unit #%d with image file: %s\n", tNum, imgName)
	if simht.simhTapeAttach(tNum, imgName) {
		mtb.simhTapeNum = tNum
		mtb.imageAttached = true
		mtb.statusReg1 = sr1_HI_DENSITY | sr1_9TRACK | sr1_BOT | sr1_UNIT_READY
		mtb.statusReg2 = sr2_PE_MODE
		busSetAttached(DEV_MTB)
		return true
	}
	return false
}

// Scan the attached SimH tape image to ensure it makes sense
// (This is just a pass-through to the equivalent function in simhTape)
func mtbScanImage(tNum int) string {
	return simht.SimhTapeScanImage(0)
}

/* This function fakes the ROM/SCP boot-from-tape routine.
Rather than copying a ROM and executing that, we simply mimic its basic actions..
* Load file 0 from tape (1 x 2k block)
* Put the loaded code at physical location 0
* ...
*/
func mtbLoadTBoot() {
	const (
		TBTSIZ_B = 2048
		TBTSIZ_W = 1024
	)
	simht.simhTapeRewind(0)
	hdr, ok := simht.simhTapeReadRecordHeader(0)
	if !ok || hdr != TBTSIZ_B {
		logging.DebugPrint(logging.DebugLog, "WARN: mtbLoadTBoot called when no bootable tape image attached\n")
		return
	}
	tapeData, ok := simht.simhTapeReadRecord(0, TBTSIZ_B)
	var byte0, byte1 byte
	var word dg.WordT
	var wdix dg.PhysAddrT
	for wdix = 0; wdix < TBTSIZ_W; wdix++ {
		byte1 = tapeData[wdix*2]
		byte0 = tapeData[wdix*2+1]
		word = dg.WordT(byte1)<<8 | dg.WordT(byte0)
		memory.WriteWord(wdix, word)
	}
	trailer, ok := simht.simhTapeReadRecordHeader(0)
	if hdr != trailer {
		logging.DebugPrint(logging.DebugLog, "WARN: mtbLoadTBoot found mismatched trailer in TBOOT file\n")
	}
}

// This is called from Bus to implement DIx from the MTB device
func mtbDataIn(cpuPtr *CPUT, iPtr *novaDataIoT, abc byte) {

	switch iPtr.f {
	case 'S':
		busSetBusy(DEV_MTB, true)
		busSetDone(DEV_MTB, false)
	case 'C':
		busSetBusy(DEV_MTB, false)
		busSetDone(DEV_MTB, false)
	}

	switch abc {
	case 'A': /* Read status register 1 - see p.IV-18 of Peripherals guide */
		cpuPtr.ac[iPtr.acd] = dg.DwordT(mtb.statusReg1)
		mtbLog.Printf("DIA - Read Status Reg 1 %s to AC%d, PC: %d\n",
			util.WordToBinStr(mtb.statusReg1), iPtr.acd, cpuPtr.pc)

	case 'B': /* Read memory addr register 1 - see p.IV-19 of Peripherals guide */
		cpuPtr.ac[iPtr.acd] = dg.DwordT(mtb.memAddrReg)
		mtbLog.Printf("DIB - Read Mem Addr Reg 1 <%d> to AC%d, PC: %dn",
			mtb.memAddrReg, iPtr.acd, cpuPtr.pc)

	case 'C': /* Read status register 2 - see p.IV-18 of Peripherals guide */
		cpuPtr.ac[iPtr.acd] = dg.DwordT(mtb.statusReg2)
		mtbLog.Printf("DIC - Read Status Reg 2 %s to AC%d, PC: %d\n",
			util.WordToBinStr(mtb.statusReg2), iPtr.acd, cpuPtr.pc)
	}

	if iPtr.f == 'S' {
		mtbDoCommand()
	}
}

// This is called from Bus to implement DOx from the MTB device
func mtbDataOut(cpuPtr *CPUT, iPtr *novaDataIoT, abc byte) {

	switch iPtr.f {
	case 'S':
		busSetBusy(DEV_MTB, true)
		busSetDone(DEV_MTB, false)
	case 'C':
		busSetBusy(DEV_MTB, false)
		busSetDone(DEV_MTB, false)
	}

	ac16 := util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])

	switch abc {
	case 'A': // Specify Command and Drive - p.IV-17
		// which command?
		for c := 0; c < mtb_cmd_COUNT; c++ {
			if (ac16 & mtb_cmd_MASK) == commandSet[c] {
				mtb.currentcmd = c
				break
			}
		}
		mtbLog.Printf("DOA - Specify Command and Drive - internal cmd #: %d, PC: %d\n",
			mtb.currentcmd, cpuPtr.pc)

	case 'B':
		mtb.memAddrReg = dg.PhysAddrT(ac16)
		mtbLog.Printf("DOB - Write Memory Address Register from AC%d, Value: %d, PC: %d\n",
			iPtr.acd, ac16, cpuPtr.pc)

	case 'C':
		mtb.negWordCntReg = int(int16(ac16))
		mtbLog.Printf("DOC - Set (neg) Word Count to %d, PC: %d\n",
			mtb.negWordCntReg, cpuPtr.pc)
	}

	if iPtr.f == 'S' {
		mtbDoCommand()
	}
}

func mtbDoCommand() {

	switch mtb.currentcmd {
	case cmd_READ:
		mtbLog.Printf("*READ* command\n ==== Word Count: %d Location: %d\n", mtb.negWordCntReg, mtb.memAddrReg)
		hdrLen, _ := simht.simhTapeReadRecordHeader(0)
		mtbLog.Printf(" ----  Header read giving length: %d\n", hdrLen)
		if hdrLen == mtb_EOF {
			mtbLog.Printf(" ----  Header is EOF indicator\n")
			mtb.statusReg1 = sr1_HI_DENSITY | sr1_9TRACK | sr1_UNIT_READY | sr1_EOF | sr1_ERROR
		} else {
			mtbLog.Printf(" ----  Calling simhTapeReadRecord with length: %d\n", hdrLen)
			var w dg.DwordT
			var wd dg.WordT
			var pAddr dg.PhysAddrT
			rec, _ := simht.simhTapeReadRecord(0, int(hdrLen))
			for w = 0; w < hdrLen; w += 2 {
				wd = (dg.WordT(rec[w]) << 8) | dg.WordT(rec[w+1])
				pAddr = memory.MemWriteWordDchChan(mtb.memAddrReg, wd)
				mtbLog.Printf(" ----  Written word (%02X | %02X := %04X) to logical address: %d, physical: %d\n", rec[w], rec[w+1], wd, mtb.memAddrReg, pAddr)
				mtb.memAddrReg++
				mtb.negWordCntReg++
				if mtb.negWordCntReg == 0 {
					break
				}
			}
			trailer, _ := simht.simhTapeReadRecordHeader(0)
			mtbLog.Printf(" ----  %d bytes loaded\n", w)
			mtbLog.Printf(" ----  Read SimH Trailer: %d\n", trailer)
			// TODO Need to verify how status should be set here...
			mtb.statusReg1 = sr1_HI_DENSITY | sr1_9TRACK | sr1_UNIT_READY
		}
		busSetBusy(DEV_MTB, false)
		busSetDone(DEV_MTB, true)

	case cmd_REWIND:
		mtbLog.Printf("*REWIND* command\n")
		simht.simhTapeRewind(0)
		mtb.statusReg1 = sr1_HI_DENSITY | sr1_9TRACK | sr1_UNIT_READY | sr1_BOT
		mtb.statusReg2 = sr2_PE_MODE
		// FIXME set flags here?

	case cmd_SPACE_FWD:
		mtbLog.Printf("*SPACE FORWARD* command\n")
		simht.SimhTapeSpaceFwd(0, 0)
		mtb.statusReg1 = sr1_HI_DENSITY | sr1_9TRACK | sr1_UNIT_READY | sr1_EOF | sr1_ERROR
		busSetBusy(DEV_MTB, false)
		busSetDone(DEV_MTB, true)

	default:
		log.Fatalln("ERROR: mtbDoCommand - Command Not Yet Implemented")
	}
}
