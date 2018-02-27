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
	"log"
	"os"
	"sync"
	"time"

	"github.com/SMerrony/dgemug/logging"

	"github.com/SMerrony/dgemug/util"

	"github.com/SMerrony/dgemug/memory"

	"github.com/SMerrony/dgemug"

	"github.com/SMerrony/simhtape/pkg/simhtape"
)

const (
	mtbMaxRecordSizeW = 16384
	mtbMaxRecordSizeB = mtbMaxRecordSizeW * 2
	mtbEOF            = 0
	mtbCmdCount       = 11
	mtbCmdMask        = 0x00b8

	mtbCmdReadBits        = 0x0000
	mtbCmdRewindBits      = 0x0008
	mtbCmdCtrlModeBits    = 0x0010
	mtbCmdSpaceFwdBits    = 0x0018
	mtbCmdSpaceRevBits    = 0x0020
	mtbCmdWiteBits        = 0x0028
	mtbCmdWriteEOFBits    = 0x0030
	mtbCmdEraseBits       = 0x0038
	mtbCmdReadNonStopBits = 0x0080
	mtbCmdUnloadBits      = 0x0088
	mtbCmdDriveModeBits   = 0x0090

	mtbCmdRead        = 0
	mtbCmdRewind      = 1
	mtbCmdCtrlMode    = 2
	mtbCmdSpaceFwd    = 3
	mtbCmdSpaceRev    = 4
	mtbCmdWrite       = 5
	mtbCmdWriteEOF    = 6
	mtbCmdErase       = 7
	mtbCmdReadNonStop = 8
	mtbCmdUnload      = 9
	mtbCmdDriveMode   = 10

	mtbSr1Error         = 1 << 15
	mtbSr1DataLate      = 1 << 14
	mtbSr1Rewinding     = 1 << 13
	mtbSr1Illegal       = 1 << 12
	mtbSr1HiDensity     = 1 << 11
	mtbSr1DataError     = 1 << 10
	mtbSr1EOT           = 1 << 9
	mtbSr1EOF           = 1 << 8
	mtbSr1BOT           = 1 << 7
	mtbSr19Track        = 1 << 6
	mtbSr1BadTape       = 1 << 5
	mtbSr1Reserved      = 1 << 4
	mtbSr1StatusChanged = 1 << 3
	mtbSr1WriteLock     = 1 << 2
	mtbSr1OddChar       = 1 << 1
	mtbSr1UnitReady     = 1

	mtbSr2Error  = 1 << 15
	mtbSr2PEMode = 1
)

const mtbStatsPeriodMs = 500 // Will update status this often

const maxTapes = 8

type mtbT struct {
	mtbDataMu              sync.RWMutex
	imageAttached          [maxTapes]bool
	fileName               [maxTapes]string
	simhFile               [maxTapes]*os.File
	statusReg1, statusReg2 dg.WordT
	memAddrReg             dg.PhysAddrT
	negWordCntReg          int
	currentCmd             int
	currentUnit            int
	// debug                  bool
}

type mtbStatT struct {
	imageAttached          [maxTapes]bool
	fileName               [maxTapes]string
	memAddrReg             dg.PhysAddrT
	currentCmd             int
	statusReg1, statusReg2 dg.WordT
}

var (
	mtb mtbT

	commandSet [mtbCmdCount]dg.WordT
)

