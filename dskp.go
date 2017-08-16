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

	DSKP_INT_INF_BLK_SIZE   = 8
	DSKP_CTRLR_INF_BLK_SIZE = 2
	DSKP_UNIT_INF_BLK_SIZE  = 7
	DSKP_CB_MAX_SIZE        = 21
	DSKP_CB_MIN_SIZE        = 10 //12 // Was 10

	STAT_XEC_STATE_RESETTING  = 0x00
	STAT_XEC_STATE_RESET_DONE = 0x01
	STAT_XEC_STATE_BEGUN      = 0x08
	STAT_XEC_STATE_MAPPED     = 0x0c
	STAT_XEC_STATE_DIAG_MODE  = 0x04

	STAT_CCS_ASYNC          = 0
	STAT_CCS_PIO_INV_CMD    = 1
	STAT_CCS_PIO_CMD_FAILED = 2
	STAT_CCS_PIO_CMD_OK     = 3

	STAT_ASYNC_NO_ERRORS = 5

	// DSKP PIO Command Set
	DSKP_PROG_LOAD             = 000
	DSKP_BEGIN                 = 002
	DSKP_SYSGEN                = 025
	DSKP_DIAG_MODE             = 024
	DSKP_SET_MAPPING           = 026
	DSKP_GET_MAPPING           = 027
	DSKP_SET_INTERFACE         = 030
	DSKP_GET_INTERFACE         = 031
	DSKP_SET_CONTROLLER        = 032
	DSKP_GET_CONTROLLER        = 033
	DSKP_SET_UNIT              = 034
	DSKP_GET_UNIT              = 035
	DSKP_GET_EXTENDED_STATUS_0 = 040
	DSKP_GET_EXTENDED_STATUS_1 = 041
	DSKP_GET_EXTENDED_STATUS_2 = 042
	DSKP_GET_EXTENDED_STATUS_3 = 043
	DSKP_START_LIST            = 0100
	DSKP_START_LIST_HP         = 0103
	DSKP_RESTART               = 0116
	DSKP_CANCEL_LIST           = 0123
	DSKP_UNIT_STATUS           = 0131
	DSKP_TRESPASS              = 0132
	DSKP_GET_LIST_STATUS       = 0133
	DSKP_RESET                 = 0777

	// DSKP CB Command Set/OpCodes
	DSKP_CB_OP_NO_OP               = 0
	DSKP_CB_OP_WRITE               = 0100
	DSKP_CB_OP_WRITE_VERIFY        = 0101
	DSKP_CB_OP_WRITE_1_WORD        = 0104
	DSKP_CB_OP_WRITE_VERIFY_1_WORD = 0105
	DSKP_CB_OP_WRITE_MOD_BITMAP    = 0142
	DSKP_CB_OP_READ                = 0200
	DSKP_CB_OP_READ_VERIFY         = 0201
	DSKP_CB_OP_READ_VERIFY_1_WORD  = 0205
	DSKP_CB_OP_READ_RAW_DATA       = 0210
	DSKP_CB_OP_READ_HEADERS        = 0220
	DSKP_CB_OP_READ_MOD_BITMAP     = 0242
	DSKP_CB_OP_RECALIBRATE_DISK    = 0400

	// DSKP CB FIELDS
	DSKP_CB_LINK_ADDR_HIGH        = 0
	DSKP_CB_LINK_ADDR_LOW         = 1
	DSKP_CB_INA_FLAGS_OPCODE      = 2
	DSKP_CB_PAGENO_LIST_ADDR_HIGH = 3
	DSKP_CB_PAGENO_LIST_ADDR_LOW  = 4
	DSKP_CB_TXFER_ADDR_HIGH       = 5
	DSKP_CB_TXFER_ADDR_LOW        = 6
	DSKP_CB_DEV_ADDR_HIGH         = 7
	DSKP_CB_DEV_ADDR_LOW          = 8
	DSKP_CB_UNIT_NO               = 9
	DSKP_CB_TXFER_COUNT           = 10
	DSKP_CB_CB_STATUS             = 11
	DSKP_CB_RES1                  = 12
	DSKP_CB_RES2                  = 13
	DSKP_CB_ERR_STATUS            = 14
	DSKP_CB_UNIT_STATUS           = 15
	DSKP_CB_RETRIES_DONE          = 16
	DSKP_CB_SOFT_RTN_TXFER_COUNT  = 17
	DSKP_CB_PHYS_CYL              = 18
	DSKP_CB_PHYS_HEAD_SECT        = 19
	DSKP_CB_DISK_ERR_CODE         = 20

	// Mapping bits
	DSKP_MAP_SLOT_LOAD_INTS = 1 << 15
	DSKP_MAP_INT_BMC_PHYS   = 1 << 14
	DSKP_MAP_UPSTREAM_LOAD  = 1 << 13
	DSKP_MAP_UPSTREAM_HPT   = 1 << 12

	// calculated consts
	// dskpPhysicalByteSize is the total  # bytes on a DSKP-type disk
	dskpPhysicalByteSize = dskpSurfacesPerDisk * dskpHeadsPerSurface * dskpSectorsPerTrack * dskpBytesPerSector * dskpPhysicalCylinders
	// dskpPhysicalBlockSize is the total # blocks on a DSKP-type disk
	dskpPhysicalBlockSize = dskpSurfacesPerDisk * dskpHeadsPerSurface * dskpSectorsPerTrack * dskpPhysicalCylinders
)

