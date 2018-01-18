// dskp.go

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

// Here we are emulating the DSKP device, specifically model 6239/6240
// controller/drive combination with 14-inch platters which provide 592MB of formatted capacity.
//
// All communication with the drive is via CPU PIO instructions and memory
// accessed via the BMC interface running at 2.2MB/sec in mapped or physical mode.
// There is also a small set of flags and pulses shared between the controller and the CPU.

// ASYNCHRONOUS interrupts occur on completion of a CB (list), or when an error
// occurs during CB processing.

// SYNCHRONOUS interrupts occur after a PIO command executes.

package main

import (
	"bufio"
	"fmt"
	"log"
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
	"mvemg/util"
	"os"
	"sync"
	"time"
)

const (
	// Physical disk characteristics
	dskpSurfacesPerDisk   = 8
	dskpHeadsPerSurface   = 2
	dskpSectorsPerTrack   = 75
	dskpWordsPerSector    = 256
	dskpBytesPerSector    = dskpWordsPerSector * 2
	dskpPhysicalCylinders = 981
	dskpUserCylinders     = 978
	dskpLogicalBlocks     = 1157952 // ??? 1147943 17<<16 | 43840
	dskpLogicalBlocksH    = dskpLogicalBlocks >> 16
	dskpLogicalBlocksL    = dskpLogicalBlocks & 0x0ffff
	dskpUcodeRev          = 99

	dskpMaxQueuedCBs = 30 // See p.2-13

	dskpIntInfBlkSize   = 8
	dskpCtrlrInfBlkSize = 2
	dskpUnitInfBlkSize  = 7
	dskpCbMaxSize       = 21
	dskpCbMinSize       = 10 //12 // Was 10

	dskpAsynchStatRetryInterval = time.Millisecond

	statXecStateResetting = 0x00
	statXecStateResetDone = 0x01
	statXecStateBegun     = 0x08
	statXecStateMapped    = 0x0c
	statXecStateDiagMode  = 0x04

	statCcsAsync        = 0
	statCcsPioInvCmd    = 1
	statCcsPioCmdFailed = 2
	statCcsPioCmdOk     = 3

	statAsyncNoErrors = 5

	// DSKP PIO Command Set
	dskpPioProgLoad        = 000
	dskpPioBegin           = 002
	dskpPioSysgen          = 025
	dskpDiagMode           = 024
	dskpSetMapping         = 026
	dskpGetMapping         = 027
	dskpSetInterface       = 030
	dskpGetInterface       = 031
	dskpSetController      = 032
	dskpGetController      = 033
	dskpSetUnit            = 034
	dskpGetUnit            = 035
	dskpGetExtendedStatus0 = 040
	dskpGetExtendedStatus1 = 041
	dskpGetExtendedStatus2 = 042
	dskpGetExtendedStatus3 = 043
	dskpStartList          = 0100
	dskpStartListHp        = 0103
	dskpRestart            = 0116
	dskpCancelList         = 0123
	dskpUnitStatus         = 0131
	dskpTrespass           = 0132
	dskpGetListStatus      = 0133
	dskpPioReset           = 0777

	// DSKP CB Command Set/OpCodes
	dskpCbOpNoOp             = 0
	dskpCbOpWrite            = 0100
	dskpCbOpWriteVerify      = 0101
	dskpCbOpWrite1Word       = 0104
	dskpCbOpWriteVerify1Word = 0105
	dskpCbOpWriteModBitmap   = 0142
	dskpCbOpRead             = 0200
	dskpCbOpReadVerify       = 0201
	dskpCbOpReadVerify1Word  = 0205
	dskpCbOpReadRawData      = 0210
	dskpCbOpReadHeaders      = 0220
	dskpCbOpReadModBitmap    = 0242
	dskpCbOpRecalibrateDisk  = 0400

	// DSKP CB FIELDS
	dskpCbLINK_ADDR_HIGH        = 0
	dskpCbLINK_ADDR_LOW         = 1
	dskpCbINA_FLAGS_OPCODE      = 2
	dskpCbPAGENO_LIST_ADDR_HIGH = 3
	dskpCbPAGENO_LIST_ADDR_LOW  = 4
	dskpCbTXFER_ADDR_HIGH       = 5
	dskpCbTXFER_ADDR_LOW        = 6
	dskpCbDEV_ADDR_HIGH         = 7
	dskpCbDEV_ADDR_LOW          = 8
	dskpCbUNIT_NO               = 9
	dskpCbTXFER_COUNT           = 10
	dskpCbCB_STATUS             = 11
	dskpCbRES1                  = 12
	dskpCbRES2                  = 13
	dskpCbERR_STATUS            = 14
	dskpCbUNIT_STATUS           = 15
	dskpCbRETRIES_DONE          = 16
	dskpCbSOFT_RTN_TXFER_COUNT  = 17
	dskpCbPHYS_CYL              = 18
	dskpCbPHYS_HEAD_SECT        = 19
	dskpCbDISK_ERR_CODE         = 20

	// Mapping bits
	dskpMapSlotLoadInts = 1 << 15
	dskpMapIntBmcPhys   = 1 << 14
	dskpMapUpstreamLoad = 1 << 13
	dskpMapUpstreamHpt  = 1 << 12

	// calculated consts
	// dskpPhysicalByteSize is the total  # bytes on a DSKP-type disk
	dskpPhysicalByteSize = dskpSurfacesPerDisk * dskpHeadsPerSurface * dskpSectorsPerTrack * dskpBytesPerSector * dskpPhysicalCylinders
	// dskpPhysicalBlockSize is the total # blocks on a DSKP-type disk
	dskpPhysicalBlockSize = dskpSurfacesPerDisk * dskpHeadsPerSurface * dskpSectorsPerTrack * dskpPhysicalCylinders
)