func mtbInit(statsChan chan mtbStatT) bool {
	commandSet[mtbCmdRead] = mtbCmdReadBits
	commandSet[mtbCmdRewind] = mtbCmdRewindBits
	commandSet[mtbCmdCtrlMode] = mtbCmdCtrlModeBits
	commandSet[mtbCmdSpaceFwd] = mtbCmdSpaceFwdBits
	commandSet[mtbCmdSpaceRev] = mtbCmdSpaceRevBits
	commandSet[mtbCmdWrite] = mtbCmdWiteBits
	commandSet[mtbCmdWriteEOF] = mtbCmdWriteEOFBits
	commandSet[mtbCmdErase] = mtbCmdEraseBits
	commandSet[mtbCmdReadNonStop] = mtbCmdReadNonStopBits
	commandSet[mtbCmdUnload] = mtbCmdUnloadBits
	commandSet[mtbCmdDriveMode] = mtbCmdDriveModeBits

	go mtbStatSender(statsChan)

	busAddDevice(devMTB, "MTB", pmbMTB, false, true, true)

	mtb.mtbDataMu.Lock()
	mtb.statusReg1 = mtbSr1HiDensity | mtbSr19Track | mtbSr1UnitReady
	mtb.statusReg2 = mtbSr2PEMode
	mtb.mtbDataMu.Unlock()

	busSetResetFunc(devMTB, mtbReset)
	busSetDataInFunc(devMTB, mtbDataIn)
	busSetDataOutFunc(devMTB, mtbDataOut)

	logging.DebugPrint(logging.MtbLog, "MTB Initialised via call to mtbInit()\n")
	return true
}

// mtbStatSender provides a near real-time view of MTB status and should be run as a Goroutine
// TODO = only handles Unit 0 at the moment
func mtbStatSender(sChan chan mtbStatT) {
	var stats mtbStatT
	logging.DebugPrint(logging.DebugLog, "dskpStatSender() started\n")
	for {
		mtb.mtbDataMu.RLock()
		if mtb.imageAttached[0] {
			stats.imageAttached[0] = true
			stats.fileName[0] = mtb.fileName[0]
			stats.memAddrReg = mtb.memAddrReg
			stats.currentCmd = mtb.currentCmd // Could decode this
			stats.statusReg1 = mtb.statusReg1
			stats.statusReg2 = mtb.statusReg2
		} else {
			stats = mtbStatT{}
		}
		mtb.mtbDataMu.RUnlock()
		// Non-blocking send of stats
		select {
		case sChan <- stats:
			// stats sent
		default:
			// do not block
		}
		time.Sleep(mtbStatsPeriodMs * time.Millisecond)
	}
}

// Reset the MTB to startup state
func mtbReset() {
	mtb.mtbDataMu.Lock()
	for t := 0; t < maxTapes; t++ {
		if mtb.imageAttached[t] {
			simhtape.Rewind(mtb.simhFile[t])
		}
	}
	// BOT is an error state...
	mtb.statusReg1 = mtbSr1Error | mtbSr1HiDensity | mtbSr19Track | mtbSr1BOT | mtbSr1UnitReady
	mtb.statusReg2 = mtbSr2PEMode
	mtb.memAddrReg = 0
	mtb.negWordCntReg = 0
	mtb.currentCmd = 0
	mtb.currentUnit = 0
	mtb.mtbDataMu.Unlock()
	logging.DebugPrint(logging.MtbLog, "MTB Reset via call to mtbReset()\n")
}

// Attach a SimH tape image file to the emulated tape drive
func mtbAttach(tNum int, imgName string) bool {
	logging.DebugPrint(logging.MtbLog, "mtbAttach called on unit #%d with image file: %s\n", tNum, imgName)
	f, err := os.Open(imgName)
	if err != nil {
		logging.DebugPrint(logging.MtbLog, "ERROR: Could not open simH Tape Image file: %s, due to: %s\n", imgName, err.Error())
		return false
	}
	mtb.mtbDataMu.Lock()
	mtb.fileName[tNum] = imgName
	mtb.simhFile[tNum] = f
	mtb.imageAttached[tNum] = true
	mtb.statusReg1 = mtbSr1Error | mtbSr1HiDensity | mtbSr19Track | mtbSr1BOT | mtbSr1UnitReady
	mtb.statusReg2 = mtbSr2PEMode
	mtb.mtbDataMu.Unlock()
	busSetAttached(devMTB, imgName)
	return true

}