type dskpDataT struct {
	// MV/Em internals...
	debug         bool
	dskpDataMu    sync.RWMutex
	imageAttached bool
	imageFileName string
	imageFile     *os.File
	reads, writes uint64
	// DG data...
	commandRegA, commandRegB, commandRegC dg.WordT
	statusRegA, statusRegB, statusRegC    dg.WordT
	mappingRegA, mappingRegB              dg.WordT
	intInfBlock                           [DSKP_INT_INF_BLK_SIZE]dg.WordT
	ctrlInfBlock                          [DSKP_CTRLR_INF_BLK_SIZE]dg.WordT
	unitInfBlock                          [DSKP_UNIT_INF_BLK_SIZE]dg.WordT
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
)

// dskpInit is called once by the main routine to initialise this DSKP emulator
func dskpInit(statsChann chan dskpStatT) {
	logging.DebugPrint(logging.DskpLog, "DSKP Initialising via call to dskpInit()...\n")
	dskpData.dskpDataMu.Lock()

	dskpData.debug = true

	go dskpStatSender(statsChann)

	busAddDevice(DEV_DSKP, "DSKP", DSKP_PMB, false, true, true)
	busSetResetFunc(DEV_DSKP, dskpReset)
	busSetDataInFunc(DEV_DSKP, dskpDataIn)
	busSetDataOutFunc(DEV_DSKP, dskpDataOut)

	dskpData.imageAttached = false
	dskpData.dskpDataMu.Unlock()

	dskpReset()
}