type dskpDataT struct {
	// MV/Em internals...
	dskpDataMu    sync.RWMutex
	imageAttached bool
	imageFileName string
	imageFile     *os.File
	reads, writes uint64
	// DG data...
	commandRegA, commandRegB, commandRegC dg.WordT
	statusRegA, statusRegB, statusRegC    dg.WordT
	isMapped                              bool
	mappingRegA, mappingRegB              dg.WordT
	intInfBlock                           [dskpIntInfBlkSize]dg.WordT
	ctrlInfBlock                          [dskpCtrlrInfBlkSize]dg.WordT
	unitInfBlock                          [dskpUnitInfBlkSize]dg.WordT
	// cylinder, head, sector                dg_word
	sectorNo dg.DwordT
}

const dskpStatsPeriodMs = 500 // Will send status update this often

type dskpStatT struct {
	imageAttached                      bool
	statusRegA, statusRegB, statusRegC dg.WordT
	//	cylinder, head, sector             dg_word
	sectorNo      dg.DwordT
	reads, writes uint64
}

var (
	dskpData dskpDataT
	cbChan   chan dg.PhysAddrT
)

// dskpInit is called once by the main routine to initialise this DSKP emulator
func dskpInit(statsChann chan dskpStatT) {
	logging.DebugPrint(logging.DskpLog, "DSKP Initialising via call to dskpInit()...\n")

	go dskpStatSender(statsChann)

	busAddDevice(devDSKP, "DSKP", dskpPMB, false, true, true)
	busSetResetFunc(devDSKP, dskpReset)
	busSetDataInFunc(devDSKP, dskpDataIn)
	busSetDataOutFunc(devDSKP, dskpDataOut)

	dskpData.dskpDataMu.Lock()
	dskpData.imageAttached = false
	dskpData.dskpDataMu.Unlock()

	cbChan = make(chan dg.PhysAddrT, dskpMaxQueuedCBs)
	go dskpCBprocessor(&dskpData)

	dskpReset()
}

// attempt to attach an extant MV/Em disk image to the running emulator
func dskpAttach(dNum int, imgName string) bool {
	// TODO Disk Number not currently used
	logging.DebugPrint(logging.DskpLog, "dskpAttach called for disk #%d with image <%s>\n", dNum, imgName)

	dskpData.dskpDataMu.Lock()

	dskpData.imageFile, err = os.OpenFile(imgName, os.O_RDWR, 0755)
	if err != nil {
		logging.DebugPrint(logging.DskpLog, "Failed to open image for attaching\n")
		logging.DebugPrint(logging.DebugLog, "WARN: Failed to open dskp image <%s> for ATTach\n", imgName)
		return false
	}
	dskpData.imageFileName = imgName
	dskpData.imageAttached = true

	dskpData.dskpDataMu.Unlock()

	busSetAttached(devDSKP)
	return true
}