func mtbDetach(tNum int) bool {
	logging.DebugPrint(logging.MtbLog, "mtbDetach called on unit #%d\n", tNum)
	mtb.mtbDataMu.Lock()
	mtb.fileName[tNum] = ""
	mtb.simhFile[tNum] = nil
	mtb.imageAttached[tNum] = false
	mtb.statusReg1 = mtbSr1Error | mtbSr1HiDensity | mtbSr19Track | mtbSr1BOT | mtbSr1UnitReady
	mtb.statusReg2 = mtbSr2PEMode
	mtb.mtbDataMu.Unlock()
	busSetDetached(devMTB)
	return true
}

// Scan the attached SimH tape image to ensure it makes sense
// (This is just a pass-through to the equivalent function in simhtape)
func mtbScanImage(tNum int) string {
	mtb.mtbDataMu.RLock()
	imageName := mtb.fileName[tNum]
	att := mtb.imageAttached[tNum]
	mtb.mtbDataMu.RUnlock()
	if !att {
		return "WARNING: No image attached"
	}
	return simhtape.ScanImage(imageName, false)
}

/* This function fakes the ROM/SCP boot-from-tape routine.
Rather than copying a ROM and executing that, we simply mimic its basic actions..
* Load file 0 from tape (1 x 2k block)
* Put the loaded code at physical location 0
* ...
*/
func mtbLoadTBoot() {
	const (
	// tbootSizeB = 2048
	// tbootSizeW = 1024
	)
	var (
		byte0, byte1 byte
		word         dg.WordT
		wdix, memix  dg.PhysAddrT
	)
	logging.DebugPrint(logging.MtbLog, "mtbLoadTBoot() called\n")
	tNum := 0
	mtb.mtbDataMu.Lock()
	defer mtb.mtbDataMu.Unlock()
	simhtape.Rewind(mtb.simhFile[tNum])
	logging.DebugPrint(logging.MtbLog, "... tape rewound\n")

readLoop:
	for {
		hdr, ok := simhtape.ReadMetaData(mtb.simhFile[tNum])
		// if !ok || hdr != tbootSizeB {
		if !ok {
			logging.DebugPrint(logging.DebugLog, "WARN: mtbLoadTBoot called when no bootable tape image attached\n")
			return
		}
		logging.DebugPrint(logging.MtbLog, "... header read, size is %d\n", hdr)
		switch hdr {
		case simhtape.SimhMtrTmk: // Tape Mark (separates files)
			break readLoop
		default:
			tbootSizeW := hdr / 2
			tapeData, ok := simhtape.ReadRecordData(mtb.simhFile[tNum], int(hdr))
			if ok {
				logging.DebugPrint(logging.MtbLog, "... data read\n")
			} else {
				logging.DebugPrint(logging.MtbLog, "... error reading data\n")
				logging.DebugPrint(logging.DebugLog, "WARNING: Could not read data in mtbLoadTBoot()\n")
				return
			}

			logging.DebugPrint(logging.MtbLog, "... loading data into memory starting at address %d\n", memix)
			for wdix = 0; wdix < dg.PhysAddrT(tbootSizeW); wdix++ {
				byte1 = tapeData[wdix*2]
				byte0 = tapeData[wdix*2+1]
				word = dg.WordT(byte1)<<8 | dg.WordT(byte0)
				memory.WriteWord(memix+wdix, word)
			}
			memix += dg.PhysAddrT(tbootSizeW)
			logging.DebugPrint(logging.MtbLog, "... finished loading data at address %d\n", memix+wdix)
			trailer, ok := simhtape.ReadMetaData(mtb.simhFile[tNum])
			if hdr != trailer || !ok {
				logging.DebugPrint(logging.DebugLog, "WARN: mtbLoadTBoot found mismatched trailer in TBOOT file\n")
			}
		}
	}
	simhtape.Rewind(mtb.simhFile[tNum])
	logging.DebugPrint(logging.MtbLog, "... tape rewound\n")
	logging.DebugPrint(logging.MtbLog, "... mtbLoadTBoot completed\n")
}

