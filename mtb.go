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
)

const (
	mtbMaxRecordSize = 16384
	mtbEOF           = 0
	mtbCmdCount      = 11
	mtbCmdMask       = 0x00b8

	mtbCmdREAD_BITS         = 0x0000
	mtbCmdREWIND_BITS       = 0x0008
	mtbCmdCTRL_MODE_BITS    = 0x0010
	mtbCmdSPACE_FWD_BITS    = 0x0018
	mtbCmdSPACE_REV_BITS    = 0x0020
	mtbCmdWRITE_BITS        = 0x0028
	mtbCmdWRITE_EOF_BITS    = 0x0030
	mtbCmdERASE_BITS        = 0x0038
	mtbCmdREAD_NONSTOP_BITS = 0x0080
	mtbCmdUNLOAD_BITS       = 0x0088
	mtbCmdDRIVE_MODE_BITS   = 0x0090

	mtbCmdREAD         = 0
	mtbCmdREWIND       = 1
	mtbCmdCTRL_MODE    = 2
	mtbCmdSPACE_FWD    = 3
	mtbCmdSPACE_REV    = 4
	mtbCmdWRITE        = 5
	mtbCmdWRITE_EOF    = 6
	mtbCmdERASE        = 7
	mtbCmdREAD_NONSTOP = 8
	mtbCmdUNLOAD       = 9
	mtbCmdDRIVE_MODE   = 10

	mtbSr1Error     = 0100000
	mtbSr1HiDensity = 04000
	mtbSr1EOT       = 01000
	mtbSr1EOF       = 0400
	mtbSr1BOT       = 0200
	mtbSr19Track    = 0100
	mtbSr1UnitReady = 0001

	mtbSr2Error     = 0x8000
	mtbSr2DataError = 0x0400
	mtbSr2EOT       = 0x0200
	mtbSr2EOF       = 0x0100
	mtbSr2PEMode    = 0x0001
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
	commandSet [mtbCmdCount]dg.WordT
)

func mtbInit() bool {
	commandSet[mtbCmdREAD] = mtbCmdREAD_BITS
	commandSet[mtbCmdREWIND] = mtbCmdREWIND_BITS
	commandSet[mtbCmdCTRL_MODE] = mtbCmdCTRL_MODE_BITS
	commandSet[mtbCmdSPACE_FWD] = mtbCmdSPACE_FWD_BITS
	commandSet[mtbCmdSPACE_REV] = mtbCmdSPACE_REV_BITS
	commandSet[mtbCmdWRITE] = mtbCmdWRITE_BITS
	commandSet[mtbCmdWRITE_EOF] = mtbCmdWRITE_EOF_BITS
	commandSet[mtbCmdERASE] = mtbCmdERASE_BITS
	commandSet[mtbCmdREAD_NONSTOP] = mtbCmdREAD_NONSTOP_BITS
	commandSet[mtbCmdUNLOAD] = mtbCmdUNLOAD_BITS
	commandSet[mtbCmdDRIVE_MODE] = mtbCmdDRIVE_MODE_BITS

	simht.simhTapeInit()
	busAddDevice(DEV_MTB, "MTB", MTB_PMB, false, true, true)
	mtb.imageAttached = false

	mtb.statusReg1 = mtbSr1HiDensity | mtbSr19Track | mtbSr1UnitReady
	mtb.statusReg2 = mtbSr2PEMode

	busSetResetFunc(DEV_MTB, mtbReset)
	busSetDataInFunc(DEV_MTB, mtbDataIn)
	busSetDataOutFunc(DEV_MTB, mtbDataOut)

	logging.DebugPrint(logging.MtbLog, "MTB Initialised via call to mtbInit()\n")
	return true
}

// Reset the MTB to startup state
func mtbReset() {
	simht.Rewind(0)
	mtb.statusReg1 = mtbSr1HiDensity | mtbSr19Track | mtbSr1BOT | mtbSr1UnitReady
	mtb.statusReg2 = mtbSr2PEMode
	logging.DebugPrint(logging.MtbLog, "MTB Reset via call to mtbReset()\n")
}

// Attach a SimH tape image file to the emulated tape drive
func mtbAttach(tNum int, imgName string) bool {
	logging.DebugPrint(logging.MtbLog, "mtbAttach called on unit #%d with image file: %s\n", tNum, imgName)
	if simht.Attach(tNum, imgName) {
		mtb.simhTapeNum = tNum
		mtb.imageAttached = true
		mtb.statusReg1 = mtbSr1HiDensity | mtbSr19Track | mtbSr1BOT | mtbSr1UnitReady
		mtb.statusReg2 = mtbSr2PEMode
		busSetAttached(DEV_MTB)
		return true
	}
	return false
}

// Scan the attached SimH tape image to ensure it makes sense
// (This is just a pass-through to the equivalent function in simhTape)
func mtbScanImage(tNum int) string {
	return simht.ScanImage(0)
}