func dskpStatSender(sChan chan dskpStatT) {
	var stats dskpStatT
	fmt.Printf("dskpStatSender() started\n")
	for {
		dskpData.dskpDataMu.RLock()
		if dskpData.imageAttached {
			stats.imageAttached = true
			//stats.cylinder = dskpData.cylinder
			//stats.head = dskpData.head
			//stats.sector = dskpData.sector
			stats.statusRegA = dskpData.statusRegA
			stats.statusRegB = dskpData.statusRegB
			stats.statusRegC = dskpData.statusRegC
			stats.sectorNo = dskpData.sectorNo
			stats.reads = dskpData.reads
			stats.writes = dskpData.writes
		} else {
			stats = dskpStatT{}
		}
		dskpData.dskpDataMu.RUnlock()
		// Non-blocking send of stats
		select {
		case sChan <- stats:
		default:
		}

		time.Sleep(time.Millisecond * dskpStatsPeriodMs)
	}
}

// Create an empty disk file of the correct size for the DSKP emulator to use
func dskpCreateBlank(imgName string) bool {
	newFile, err := os.Create(imgName)
	if err != nil {
		return false
	}
	defer newFile.Close()
	logging.DebugPrint(logging.DskpLog, "dskpCreateBlank attempting to write %d bytes\n", dskpPhysicalByteSize)
	w := bufio.NewWriter(newFile)
	for b := 0; b < dskpPhysicalByteSize; b++ {
		w.WriteByte(0)
	}
	w.Flush()
	return true
}

// Handle the DIA/B/C PIO commands
func dskpDataIn(cpuPtr *CPUT, iPtr *novaDataIoT, abc byte) {
	dskpData.dskpDataMu.Lock()
	switch abc {
	case 'A':
		cpuPtr.ac[iPtr.acd] = dg.DwordT(dskpData.statusRegA)
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "DIA [Read Status A] returning %s for DRV=%d, PC: %d\n", util.WordToBinStr(dskpData.statusRegA), 0, cpuPtr.pc)
		}
	case 'B':
		cpuPtr.ac[iPtr.acd] = dg.DwordT(dskpData.statusRegB)
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "DIB [Read Status B] returning %s for DRV=%d, PC: %d\n", util.WordToBinStr(dskpData.statusRegB), 0, cpuPtr.pc)
		}
	case 'C':
		cpuPtr.ac[iPtr.acd] = dg.DwordT(dskpData.statusRegC)
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "DIC [Read Status C] returning %s for DRV=%d, PC: %d\n", util.WordToBinStr(dskpData.statusRegC), 0, cpuPtr.pc)
		}
	}
	dskpData.dskpDataMu.Unlock()
	dskpHandleFlag(iPtr.f)
}

// Handle the DOA/B/C PIO commands
func dskpDataOut(cpuPtr *CPUT, iPtr *novaDataIoT, abc byte) {
	dskpData.dskpDataMu.Lock()
	switch abc {
	case 'A':
		dskpData.commandRegA = util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "DOA [Load Cmd Reg A] from AC%d containing %s, PC: %d\n",
				iPtr.acd, util.WordToBinStr(dskpData.commandRegA), cpuPtr.pc)
		}
	case 'B':
		dskpData.commandRegB = util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "DOB [Load Cmd Reg B] from AC%d containing %s, PC: %d\n",
				iPtr.acd, util.WordToBinStr(dskpData.commandRegB), cpuPtr.pc)
		}
	case 'C':
		dskpData.commandRegC = util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "DOC [Load Cmd Reg C] from AC%d containing %s, PC: %d\n",
				iPtr.acd, util.WordToBinStr(dskpData.commandRegC), cpuPtr.pc)
		}
	}
	dskpData.dskpDataMu.Unlock()
	dskpHandleFlag(iPtr.f)
}