// attempt to attach an extant MV/Em disk image to the running emulator
func dskpAttach(dNum int, imgName string) bool {
	// TODO Disk Number not currently used
	logging.DebugPrint(logging.DskpLog, "dskpAttach called for disk #%d with image <%s>\n", dNum, imgName)
	dskpData.dskpDataMu.Lock()
	defer dskpData.dskpDataMu.Unlock()
	dskpData.imageFile, err = os.OpenFile(imgName, os.O_RDWR, 0755)
	if err != nil {
		logging.DebugPrint(logging.DskpLog, "Failed to open image for attaching\n")
		logging.DebugPrint(logging.DebugLog, "WARN: Failed to open dskp image <%s> for ATTach\n", imgName)
		return false
	}
	dskpData.imageFileName = imgName
	dskpData.imageAttached = true
	busSetAttached(DEV_DSKP)
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
func dskpDataIn(cpuPtr *CPU, iPtr *decodedInstrT, abc byte) {
	dskpData.dskpDataMu.Lock()
	switch abc {
	case 'A':
		cpuPtr.ac[iPtr.acd] = dg.DwordT(dskpData.statusRegA)
		logging.DebugPrint(logging.DskpLog, "DIA [Read Status A] returning %s for DRV=%d, PC: %d\n", util.WordToBinStr(dskpData.statusRegA), 0, cpuPtr.pc)
	case 'B':
		cpuPtr.ac[iPtr.acd] = dg.DwordT(dskpData.statusRegB)
		logging.DebugPrint(logging.DskpLog, "DIB [Read Status B] returning %s for DRV=%d, PC: %d\n", util.WordToBinStr(dskpData.statusRegB), 0, cpuPtr.pc)
	case 'C':
		cpuPtr.ac[iPtr.acd] = dg.DwordT(dskpData.statusRegC)
		logging.DebugPrint(logging.DskpLog, "DIC [Read Status C] returning %s for DRV=%d, PC: %d\n", util.WordToBinStr(dskpData.statusRegC), 0, cpuPtr.pc)
	}
	dskpData.dskpDataMu.Unlock()
	dskpHandleFlag(iPtr.f)
}

// Handle the DOA/B/C PIO commands
func dskpDataOut(cpuPtr *CPU, iPtr *decodedInstrT, abc byte) {
	dskpData.dskpDataMu.Lock()
	switch abc {
	case 'A':
		dskpData.commandRegA = util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "DOA [Load Cmd Reg A] from AC%d containing %s, PC: %d\n",
				iPtr.acd, util.WordToBinStr(dskpData.commandRegA), cpuPtr.pc)
		}
	case 'B':
		dskpData.commandRegB = util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "DOB [Load Cmd Reg B] from AC%d containing %s, PC: %d\n",
				iPtr.acd, util.WordToBinStr(dskpData.commandRegB), cpuPtr.pc)
		}
	case 'C':
		dskpData.commandRegC = util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "DOC [Load Cmd Reg C] from AC%d containing %s, PC: %d\n",
				iPtr.acd, util.WordToBinStr(dskpData.commandRegC), cpuPtr.pc)
		}
	}
	dskpData.dskpDataMu.Unlock()
	dskpHandleFlag(iPtr.f)
}