/* This function fakes the ROM/SCP boot-from-tape routine.
Rather than copying a ROM and executing that, we simply mimic its basic actions..
* Load file 0 from tape (1 x 2k block)
* Put the loaded code at physical location 0
* ...
*/
func mtbLoadTBoot() {
	const (
		tbootSizeB = 2048
		tbootSizeW = 1024
	)
	simht.Rewind(0)
	hdr, ok := simht.ReadRecordHeader(0)
	if !ok || hdr != tbootSizeB {
		logging.DebugPrint(logging.DebugLog, "WARN: mtbLoadTBoot called when no bootable tape image attached\n")
		return
	}
	tapeData, ok := simht.ReadRecordData(0, tbootSizeB)
	var byte0, byte1 byte
	var word dg.WordT
	var wdix dg.PhysAddrT
	for wdix = 0; wdix < tbootSizeW; wdix++ {
		byte1 = tapeData[wdix*2]
		byte0 = tapeData[wdix*2+1]
		word = dg.WordT(byte1)<<8 | dg.WordT(byte0)
		memory.WriteWord(wdix, word)
	}
	trailer, ok := simht.ReadRecordHeader(0)
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
		logging.DebugPrint(logging.MtbLog, "DIA - Read Status Reg 1 %s to AC%d, PC: %d\n",
			util.WordToBinStr(mtb.statusReg1), iPtr.acd, cpuPtr.pc)

	case 'B': /* Read memory addr register 1 - see p.IV-19 of Peripherals guide */
		cpuPtr.ac[iPtr.acd] = dg.DwordT(mtb.memAddrReg)
		logging.DebugPrint(logging.MtbLog, "DIB - Read Mem Addr Reg 1 <%d> to AC%d, PC: %dn",
			mtb.memAddrReg, iPtr.acd, cpuPtr.pc)

	case 'C': /* Read status register 2 - see p.IV-18 of Peripherals guide */
		cpuPtr.ac[iPtr.acd] = dg.DwordT(mtb.statusReg2)
		logging.DebugPrint(logging.MtbLog, "DIC - Read Status Reg 2 %s to AC%d, PC: %d\n",
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
		for c := 0; c < mtbCmdCount; c++ {
			if (ac16 & mtbCmdMask) == commandSet[c] {
				mtb.currentcmd = c
				break
			}
		}
		logging.DebugPrint(logging.MtbLog, "DOA - Specify Command and Drive - internal cmd #: %d, PC: %d\n",
			mtb.currentcmd, cpuPtr.pc)

	case 'B':
		mtb.memAddrReg = dg.PhysAddrT(ac16)
		logging.DebugPrint(logging.MtbLog, "DOB - Write Memory Address Register from AC%d, Value: %d, PC: %d\n",
			iPtr.acd, ac16, cpuPtr.pc)

	case 'C':
		mtb.negWordCntReg = int(int16(ac16))
		logging.DebugPrint(logging.MtbLog, "DOC - Set (neg) Word Count to %d, PC: %d\n",
			mtb.negWordCntReg, cpuPtr.pc)
	}

	if iPtr.f == 'S' {
		mtbDoCommand()
	}
}

func mtbDoCommand() {

	switch mtb.currentcmd {
	case mtbCmdREAD:
		logging.DebugPrint(logging.MtbLog, "*READ* command\n ==== Word Count: %d Location: %d\n", mtb.negWordCntReg, mtb.memAddrReg)
		hdrLen, _ := simht.ReadRecordHeader(0)
		logging.DebugPrint(logging.MtbLog, " ----  Header read giving length: %d\n", hdrLen)
		if hdrLen == mtbEOF {
			logging.DebugPrint(logging.MtbLog, " ----  Header is EOF indicator\n")
			mtb.statusReg1 = mtbSr1HiDensity | mtbSr19Track | mtbSr1UnitReady | mtbSr1EOF | mtbSr1Error
		} else {
			logging.DebugPrint(logging.MtbLog, " ----  Calling simhTapeReadRecord with length: %d\n", hdrLen)
			var w dg.DwordT
			var wd dg.WordT
			var pAddr dg.PhysAddrT
			rec, _ := simht.ReadRecordData(0, int(hdrLen))
			for w = 0; w < hdrLen; w += 2 {
				wd = (dg.WordT(rec[w]) << 8) | dg.WordT(rec[w+1])
				pAddr = memory.MemWriteWordDchChan(mtb.memAddrReg, wd)
				logging.DebugPrint(logging.MtbLog, " ----  Written word (%02X | %02X := %04X) to logical address: %d, physical: %d\n", rec[w], rec[w+1], wd, mtb.memAddrReg, pAddr)
				mtb.memAddrReg++
				mtb.negWordCntReg++
				if mtb.negWordCntReg == 0 {
					break
				}
			}
			trailer, _ := simht.ReadRecordHeader(0)
			logging.DebugPrint(logging.MtbLog, " ----  %d bytes loaded\n", w)
			logging.DebugPrint(logging.MtbLog, " ----  Read SimH Trailer: %d\n", trailer)
			// TODO Need to verify how status should be set here...
			mtb.statusReg1 = mtbSr1HiDensity | mtbSr19Track | mtbSr1UnitReady
		}
		busSetBusy(DEV_MTB, false)
		busSetDone(DEV_MTB, true)

	case mtbCmdREWIND:
		logging.DebugPrint(logging.MtbLog, "*REWIND* command\n")
		simht.Rewind(0)
		mtb.statusReg1 = mtbSr1HiDensity | mtbSr19Track | mtbSr1UnitReady | mtbSr1BOT
		mtb.statusReg2 = mtbSr2PEMode
		// FIXME set flags here?

	case mtbCmdSPACE_FWD:
		logging.DebugPrint(logging.MtbLog, "*SPACE FORWARD* command\n")
		simht.SpaceFwd(0, 0)
		mtb.statusReg1 = mtbSr1HiDensity | mtbSr19Track | mtbSr1UnitReady | mtbSr1EOF | mtbSr1Error
		busSetBusy(DEV_MTB, false)
		busSetDone(DEV_MTB, true)

	default:
		log.Fatalln("ERROR: mtbDoCommand - Command Not Yet Implemented")
	}
}