func dskpDoPioCommand() {

	var addr, w dg.PhysAddrT

	dskpData.dskpDataMu.Lock()

	pioCmd := dskpExtractPioCommand(dskpData.commandRegC)
	switch pioCmd {
	case dskpPioProgLoad:
		log.Panicln("dskpProgLoad command not yet implemented")

	case dskpPioBegin:
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... BEGIN command, unit # %d\n", dskpData.commandRegA)
		}
		// pretend we have succesfully booted ourself
		dskpData.statusRegB = 0
		dskpSetPioStatusRegC(statXecStateBegun, statCcsPioCmdOk, dskpPioBegin, util.TestWbit(dskpData.commandRegC, 15))

	case dskpGetMapping:
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... GET MAPPING command\n")
		}
		dskpData.statusRegA = dskpData.mappingRegA
		dskpData.statusRegB = dskpData.mappingRegB
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... ... Status Reg A set to %s\n", util.WordToBinStr(dskpData.statusRegA))
			logging.DebugPrint(logging.DskpLog, "... ... Status Reg B set to %s\n", util.WordToBinStr(dskpData.statusRegB))
		}
		dskpSetPioStatusRegC(0, statCcsPioCmdOk, dskpGetMapping, util.TestWbit(dskpData.commandRegC, 15))

	case dskpSetMapping:
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... SET MAPPING command\n")
		}
		dskpData.mappingRegA = dskpData.commandRegA
		dskpData.mappingRegB = dskpData.commandRegB
		dskpData.isMapped = true
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... ... Mapping Reg A set to %s\n", util.WordToBinStr(dskpData.commandRegA))
			logging.DebugPrint(logging.DskpLog, "... ... Mapping Reg B set to %s\n", util.WordToBinStr(dskpData.commandRegB))
		}
		dskpSetPioStatusRegC(statXecStateMapped, statCcsPioCmdOk, dskpSetMapping, util.TestWbit(dskpData.commandRegC, 15))

	case dskpGetInterface:
		addr = dg.PhysAddrT(util.DWordFromTwoWords(dskpData.commandRegA, dskpData.commandRegB))
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... GET INTERFACE INFO command\n")
			logging.DebugPrint(logging.DskpLog, "... ... Destination Start Address: %d\n", addr)
		}
		for w = 0; w < dskpIntInfBlkSize; w++ {
			memory.WriteWordBmcChan(&addr, dskpData.intInfBlock[w])
			if debugLogging {
				logging.DebugPrint(logging.DskpLog, "... ... Word %d: %s\n", w, util.WordToBinStr(dskpData.intInfBlock[w]))
			}
		}
		dskpSetPioStatusRegC(0, statCcsPioCmdOk, dskpGetInterface, util.TestWbit(dskpData.commandRegC, 15))

	case dskpSetInterface:
		addr = dg.PhysAddrT(util.DWordFromTwoWords(dskpData.commandRegA, dskpData.commandRegB))
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... SET INTERFACE INFO command\n")
			logging.DebugPrint(logging.DskpLog, "... ... Origin Start Address: %d\n", addr)
		}
		// only a few fields can be changed...
		w = 5
		dskpData.intInfBlock[w], _ = memory.ReadWordBmcChan(addr + w)
		dskpData.intInfBlock[w] &= 0xff00
		w = 6
		dskpData.intInfBlock[w], _ = memory.ReadWordBmcChan(addr + w)
		w = 7
		dskpData.intInfBlock[w], _ = memory.ReadWordBmcChan(addr + w)
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... ... Word 5: %s\n", util.WordToBinStr(dskpData.intInfBlock[5]))
			logging.DebugPrint(logging.DskpLog, "... ... Word 6: %s\n", util.WordToBinStr(dskpData.intInfBlock[6]))
			logging.DebugPrint(logging.DskpLog, "... ... Word 7: %s\n", util.WordToBinStr(dskpData.intInfBlock[7]))
		}
		dskpSetPioStatusRegC(0, statCcsPioCmdOk, dskpSetInterface, util.TestWbit(dskpData.commandRegC, 15))

	case dskpGetUnit:
		addr = dg.PhysAddrT(util.DWordFromTwoWords(dskpData.commandRegA, dskpData.commandRegB))
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... GET UNIT INFO command\n")
			logging.DebugPrint(logging.DskpLog, "... ... Destination Start Address: %d\n", addr)
		}
		for w = 0; w < dskpUnitInfBlkSize; w++ {
			memory.WriteWordBmcChan(&addr, dskpData.unitInfBlock[w])
			if debugLogging {
				logging.DebugPrint(logging.DskpLog, "... ... Word %d: %s\n", w, util.WordToBinStr(dskpData.unitInfBlock[w]))
			}
		}
		dskpSetPioStatusRegC(0, statCcsPioCmdOk, dskpGetUnit, util.TestWbit(dskpData.commandRegC, 15))

	case dskpSetUnit:
		addr = dg.PhysAddrT(util.DWordFromTwoWords(dskpData.commandRegA, dskpData.commandRegB))
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... SET UNIT INFO command\n")
			logging.DebugPrint(logging.DskpLog, "... ... Origin Start Address: %d\n", addr)
		}
		// only the first word is writable according to p.2-16
		// TODO check no active CBs first
		dskpData.unitInfBlock[0] = memory.ReadWord(addr)
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... ... Overwrote word 0 of UIB with: %s\n", util.WordToBinStr(dskpData.unitInfBlock[0]))
		}
		dskpSetPioStatusRegC(0, statCcsPioCmdOk, dskpSetUnit, util.TestWbit(dskpData.commandRegC, 15))

	case dskpPioReset:
		// dskpReset() has to do its own locking...
		dskpData.dskpDataMu.Unlock()
		dskpReset()
		dskpData.dskpDataMu.Lock()

	case dskpSetController:
		addr = dg.PhysAddrT(util.DWordFromTwoWords(dskpData.commandRegA, dskpData.commandRegB))
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... SET CONTROLLER INFO command\n")
			logging.DebugPrint(logging.DskpLog, "... ... Origin Start Address: %d\n", addr)
		}
		dskpData.ctrlInfBlock[0] = memory.ReadWord(addr)
		dskpData.ctrlInfBlock[1] = memory.ReadWord(addr + 1)
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... ... Word 0: %s\n", util.WordToBinStr(dskpData.ctrlInfBlock[0]))
			logging.DebugPrint(logging.DskpLog, "... ... Word 1: %s\n", util.WordToBinStr(dskpData.ctrlInfBlock[1]))
		}
		dskpSetPioStatusRegC(0, statCcsPioCmdOk, dskpSetController, util.TestWbit(dskpData.commandRegC, 15))

	case dskpStartList:
		addr = dg.PhysAddrT(util.DWordFromTwoWords(dskpData.commandRegA, dskpData.commandRegB))
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... START LIST command\n")
			logging.DebugPrint(logging.DskpLog, "... ..... First CB Address: %d\n", addr)
			logging.DebugPrint(logging.DskpLog, "... ..... CB Channel Q length: %d\n", len(cbChan))
		}
		// TODO should check addr validity before starting processing
		//dskpProcessCB(addr)
		cbChan <- addr
		dskpData.statusRegA = util.DWordGetUpperWord(dg.DwordT(addr)) // return address of 1st CB processed
		dskpData.statusRegB = util.DWordGetLowerWord(dg.DwordT(addr))
		dskpSetPioStatusRegC(0, statCcsPioCmdOk, dskpStartList, util.TestWbit(dskpData.commandRegC, 15))

	default:
		log.Panicf("DSKP command %d not yet implemented\n", pioCmd)
	}
	dskpData.dskpDataMu.Unlock()
}