// This is called from Bus to implement DIx from the MTB device
func mtbDataIn(cpuPtr *CPUT, iPtr *novaDataIoT, abc byte) {

	mtb.mtbDataMu.RLock()
	switch abc {
	case 'A': /* Read status register 1 - see p.IV-18 of Peripherals guide */
		cpuPtr.ac[iPtr.acd] = dg.DwordT(mtb.statusReg1)
		logging.DebugPrint(logging.MtbLog, "DIA - Read Status Reg 1 %s to AC%d, PC: %d\n",
			util.WordToBinStr(mtb.statusReg1), iPtr.acd, cpuPtr.pc)

	case 'B': /* Read memory addr register 1 - see p.IV-19 of Peripherals guide */
		cpuPtr.ac[iPtr.acd] = dg.DwordT(mtb.memAddrReg)
		logging.DebugPrint(logging.MtbLog, "DIB - Read Mem Addr Reg 1 <%d> to AC%d, PC: %d\n",
			mtb.memAddrReg, iPtr.acd, cpuPtr.pc)

	case 'C': /* Read status register 2 - see p.IV-18 of Peripherals guide */
		cpuPtr.ac[iPtr.acd] = dg.DwordT(mtb.statusReg2)
		logging.DebugPrint(logging.MtbLog, "DIC - Read Status Reg 2 %s to AC%d, PC: %d\n",
			util.WordToBinStr(mtb.statusReg2), iPtr.acd, cpuPtr.pc)
	}
	mtb.mtbDataMu.RUnlock()

	mtbHandleFlag(iPtr.f)
}

// This is called from Bus to implement DOx from the MTB device
func mtbDataOut(cpuPtr *CPUT, iPtr *novaDataIoT, abc byte) {

	ac16 := util.DwordGetLowerWord(cpuPtr.ac[iPtr.acd])

	mtb.mtbDataMu.Lock()
	switch abc {
	case 'A': // Specify Command and Drive - p.IV-17
		// which command?
		for c := 0; c < mtbCmdCount; c++ {
			if (ac16 & mtbCmdMask) == commandSet[c] {
				mtb.currentCmd = c
				break
			}
		}
		// which unit?
		mtb.currentUnit = mtbExtractUnit(ac16)
		logging.DebugPrint(logging.MtbLog, "DOA - Specify Command and Drive - internal cmd #: %d, unit: %d, PC: %d\n",
			mtb.currentCmd, mtb.currentUnit, cpuPtr.pc)

	case 'B':
		mtb.memAddrReg = dg.PhysAddrT(ac16)
		logging.DebugPrint(logging.MtbLog, "DOB - Write Memory Address Register from AC%d, Value: %d, PC: %d\n",
			iPtr.acd, ac16, cpuPtr.pc)

	case 'C':
		mtb.negWordCntReg = int(int16(ac16))
		logging.DebugPrint(logging.MtbLog, "DOC - Set (neg) Word Count to %d, PC: %d\n",
			mtb.negWordCntReg, cpuPtr.pc)
	}
	mtb.mtbDataMu.Unlock()

	mtbHandleFlag(iPtr.f)
}

func mtbExtractUnit(word dg.WordT) int {
	return int(word & 0x07)
}

// mtbHandleFlag actions the flag/pulse to the MTB controller
func mtbHandleFlag(f byte) {
	switch f {
	case 'S':
		logging.DebugPrint(logging.MtbLog, "... S flag set\n")
		mtb.mtbDataMu.RLock()
		if mtb.currentCmd != mtbCmdRewind {
			busSetBusy(devMTB, true)
		}
		mtb.mtbDataMu.RUnlock()
		busSetDone(devMTB, false)
		mtbDoCommand()
		busSetBusy(devMTB, false)
		busSetDone(devMTB, true)

	case 'C':
		// if we were performing MTB operations in a Goroutine, this would interrupt them...
		logging.DebugPrint(logging.MtbLog, "... C flag set\n")
		mtb.mtbDataMu.Lock()
		mtb.statusReg1 = mtbSr1HiDensity | mtbSr19Track | mtbSr1UnitReady // ???
		mtb.statusReg2 = mtbSr2PEMode                                     // ???
		mtb.mtbDataMu.Unlock()
		busSetBusy(devMTB, false)
		busSetDone(devMTB, false)

	case 'P':
		// 'Reserved'
		logging.DebugPrint(logging.MtbLog, "WARNING: Received 'P' flag - which is reserved")

	default:
		// empty flag - nothing to do
	}
}