func dskpDoPioCommand() {

	var addr, w dg.PhysAddrT

	pioCmd := dskpExtractPioCommand(dskpData.commandRegC)
	switch pioCmd {
	case DSKP_PROG_LOAD:
		log.Panicln("DSKP_PROG_LOAD command not yet implemented")

	case DSKP_BEGIN:
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... BEGIN command, unit # %d\n", dskpData.commandRegA)
		}
		// pretend we have succesfully booted ourself
		dskpData.statusRegB = 0
		dskpSetPioStatusRegC(STAT_XEC_STATE_BEGUN, STAT_CCS_PIO_CMD_OK, DSKP_BEGIN, util.TestWbit(dskpData.commandRegC, 15))
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... ..... returning %s\n", util.WordToBinStr(dskpData.statusRegC))
		}

	case DSKP_GET_MAPPING:
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... GET MAPPING command\n")
		}
		dskpData.statusRegA = dskpData.mappingRegA
		dskpData.statusRegB = dskpData.mappingRegB
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_GET_MAPPING, util.TestWbit(dskpData.commandRegC, 15))
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... ... Status Reg A set to %s\n", util.WordToBinStr(dskpData.statusRegA))
			logging.DebugPrint(logging.DskpLog, "... ... Status Reg B set to %s\n", util.WordToBinStr(dskpData.statusRegB))
		}
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_GET_MAPPING, util.TestWbit(dskpData.commandRegC, 15))

	case DSKP_SET_MAPPING:
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... SET MAPPING command\n")
		}
		dskpData.mappingRegA = dskpData.commandRegA
		dskpData.mappingRegB = dskpData.commandRegB
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... ... Mapping Reg A set to %s\n", util.WordToBinStr(dskpData.commandRegA))
			logging.DebugPrint(logging.DskpLog, "... ... Mapping Reg B set to %s\n", util.WordToBinStr(dskpData.commandRegB))
		}
		dskpSetPioStatusRegC(STAT_XEC_STATE_MAPPED, STAT_CCS_PIO_CMD_OK, DSKP_SET_MAPPING, util.TestWbit(dskpData.commandRegC, 15))

	case DSKP_GET_INTERFACE:
		addr = dg.PhysAddrT(util.DWordFromTwoWords(dskpData.commandRegA, dskpData.commandRegB))
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... GET INTERFACE INFO command\n")
			logging.DebugPrint(logging.DskpLog, "... ... Destination Start Address: %d\n", addr)
		}
		for w = 0; w < DSKP_INT_INF_BLK_SIZE; w++ {
			memory.MemWriteWordBmcChan(addr+w, dskpData.intInfBlock[w])
			if dskpData.debug {
				logging.DebugPrint(logging.DskpLog, "... ... Word %d: %s\n", w, util.WordToBinStr(dskpData.intInfBlock[w]))
			}
		}
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_GET_INTERFACE, util.TestWbit(dskpData.commandRegC, 15))

	case DSKP_SET_INTERFACE:
		addr = dg.PhysAddrT(util.DWordFromTwoWords(dskpData.commandRegA, dskpData.commandRegB))
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... SET INTERFACE INFO command\n")
			logging.DebugPrint(logging.DskpLog, "... ... Origin Start Address: %d\n", addr)
		}
		// only a few fields can be changed...
		w = 5
		dskpData.intInfBlock[w] = memory.MemReadWordBmcChan(addr+w) & 0xff00
		w = 6
		dskpData.intInfBlock[w] = memory.MemReadWordBmcChan(addr + w)
		w = 7
		dskpData.intInfBlock[w] = memory.MemReadWordBmcChan(addr + w)
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... ... Word 5: %s\n", util.WordToBinStr(dskpData.intInfBlock[5]))
			logging.DebugPrint(logging.DskpLog, "... ... Word 6: %s\n", util.WordToBinStr(dskpData.intInfBlock[6]))
			logging.DebugPrint(logging.DskpLog, "... ... Word 7: %s\n", util.WordToBinStr(dskpData.intInfBlock[7]))
		}
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_SET_INTERFACE, util.TestWbit(dskpData.commandRegC, 15))

	case DSKP_GET_UNIT:
		addr = dg.PhysAddrT(util.DWordFromTwoWords(dskpData.commandRegA, dskpData.commandRegB))
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... GET UNIT INFO command\n")
			logging.DebugPrint(logging.DskpLog, "... ... Destination Start Address: %d\n", addr)
		}
		for w = 0; w < DSKP_UNIT_INF_BLK_SIZE; w++ {
			memory.MemWriteWordBmcChan(addr+w, dskpData.unitInfBlock[w])
			if dskpData.debug {
				logging.DebugPrint(logging.DskpLog, "... ... Word %d: %s\n", w, util.WordToBinStr(dskpData.unitInfBlock[w]))
			}
		}
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_GET_UNIT, util.TestWbit(dskpData.commandRegC, 15))

	case DSKP_SET_UNIT:
		addr = dg.PhysAddrT(util.DWordFromTwoWords(dskpData.commandRegA, dskpData.commandRegB))
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... SET UNIT INFO command\n")
			logging.DebugPrint(logging.DskpLog, "... ... Origin Start Address: %d\n", addr)
		}
		// only the first word is writable according to p.2-16
		// TODO check no active CBs first
		dskpData.unitInfBlock[0] = memory.ReadWord(addr)
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_SET_UNIT, util.TestWbit(dskpData.commandRegC, 15))
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... ... Overwrote word 0 of UIB with: %s\n", util.WordToBinStr(dskpData.unitInfBlock[0]))
		}

	case DSKP_RESET:
		dskpReset()

	case DSKP_SET_CONTROLLER:
		addr = dg.PhysAddrT(util.DWordFromTwoWords(dskpData.commandRegA, dskpData.commandRegB))
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... SET CONTROLLER INFO command\n")
			logging.DebugPrint(logging.DskpLog, "... ... Origin Start Address: %d\n", addr)
		}
		dskpData.ctrlInfBlock[0] = memory.ReadWord(addr)
		dskpData.ctrlInfBlock[1] = memory.ReadWord(addr + 1)
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_SET_CONTROLLER, util.TestWbit(dskpData.commandRegC, 15))
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... ... Word 0: %s\n", util.WordToBinStr(dskpData.ctrlInfBlock[0]))
			logging.DebugPrint(logging.DskpLog, "... ... Word 1: %s\n", util.WordToBinStr(dskpData.ctrlInfBlock[1]))
		}

	case DSKP_START_LIST:
		addr = dg.PhysAddrT(util.DWordFromTwoWords(dskpData.commandRegA, dskpData.commandRegB))
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... START LIST command\n")
			logging.DebugPrint(logging.DskpLog, "... ..... First CB Address: %d\n", addr)
		}
		// TODO should check addr validity before starting processing
		dskpProcessCB(addr)
		dskpData.statusRegA = util.DWordGetUpperWord(dg.DwordT(addr)) // return address of 1st CB processed
		dskpData.statusRegB = util.DWordGetLowerWord(dg.DwordT(addr))
		//dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_START_LIST,  util.TestWbit(dskpData.commandRegC, 15))

	default:
		log.Panicf("DSKP command %d not yet implemented\n", pioCmd)
	}
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
		busSetBusy(DEV_DSKP, true)
		busSetDone(DEV_DSKP, false)
		dskpData.dskpDataMu.Lock()
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... S flag set\n")
		}
		dskpDoPioCommand()
		dskpData.dskpDataMu.Unlock()
		busSetBusy(DEV_DSKP, false)
		busSetDone(DEV_DSKP, true)

	case 'C':
		dskpData.dskpDataMu.Lock()
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... C flag set, clearing DONE flag\n")
		}
		busSetDone(DEV_DSKP, false)
		// TODO clear pending interrupt
		dskpSetPioStatusRegC(STAT_XEC_STATE_MAPPED,
			STAT_CCS_PIO_CMD_OK,
			dg.WordT(dskpExtractPioCommand(dskpData.commandRegC)),
			util.TestWbit(dskpData.commandRegC, 15))
		dskpData.dskpDataMu.Unlock()
	case 'P':
		if dskpData.debug {
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

func dskpProcessCB(addr dg.PhysAddrT) {
	var (
		cb            [DSKP_CB_MAX_SIZE]dg.WordT
		w, cbLength   int
		nextCB        dg.PhysAddrT
		sect          dg.DwordT
		physTransfers bool
		physAddr      dg.PhysAddrT
		readBuff      = make([]byte, dskpBytesPerSector)
		writeBuff     = make([]byte, dskpBytesPerSector)
		tmpWd         dg.WordT
	)
	cbLength = DSKP_CB_MIN_SIZE + dskpGetCBextendedStatusSize()
	if dskpData.debug {
		logging.DebugPrint(logging.DskpLog, "... Processing CB, extended status size is: %d\n", dskpGetCBextendedStatusSize())
	}
	// copy CB contents from host memory
	for w = 0; w < cbLength; w++ {
		cb[w] = memory.MemReadWordBmcChan(addr + dg.PhysAddrT(w))
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... CB[%d]: %d\n", w, cb[w])
		}
	}

	opCode := cb[DSKP_CB_INA_FLAGS_OPCODE] & 0x03ff
	nextCB = dg.PhysAddrT(util.DWordFromTwoWords(cb[DSKP_CB_LINK_ADDR_HIGH], cb[DSKP_CB_LINK_ADDR_LOW]))
	if dskpData.debug {
		logging.DebugPrint(logging.DskpLog, "... CB OpCode: %d\n", opCode)
		logging.DebugPrint(logging.DskpLog, "... .. Next CB Addr: %d\n", nextCB)
	}
	switch opCode {

	case DSKP_CB_OP_RECALIBRATE_DISK:
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... .. RECALIBRATE\n")
		}
		//dskpData.cylinder = 0
		//dskpData.head = 0
		//dskpData.sector = 0
		dskpData.sectorNo = 0
		dskpPositionDiskImage()
		if cbLength >= DSKP_CB_ERR_STATUS+1 {
			cb[DSKP_CB_ERR_STATUS] = 0
		}
		if cbLength >= DSKP_CB_UNIT_STATUS+1 {
			cb[DSKP_CB_UNIT_STATUS] = 1 << 13 // b0010000000000000; // Ready
		}
		if cbLength >= DSKP_CB_CB_STATUS+1 {
			cb[DSKP_CB_CB_STATUS] = 1 // finally, set Done bit
		}
		dskpSetAsyncStatusRegC(STAT_XEC_STATE_MAPPED, STAT_ASYNC_NO_ERRORS)

	case DSKP_CB_OP_READ:
		dskpData.sectorNo = util.DWordFromTwoWords(cb[DSKP_CB_DEV_ADDR_HIGH], cb[DSKP_CB_DEV_ADDR_LOW])
		if util.TestWbit(cb[DSKP_CB_PAGENO_LIST_ADDR_HIGH], 0) {
			// logical premapped host address
			physTransfers = false
			log.Fatal("DSKP - CB READ from premapped logical addresses  Not Yet Implemented")
		} else {
			physTransfers = true
			physAddr = dg.PhysAddrT(util.DWordFromTwoWords(cb[DSKP_CB_TXFER_ADDR_HIGH], cb[DSKP_CB_TXFER_ADDR_LOW]))
		}
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... .. CB READ command, SECCNT: %d\n", cb[DSKP_CB_TXFER_COUNT])
			logging.DebugPrint(logging.DskpLog, "... .. .. .... from sector:     %d\n", dskpData.sectorNo)
			logging.DebugPrint(logging.DskpLog, "... .. .. .... from phys addr:  %d\n", physAddr)
			logging.DebugPrint(logging.DskpLog, "... .. .. .... physical txfer?: %d\n", util.BoolToInt(physTransfers))
		}
		for sect = 0; sect < dg.DwordT(cb[DSKP_CB_TXFER_COUNT]); sect++ {
			dskpData.sectorNo += sect
			dskpPositionDiskImage()
			dskpData.imageFile.Read(readBuff)
			for w = 0; w < dskpWordsPerSector; w++ {
				tmpWd = (dg.WordT(readBuff[w*2]) << 8) | dg.WordT(readBuff[(w*2)+1])
				memory.MemWriteWordBmcChan(physAddr+(dg.PhysAddrT(sect)*dskpWordsPerSector)+dg.PhysAddrT(w), tmpWd)
			}
			dskpData.reads++
		}
		if cbLength >= DSKP_CB_ERR_STATUS+1 {
			cb[DSKP_CB_ERR_STATUS] = 0
		}
		if cbLength >= DSKP_CB_UNIT_STATUS+1 {
			cb[DSKP_CB_UNIT_STATUS] = 1 << 13 // b0010000000000000; // Ready
		}
		if cbLength >= DSKP_CB_CB_STATUS+1 {
			cb[DSKP_CB_CB_STATUS] = 1 // finally, set Done bit
		}

		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... .. .... READ command finished\n")
			logging.DebugPrint(logging.DskpLog, "Last buffer: %X\n", readBuff)
		}
		dskpSetAsyncStatusRegC(STAT_XEC_STATE_MAPPED, STAT_ASYNC_NO_ERRORS)
		//dskpSetAsyncStatusRegC(0, STAT_ASYNC_NO_ERRORS)

	case DSKP_CB_OP_WRITE:
		dskpData.sectorNo = util.DWordFromTwoWords(cb[DSKP_CB_DEV_ADDR_HIGH], cb[DSKP_CB_DEV_ADDR_LOW])
		if util.TestWbit(cb[DSKP_CB_PAGENO_LIST_ADDR_HIGH], 0) {
			// logical premapped host address
			physTransfers = false
			log.Fatal("DSKP - CB WRITE from premapped logical addresses  Not Yet Implemented")
		} else {
			physTransfers = true
			physAddr = dg.PhysAddrT(util.DWordFromTwoWords(cb[DSKP_CB_TXFER_ADDR_HIGH], cb[DSKP_CB_TXFER_ADDR_LOW]))
		}
		if dskpData.debug {
			logging.DebugPrint(logging.DskpLog, "... .. CB WRITE command, SECCNT: %d\n", cb[DSKP_CB_TXFER_COUNT])
			logging.DebugPrint(logging.DskpLog, "... .. .. ..... to sector:       %d\n", dskpData.sectorNo)
			logging.DebugPrint(logging.DskpLog, "... .. .. ..... from phys addr:  %d\n", physAddr)
			logging.DebugPrint(logging.DskpLog, "... .. .. ..... physical txfer?: %d\n", util.BoolToInt(physTransfers))
		}
		for sect = 0; sect < dg.DwordT(cb[DSKP_CB_TXFER_COUNT]); sect++ {
			dskpData.sectorNo += sect
			dskpPositionDiskImage()
			for w = 0; w < dskpWordsPerSector; w++ {
				tmpWd = memory.MemReadWordBmcChan(physAddr + (dg.PhysAddrT(sect) * dskpWordsPerSector) + dg.PhysAddrT(w))
				writeBuff[w*2] = byte(tmpWd >> 8)
				writeBuff[(w*2)+1] = byte(tmpWd & 0x00ff)
			}
			dskpData.imageFile.Write(writeBuff)
			if dskpData.debug {
				logging.DebugPrint(logging.DskpLog, "Wrote buffer: %X\n", writeBuff)
			}
			dskpData.writes++
		}
		if cbLength >= DSKP_CB_ERR_STATUS+1 {
			cb[DSKP_CB_ERR_STATUS] = 0
		}
		if cbLength >= DSKP_CB_UNIT_STATUS+1 {
			cb[DSKP_CB_UNIT_STATUS] = 1 << 13 // b0010000000000000; // Ready
		}
		if cbLength >= DSKP_CB_CB_STATUS+1 {
			cb[DSKP_CB_CB_STATUS] = 1 // finally, set Done bit
		}

		dskpSetAsyncStatusRegC(STAT_XEC_STATE_MAPPED, STAT_ASYNC_NO_ERRORS)
		//dskpSetAsyncStatusRegC(0, STAT_ASYNC_NO_ERRORS)

	default:
		log.Fatalf("DSKP CB Command %d not yet implemented\n", opCode)
	}
	// write back CB
	for w = 0; w < cbLength; w++ {
		memory.MemWriteWordBmcChan(addr+dg.PhysAddrT(w), cb[w])
	}
	// chain to next CB?
	if nextCB != 0 {
		dskpProcessCB(nextCB)
	}
}