func dskpExtractPioCommand(word dg.WordT) uint {
	res := uint((word & 01776) >> 1) // mask penultimate 9 bits
	return res
}

func dskpGetCBextendedStatusSize() int {
	word := dskpData.intInfBlock[5]
	word >>= 8
	word &= 0x0f
	return int(word)
}

// Handle flag/pulse to DSKP
func dskpHandleFlag(f byte) {
	switch f {
	case 'S':
		busSetBusy(devDSKP, true)
		busSetDone(devDSKP, false)
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... S flag set\n")
		}
		dskpDoPioCommand()

		busSetBusy(devDSKP, false)
		// set the DONE flag if the return bit was set
		dskpData.dskpDataMu.RLock()
		if util.TestWbit(dskpData.commandRegC, 15) {
			busSetDone(devDSKP, true)
		}
		dskpData.dskpDataMu.RUnlock()

	case 'C':
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... C flag set, clearing DONE flag\n")
		}
		busSetDone(devDSKP, false)
		// TODO clear pending interrupt
		//dskpData.statusRegC = 0
		dskpSetPioStatusRegC(statXecStateMapped,
			statCcsPioCmdOk,
			dg.WordT(dskpExtractPioCommand(dskpData.commandRegC)),
			util.TestWbit(dskpData.commandRegC, 15))

	case 'P':
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... P flag set\n")
		}
		log.Fatalln("P flag not yet implemented in DSKP")

	default:
		// no/empty flag - nothing to do
	}
}