func mtbDoCommand() {
	mtb.mtbDataMu.Lock()
	defer mtb.mtbDataMu.Unlock()

	switch mtb.currentCmd {
	case mtbCmdRead:
		logging.DebugPrint(logging.MtbLog, "*READ* command\n ---- Unit: %d\n ---- Word Count: %d Location: %d\n", mtb.currentUnit, mtb.negWordCntReg, mtb.memAddrReg)
		hdrLen, _ := simhtape.ReadMetaData(mtb.simhFile[mtb.currentUnit])
		logging.DebugPrint(logging.MtbLog, " ----  Header read giving length: %d\n", hdrLen)
		if hdrLen == mtbEOF {
			logging.DebugPrint(logging.MtbLog, " ----  Header is EOF indicator\n")
			mtb.statusReg1 = mtbSr1HiDensity | mtbSr19Track | mtbSr1UnitReady | mtbSr1EOF | mtbSr1Error
		} else {
			logging.DebugPrint(logging.MtbLog, " ----  Calling simhtape.ReadRecord with length: %d\n", hdrLen)
			var w uint32
			var wd dg.WordT
			var pAddr dg.PhysAddrT
			rec, _ := simhtape.ReadRecordData(mtb.simhFile[mtb.currentUnit], int(hdrLen))
			for w = 0; w < hdrLen; w += 2 {
				wd = (dg.WordT(rec[w]) << 8) | dg.WordT(rec[w+1])
				pAddr = memory.WriteWordDchChan(&mtb.memAddrReg, wd)
				logging.DebugPrint(logging.MtbLog, " ----  Written word (%02X | %02X := %04X) to logical address: %d, physical: %d\n", rec[w], rec[w+1], wd, mtb.memAddrReg, pAddr)
				// memAddrReg is auto-incremented for every word written  *******
				// auto-incremement the (two's complement) word count
				mtb.negWordCntReg++
				if mtb.negWordCntReg == 0 {
					break
				}
			}
			trailer, _ := simhtape.ReadMetaData(mtb.simhFile[mtb.currentUnit])
			logging.DebugPrint(logging.MtbLog, " ----  %d bytes loaded\n", w)
			logging.DebugPrint(logging.MtbLog, " ----  Read SimH Trailer: %d\n", trailer)
			// TODO Need to verify how status should be set here...
			mtb.statusReg1 = mtbSr1HiDensity | mtbSr19Track | mtbSr1UnitReady
		}

	case mtbCmdRewind:
		logging.DebugPrint(logging.MtbLog, "*REWIND* command\n ------ Unit: #%d\n", mtb.currentUnit)
		simhtape.Rewind(mtb.simhFile[mtb.currentUnit])
		mtb.statusReg1 = mtbSr1Error | mtbSr1HiDensity | mtbSr19Track | mtbSr1UnitReady | mtbSr1BOT

	case mtbCmdSpaceFwd:
		logging.DebugPrint(logging.MtbLog, "*SPACE FORWARD* command\n ----- ------- Unit: #%d\n", mtb.currentUnit)
		simhtape.SpaceFwd(mtb.simhFile[mtb.currentUnit], 0)
		mtb.memAddrReg = 0xffffffff
		mtb.statusReg1 = mtbSr1HiDensity | mtbSr19Track | mtbSr1UnitReady | mtbSr1EOF | mtbSr1Error

	case mtbCmdSpaceRev:
		log.Fatalln("ERROR: mtbDoCommand - SPACE REVERSE command Not Yet Implemented")
	case mtbCmdUnload:
		log.Fatalln("ERROR: mtbDoCommand - UNLOAD command Not Yet Implemented")
	default:
		log.Fatalf("ERROR: mtbDoCommand - Command #%d Not Yet Implemented\n", mtb.currentCmd)
	}
}