func dskpReset() {
	dskpData.dskpDataMu.Lock()
	dskpResetMapping()
	dskpResetIntInfBlk()
	dskpResetCtrlrInfBlock()
	dskpResetUnitInfBlock()
	dskpData.statusRegB = 0
	dskpSetPioStatusRegC(STAT_XEC_STATE_RESET_DONE, 0, DSKP_RESET, util.TestWbit(dskpData.commandRegC, 15))
	if dskpData.debug {
		logging.DebugPrint(logging.DskpLog, "DSKP Reset via call to dskpReset()\n")
	}
	dskpData.dskpDataMu.Unlock()
}

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
	dskpData.mappingRegB = DSKP_MAP_INT_BMC_PHYS | DSKP_MAP_UPSTREAM_LOAD | DSKP_MAP_UPSTREAM_HPT
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

func dskpSetAsyncStatusRegC(stat byte, asyncIntCode dg.WordT) {
	dskpData.statusRegC = dg.WordT(stat) << 12
	dskpData.statusRegC |= (asyncIntCode & 0x03ff)
	if dskpData.debug {
		logging.DebugPrint(logging.DskpLog, "DSKP PIO ASYNCHRONOUS status C set to: %s\n",
			util.WordToBinStr(dskpData.statusRegC))
	}
}

func dskpSetPioStatusRegC(stat byte, ccs byte, cmdEcho dg.WordT, rr bool) {
	dskpData.statusRegC = dg.WordT(stat) << 12
	dskpData.statusRegC |= (dg.WordT(ccs) & 3) << 10
	dskpData.statusRegC |= (cmdEcho & 0x01ff) << 1
	if rr {
		dskpData.statusRegC |= 1
	}
	if dskpData.debug {
		logging.DebugPrint(logging.DskpLog, "DSKP PIO (SYNCH) status C set to: %s\n",
			util.WordToBinStr(dskpData.statusRegC))
	}
}