// seek to the disk position according to sector number in dskpData structure
func dskpPositionDiskImage() {
	var offset = int64(dskpData.sectorNo) * dskpBytesPerSector
	_, err := dskpData.imageFile.Seek(offset, 0)
	if err != nil {
		log.Fatalln("DSKP could not position disk image")
	}
	// TODO Set C/H/S???
}

// CB processing in a goroutine
func dskpCBprocessor(dataPtr *dskpDataT) {
	var (
		cb            [dskpCbMaxSize]dg.WordT
		w, cbLength   int
		nextCB        dg.PhysAddrT
		sect          dg.DwordT
		physTransfers bool
		physAddr      dg.PhysAddrT
		readBuff      = make([]byte, dskpBytesPerSector)
		writeBuff     = make([]byte, dskpBytesPerSector)
		tmpWd         dg.WordT
	)
	for {
		cbAddr := <-cbChan
		cbLength = dskpCbMinSize + dskpGetCBextendedStatusSize()
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... Processing CB, extended status size is: %d\n", dskpGetCBextendedStatusSize())
		}
		// copy CB contents from host memory
		addr := cbAddr
		for w = 0; w < cbLength; w++ {
			cb[w], addr = memory.ReadWordBmcChan(addr)
			if debugLogging {
				logging.DebugPrint(logging.DskpLog, "... CB[%d]: %d\n", w, cb[w])
			}
		}

		opCode := cb[dskpCbINA_FLAGS_OPCODE] & 0x03ff
		nextCB = dg.PhysAddrT(util.DWordFromTwoWords(cb[dskpCbLINK_ADDR_HIGH], cb[dskpCbLINK_ADDR_LOW]))
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "... CB OpCode: %d\n", opCode)
			logging.DebugPrint(logging.DskpLog, "... .. Next CB Addr: %d\n", nextCB)
		}
		switch opCode {

		case dskpCbOpRecalibrateDisk:
			dataPtr.dskpDataMu.Lock()
			if debugLogging {
				logging.DebugPrint(logging.DskpLog, "... .. RECALIBRATE\n")
			}
			//dataPtr.cylinder = 0
			//dataPtr.head = 0
			//dataPtr.sector = 0
			dataPtr.sectorNo = 0
			dskpPositionDiskImage()
			dataPtr.dskpDataMu.Unlock()
			if cbLength >= dskpCbERR_STATUS+1 {
				cb[dskpCbERR_STATUS] = 0
			}
			if cbLength >= dskpCbUNIT_STATUS+1 {
				cb[dskpCbUNIT_STATUS] = 1 << 13 // b0010000000000000; // Ready
			}
			if cbLength >= dskpCbCB_STATUS+1 {
				cb[dskpCbCB_STATUS] = 1 // finally, set Done bit
			}

		case dskpCbOpRead:
			dataPtr.dskpDataMu.Lock()
			dataPtr.sectorNo = util.DWordFromTwoWords(cb[dskpCbDEV_ADDR_HIGH], cb[dskpCbDEV_ADDR_LOW])
			if util.TestWbit(cb[dskpCbPAGENO_LIST_ADDR_HIGH], 0) {
				// logical premapped host address
				physTransfers = false
				log.Fatal("DSKP - CB READ from premapped logical addresses  Not Yet Implemented")
			} else {
				physTransfers = true
				physAddr = dg.PhysAddrT(util.DWordFromTwoWords(cb[dskpCbTXFER_ADDR_HIGH], cb[dskpCbTXFER_ADDR_LOW]))
			}
			if debugLogging {
				logging.DebugPrint(logging.DskpLog, "... .. CB READ command, SECCNT: %d\n", cb[dskpCbTXFER_COUNT])
				logging.DebugPrint(logging.DskpLog, "... .. .. .... from sector:     %d\n", dataPtr.sectorNo)
				logging.DebugPrint(logging.DskpLog, "... .. .. .... from phys addr:  %d\n", physAddr)
				logging.DebugPrint(logging.DskpLog, "... .. .. .... physical txfer?: %d\n", util.BoolToInt(physTransfers))
			}
			for sect = 0; sect < dg.DwordT(cb[dskpCbTXFER_COUNT]); sect++ {
				dataPtr.sectorNo += sect
				dskpPositionDiskImage()
				dataPtr.imageFile.Read(readBuff)
				addr = physAddr + (dg.PhysAddrT(sect) * dskpWordsPerSector)
				for w = 0; w < dskpWordsPerSector; w++ {
					tmpWd = (dg.WordT(readBuff[w*2]) << 8) | dg.WordT(readBuff[(w*2)+1])
					memory.WriteWordBmcChan(&addr, tmpWd)
				}
				dataPtr.reads++
			}
			if cbLength >= dskpCbERR_STATUS+1 {
				cb[dskpCbERR_STATUS] = 0
			}
			if cbLength >= dskpCbUNIT_STATUS+1 {
				cb[dskpCbUNIT_STATUS] = 1 << 13 // b0010000000000000; // Ready
			}
			if cbLength >= dskpCbCB_STATUS+1 {
				cb[dskpCbCB_STATUS] = 1 // finally, set Done bit
			}

			if debugLogging {
				logging.DebugPrint(logging.DskpLog, "... .. .... READ command finished\n")
				logging.DebugPrint(logging.DskpLog, "Last buffer: %X\n", readBuff)
			}
			dataPtr.dskpDataMu.Unlock()

		case dskpCbOpWrite:
			dataPtr.dskpDataMu.Lock()
			dataPtr.sectorNo = util.DWordFromTwoWords(cb[dskpCbDEV_ADDR_HIGH], cb[dskpCbDEV_ADDR_LOW])
			if util.TestWbit(cb[dskpCbPAGENO_LIST_ADDR_HIGH], 0) {
				// logical premapped host address
				physTransfers = false
				log.Fatal("DSKP - CB WRITE from premapped logical addresses  Not Yet Implemented")
			} else {
				physTransfers = true
				physAddr = dg.PhysAddrT(util.DWordFromTwoWords(cb[dskpCbTXFER_ADDR_HIGH], cb[dskpCbTXFER_ADDR_LOW]))
			}
			if debugLogging {
				logging.DebugPrint(logging.DskpLog, "... .. CB WRITE command, SECCNT: %d\n", cb[dskpCbTXFER_COUNT])
				logging.DebugPrint(logging.DskpLog, "... .. .. ..... to sector:       %d\n", dataPtr.sectorNo)
				logging.DebugPrint(logging.DskpLog, "... .. .. ..... from phys addr:  %d\n", physAddr)
				logging.DebugPrint(logging.DskpLog, "... .. .. ..... physical txfer?: %d\n", util.BoolToInt(physTransfers))
			}
			for sect = 0; sect < dg.DwordT(cb[dskpCbTXFER_COUNT]); sect++ {
				dataPtr.sectorNo += sect
				dskpPositionDiskImage()
				memAddr := physAddr + (dg.PhysAddrT(sect) * dskpWordsPerSector)
				for w = 0; w < dskpWordsPerSector; w++ {
					tmpWd, memAddr = memory.ReadWordBmcChan(memAddr)
					writeBuff[w*2] = byte(tmpWd >> 8)
					writeBuff[(w*2)+1] = byte(tmpWd & 0x00ff)
				}
				dataPtr.imageFile.Write(writeBuff)
				if debugLogging {
					logging.DebugPrint(logging.DskpLog, "Wrote buffer: %X\n", writeBuff)
				}
				dataPtr.writes++
			}
			if cbLength >= dskpCbERR_STATUS+1 {
				cb[dskpCbERR_STATUS] = 0
			}
			if cbLength >= dskpCbUNIT_STATUS+1 {
				cb[dskpCbUNIT_STATUS] = 1 << 13 // b0010000000000000; // Ready
			}
			if cbLength >= dskpCbCB_STATUS+1 {
				cb[dskpCbCB_STATUS] = 1 // finally, set Done bit
			}
			dataPtr.dskpDataMu.Unlock()

		default:
			log.Fatalf("DSKP CB Command %d not yet implemented\n", opCode)
		}

		// write back CB
		addr = cbAddr
		for w = 0; w < cbLength; w++ {
			memory.WriteWordBmcChan(&addr, cb[w])
		}

		if nextCB == 0 {
			// send ASYNCH status. See p.4-15
			if debugLogging {
				logging.DebugPrint(logging.DskpLog, "...ready to set ASYNC status\n")
			}
			for busGetBusy(devDSKP) || busGetDone(devDSKP) {
				time.Sleep(dskpAsynchStatRetryInterval)
			}
			dataPtr.dskpDataMu.Lock()
			dataPtr.statusRegC = dg.WordT(statXecStateMapped) << 12
			dataPtr.statusRegC |= (statAsyncNoErrors & 0x03ff)
			if debugLogging {
				logging.DebugPrint(logging.DskpLog, "DSKP ASYNCHRONOUS status C set to: %s\n",
					util.WordToBinStr(dataPtr.statusRegC))
			}
			dataPtr.dskpDataMu.Unlock()
			if debugLogging {
				logging.DebugPrint(logging.DskpLog, "...set ASYNC status\n")
			}
			busSetDone(devDSKP, true)
		} else {
			// chain to next CB
			//dskpProcessCB(nextCB)
			cbChan <- nextCB
		}
	}
}

func dskpReset() {
	dskpData.dskpDataMu.Lock()
	dskpResetMapping()
	dskpResetIntInfBlk()
	dskpResetCtrlrInfBlock()
	dskpResetUnitInfBlock()
	dskpData.statusRegB = 0
	dskpSetPioStatusRegC(statXecStateResetDone, 0, dskpPioReset, util.TestWbit(dskpData.commandRegC, 15))
	dskpData.dskpDataMu.Unlock()
	if debugLogging {
		logging.DebugPrint(logging.DskpLog, "DSKP ***Reset*** via call to dskpReset()\n")
	}

}

// N.B. We assume dskpData is LOCKED before calling ANY of the following functions

// setup the controller information block to power-up defaults p.2-15
func dskpResetCtrlrInfBlock() {
	dskpData.ctrlInfBlock[0] = 0
	dskpData.ctrlInfBlock[1] = 0
}

// setup the interface information block to power-up defaults
func dskpResetIntInfBlk() {
	dskpData.intInfBlock[0] = 0101
	dskpData.intInfBlock[1] = dskpUcodeRev
	dskpData.intInfBlock[2] = 3
	dskpData.intInfBlock[3] = 8<<11 | dskpMaxQueuedCBs
	dskpData.intInfBlock[4] = 0
	dskpData.intInfBlock[5] = 11 << 8
	dskpData.intInfBlock[6] = 0
	dskpData.intInfBlock[7] = 0
}

// set mapping options after IORST, power-up or Reset
func dskpResetMapping() {
	dskpData.mappingRegA = 0x4000 // DMA over the BMC
	dskpData.mappingRegB = dskpMapIntBmcPhys | dskpMapUpstreamLoad | dskpMapUpstreamHpt
	dskpData.isMapped = false
}

// setup the unit information block to power-up defaults pp.2-16
func dskpResetUnitInfBlock() {
	dskpData.unitInfBlock[0] = 0
	dskpData.unitInfBlock[1] = 9<<12 | dskpUcodeRev
	dskpData.unitInfBlock[2] = dg.WordT(dskpLogicalBlocksH) // 17.
	dskpData.unitInfBlock[3] = dg.WordT(dskpLogicalBlocksL) // 43840.
	dskpData.unitInfBlock[4] = dskpBytesPerSector
	dskpData.unitInfBlock[5] = dskpUserCylinders
	dskpData.unitInfBlock[6] = ((dskpSurfacesPerDisk * dskpHeadsPerSurface) << 8) | (0x00ff & dskpSectorsPerTrack)
}

// this is used to set the SYNCHRONOUS standard return as per p.3-22
func dskpSetPioStatusRegC(stat byte, ccs byte, cmdEcho dg.WordT, rr bool) {
	if stat == 0 && dskpData.isMapped {
		stat = statXecStateMapped
	}
	if rr || cmdEcho == dskpPioReset {
		dskpData.statusRegC = dg.WordT(stat) << 12
		dskpData.statusRegC |= (dg.WordT(ccs) & 3) << 10
		dskpData.statusRegC |= (cmdEcho & 0x01ff) << 1
		if rr {
			dskpData.statusRegC |= 1
		}
		if debugLogging {
			logging.DebugPrint(logging.DskpLog, "DSKP PIO (SYNCH) status C set to: %s\n",
				util.WordToBinStr(dskpData.statusRegC))
		}
	}
}
