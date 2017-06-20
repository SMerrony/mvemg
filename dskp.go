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
// controller/drive combination which provides 592MB of formatted capacity.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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
	dskpUcodeRev          = 99

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
	// DSKP_PHYSICAL_BYTE_SIZE is the total  # bytes on a DSKP-type disk
	DSKP_PHYSICAL_BYTE_SIZE = dskpSurfacesPerDisk * dskpHeadsPerSurface * dskpSectorsPerTrack * dskpBytesPerSector * dskpPhysicalCylinders
	// DSKP_PHYSICAL_BLOCK_SIZE is the total # blocks on a DSKP-type disk
	DSKP_PHYSICAL_BLOCK_SIZE = dskpSurfacesPerDisk * dskpHeadsPerSurface * dskpSectorsPerTrack * dskpPhysicalCylinders
)

type dskpDataT struct {
	// MV/Em internals...
	debug         bool
	imageAttached bool
	imageFileName string
	imageFile     *os.File
	reads, writes uint64
	// DG data...
	commandRegA, commandRegB, commandRegC dg_word
	statusRegA, statusRegB, statusRegC    dg_word
	mappingRegA, mappingRegB              dg_word
	intInfBlock                           [DSKP_INT_INF_BLK_SIZE]dg_word
	ctrlInfBlock                          [DSKP_CTRLR_INF_BLK_SIZE]dg_word
	unitInfBlock                          [DSKP_UNIT_INF_BLK_SIZE]dg_word
	cylinder, head, sector                dg_word
	sectorNo                              dg_dword
}

const dskpStatsPeriodMs = 333 // Will send status update this often

type dskpStatT struct {
	imageAttached                      bool
	statusRegA, statusRegB, statusRegC dg_word
	cylinder, head, sector             dg_word
	sectorNo                           dg_dword
	reads, writes                      uint64
}

var (
	dskpData dskpDataT
)

// dskpInit is called once by the main routine to initialise this DSKP emulator
func dskpInit(statsChann chan dskpStatT) {
	debugPrint(dskpLog, "DSKP Initialising via call to dskpInit()...\n")

	dskpData.debug = true

	go dskpStatSender(statsChann)

	busAddDevice(DEV_DSKP, "DSKP", DSKP_PMB, false, true, true)
	busSetResetFunc(DEV_DSKP, dskpReset)
	busSetDataInFunc(DEV_DSKP, dskpDataIn)
	busSetDataOutFunc(DEV_DSKP, dskpDataOut)

	dskpData.imageAttached = false
	dskpReset()
}

// attempt to attach an extant MV/Em disk image to the running emulator
func dskpAttach(dNum int, imgName string) bool {
	// TODO Disk Number not currently used
	debugPrint(dskpLog, "dskpAttach called for disk #%d with image <%s>\n", dNum, imgName)
	dskpData.imageFile, err = os.OpenFile(imgName, os.O_RDWR, 0755)
	if err != nil {
		debugPrint(dskpLog, "Failed to open image for attaching\n")
		debugPrint(debugLog, "WARN: Failed to open dskp image <%s> for ATTach\n", imgName)
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

		if dskpData.imageAttached {
			stats.imageAttached = true
			stats.cylinder = dskpData.cylinder
			stats.head = dskpData.head
			stats.sector = dskpData.sector
			stats.statusRegA = dskpData.statusRegA
			stats.statusRegB = dskpData.statusRegB
			stats.statusRegC = dskpData.statusRegC
			stats.sectorNo = dskpData.sectorNo
			stats.reads = dskpData.reads
			stats.writes = dskpData.writes
		} else {
			stats = dskpStatT{}
		}
		//fmt.Printf("dskpStatSender()  before send\n")
		select {
		case sChan <- stats:
			//fmt.Printf("dskpStatSender() after sending C: %d, H: %d S: %d #: %d\n",
			//	stats.cylinder, stats.head, stats.sector, stats.sectorNo)
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
	debugPrint(dskpLog, "dskpCreateBlank attempting to write %d bytes\n", DSKP_PHYSICAL_BYTE_SIZE)
	w := bufio.NewWriter(newFile)
	for b := 0; b < DSKP_PHYSICAL_BYTE_SIZE; b++ {
		w.WriteByte(0)
	}
	w.Flush()
	return true
}

// Handle the DIA/B/C PIO commands
func dskpDataIn(cpuPtr *Cpu, iPtr *DecodedInstr, abc byte) {
	switch abc {
	case 'A':
		cpuPtr.ac[iPtr.acd] = dg_dword(dskpData.statusRegA)
		debugPrint(dskpLog, "DIA [Read Status A] returning %s for DRV=%d, PC: %d\n", wordToBinStr(dskpData.statusRegA), 0, cpuPtr.pc)
	case 'B':
		cpuPtr.ac[iPtr.acd] = dg_dword(dskpData.statusRegB)
		debugPrint(dskpLog, "DIB [Read Status B] returning %s for DRV=%d, PC: %d\n", wordToBinStr(dskpData.statusRegB), 0, cpuPtr.pc)
	case 'C':
		cpuPtr.ac[iPtr.acd] = dg_dword(dskpData.statusRegC)
		debugPrint(dskpLog, "DIC [Read Status C] returning %s for DRV=%d, PC: %d\n", wordToBinStr(dskpData.statusRegC), 0, cpuPtr.pc)
	}
	dskpHandleFlag(iPtr.f)
}

// Handle the DOA/B/C PIO commands
func dskpDataOut(cpuPtr *Cpu, iPtr *DecodedInstr, abc byte) {
	switch abc {
	case 'A':
		dskpData.commandRegA = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if dskpData.debug {
			debugPrint(dskpLog, "DOA [Load Cmd Reg A] from AC%d containing %s, PC: %d\n",
				iPtr.acd, wordToBinStr(dskpData.commandRegA), cpuPtr.pc)
		}
	case 'B':
		dskpData.commandRegB = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if dskpData.debug {
			debugPrint(dskpLog, "DOB [Load Cmd Reg B] from AC%d containing %s, PC: %d\n",
				iPtr.acd, wordToBinStr(dskpData.commandRegB), cpuPtr.pc)
		}
	case 'C':
		dskpData.commandRegC = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if dskpData.debug {
			debugPrint(dskpLog, "DOC [Load Cmd Reg C] from AC%d containing %s, PC: %d\n",
				iPtr.acd, wordToBinStr(dskpData.commandRegC), cpuPtr.pc)
		}
	}
	dskpHandleFlag(iPtr.f)
}

func dskpDoPioCommand() {

	var addr, w dg_phys_addr

	pioCmd := dskpExtractPioCommand(dskpData.commandRegC)
	switch pioCmd {
	case DSKP_PROG_LOAD:
		log.Panicln("DSKP_PROG_LOAD command not yet implemented")

	case DSKP_BEGIN:
		if dskpData.debug {
			debugPrint(dskpLog, "... BEGIN command, unit # %d\n", dskpData.commandRegA)
		}
		// pretend we have succesfully booted ourself
		dskpData.statusRegB = 0
		dskpSetPioStatusRegC(STAT_XEC_STATE_BEGUN, STAT_CCS_PIO_CMD_OK, DSKP_BEGIN, testWbit(dskpData.commandRegC, 15))
		if dskpData.debug {
			debugPrint(dskpLog, "... ..... returning %s\n", wordToBinStr(dskpData.statusRegC))
		}

	case DSKP_GET_MAPPING:
		if dskpData.debug {
			debugPrint(dskpLog, "... GET MAPPING command\n")
		}
		dskpData.statusRegA = dskpData.mappingRegA
		dskpData.statusRegB = dskpData.mappingRegB
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_GET_MAPPING, testWbit(dskpData.commandRegC, 15))
		if dskpData.debug {
			debugPrint(dskpLog, "... ... Status Reg A set to %s\n", wordToBinStr(dskpData.statusRegA))
			debugPrint(dskpLog, "... ... Status Reg B set to %s\n", wordToBinStr(dskpData.statusRegB))
		}
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_GET_MAPPING, testWbit(dskpData.commandRegC, 15))

	case DSKP_SET_MAPPING:
		if dskpData.debug {
			debugPrint(dskpLog, "... SET MAPPING command\n")
		}
		dskpData.mappingRegA = dskpData.commandRegA
		dskpData.mappingRegB = dskpData.commandRegB
		if dskpData.debug {
			debugPrint(dskpLog, "... ... Mapping Reg A set to %s\n", wordToBinStr(dskpData.commandRegA))
			debugPrint(dskpLog, "... ... Mapping Reg B set to %s\n", wordToBinStr(dskpData.commandRegB))
		}
		dskpSetPioStatusRegC(STAT_XEC_STATE_MAPPED, STAT_CCS_PIO_CMD_OK, DSKP_SET_MAPPING, testWbit(dskpData.commandRegC, 15))

	case DSKP_GET_INTERFACE:
		addr = dg_phys_addr(dskpData.commandRegA)<<16 | dg_phys_addr(dskpData.commandRegB)
		if dskpData.debug {
			debugPrint(dskpLog, "... GET INTERFACE INFO command\n")
			debugPrint(dskpLog, "... ... Destination Start Address: %d\n", addr)
		}
		for w = 0; w < DSKP_INT_INF_BLK_SIZE; w++ {
			memWriteWord(addr+w, dskpData.intInfBlock[w])
			if dskpData.debug {
				debugPrint(dskpLog, "... ... Word %d: %s\n", w, wordToBinStr(dskpData.intInfBlock[w]))
			}
		}
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_GET_INTERFACE, testWbit(dskpData.commandRegC, 15))

	case DSKP_SET_INTERFACE:
		addr = dg_phys_addr(dskpData.commandRegA)<<16 | dg_phys_addr(dskpData.commandRegB)
		if dskpData.debug {
			debugPrint(dskpLog, "... SET INTERFACE INFO command\n")
			debugPrint(dskpLog, "... ... Origin Start Address: %d\n", addr)
		}
		// only a few fields can be changed...
		w = 5
		dskpData.intInfBlock[w] = memReadWord(addr+w) & 0xff00
		w = 6
		dskpData.intInfBlock[w] = memReadWord(addr + w)
		w = 7
		dskpData.intInfBlock[w] = memReadWord(addr + w)
		if dskpData.debug {
			debugPrint(dskpLog, "... ... Word 5: %s\n", wordToBinStr(dskpData.intInfBlock[5]))
			debugPrint(dskpLog, "... ... Word 6: %s\n", wordToBinStr(dskpData.intInfBlock[6]))
			debugPrint(dskpLog, "... ... Word 7: %s\n", wordToBinStr(dskpData.intInfBlock[7]))
		}
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_SET_INTERFACE, testWbit(dskpData.commandRegC, 15))

	case DSKP_GET_UNIT:
		addr = dg_phys_addr(dskpData.commandRegA)<<16 | dg_phys_addr(dskpData.commandRegB)
		if dskpData.debug {
			debugPrint(dskpLog, "... GET UNIT INFO command\n")
			debugPrint(dskpLog, "... ... Destination Start Address: %d\n", addr)
		}
		for w = 0; w < DSKP_UNIT_INF_BLK_SIZE; w++ {
			memWriteWord(addr+w, dskpData.unitInfBlock[w])
			if dskpData.debug {
				debugPrint(dskpLog, "... ... Word %d: %s\n", w, wordToBinStr(dskpData.unitInfBlock[w]))
			}
		}
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_GET_UNIT, testWbit(dskpData.commandRegC, 15))

	case DSKP_SET_UNIT:
		addr = dg_phys_addr(dskpData.commandRegA)<<16 | dg_phys_addr(dskpData.commandRegB)
		if dskpData.debug {
			debugPrint(dskpLog, "... SET UNIT INFO command\n")
			debugPrint(dskpLog, "... ... Origin Start Address: %d\n", addr)
		}
		// only the first word is writable according to p.2-16
		// TODO check no active CBs first
		dskpData.unitInfBlock[0] = memReadWord(addr)
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_SET_UNIT, testWbit(dskpData.commandRegC, 15))
		if dskpData.debug {
			debugPrint(dskpLog, "... ... Overwrote word 0 of UIB with: %s\n", wordToBinStr(dskpData.unitInfBlock[0]))
		}

	case DSKP_RESET:
		dskpReset()

	case DSKP_SET_CONTROLLER:
		addr = dg_phys_addr(dskpData.commandRegA)<<16 | dg_phys_addr(dskpData.commandRegB)
		if dskpData.debug {
			debugPrint(dskpLog, "... SET CONTROLLER INFO command\n")
			debugPrint(dskpLog, "... ... Origin Start Address: %d\n", addr)
		}
		dskpData.ctrlInfBlock[0] = memReadWord(addr)
		dskpData.ctrlInfBlock[1] = memReadWord(addr + 1)
		dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_SET_CONTROLLER, testWbit(dskpData.commandRegC, 15))
		if dskpData.debug {
			debugPrint(dskpLog, "... ... Word 0: %s\n", wordToBinStr(dskpData.ctrlInfBlock[0]))
			debugPrint(dskpLog, "... ... Word 1: %s\n", wordToBinStr(dskpData.ctrlInfBlock[1]))
		}

	case DSKP_START_LIST:
		addr = dg_phys_addr(dskpData.commandRegA)<<16 | dg_phys_addr(dskpData.commandRegB)
		if dskpData.debug {
			debugPrint(dskpLog, "... START LIST command\n")
			debugPrint(dskpLog, "... ..... First CB Address: %d\n", addr)
		}
		// TODO should check addr validity before starting processing
		dskpProcessCB(addr)
		dskpData.statusRegA = dwordGetUpperWord(dg_dword(addr)) // return address of 1st CB processed
		dskpData.statusRegB = dwordGetLowerWord(dg_dword(addr))
		//dskpSetPioStatusRegC(0, STAT_CCS_PIO_CMD_OK, DSKP_START_LIST, testWbit(dskpData.commandRegC, 15))

	default:
		log.Panicf("DSKP command %d not yet implemented\n", pioCmd)
	}
}

func dskpExtractPioCommand(word dg_word) uint {
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
		if dskpData.debug {
			debugPrint(dskpLog, "... S flag set\n")
		}
		dskpDoPioCommand()
		busSetBusy(DEV_DSKP, false)
		busSetDone(DEV_DSKP, true)

	case 'C':
		if dskpData.debug {
			debugPrint(dskpLog, "... C flag set, clearing DONE flag\n")
		}
		busSetDone(DEV_DSKP, false)
		// TODO clear pending interrupt
		dskpSetPioStatusRegC(STAT_XEC_STATE_MAPPED,
			STAT_CCS_PIO_CMD_OK,
			dg_word(dskpExtractPioCommand(dskpData.commandRegC)),
			testWbit(dskpData.commandRegC, 15))

	case 'P':
		if dskpData.debug {
			debugPrint(dskpLog, "... P flag set\n")
		}
		log.Fatalln("P flag not yet implemented in DSKP")

	default:
		// no/empty flag - nothing to do
	}
}

// seek to the disk position according to sector number
func dskpPositionDiskImage(sectorNum dg_dword) {
	var offset int64 = int64(sectorNum) * dskpBytesPerSector
	_, err := dskpData.imageFile.Seek(offset, 0)
	if err != nil {
		log.Fatalln("DSKP could not position disk image")
	}
	// TODO Set C/H/S???
	dskpData.sectorNo = sectorNum
}

func dskpProcessCB(addr dg_phys_addr) {
	var (
		cb                 [DSKP_CB_MAX_SIZE]dg_word
		w, cbLength        int
		nextCB             dg_phys_addr
		sect, sectorNumber dg_dword
		physTransfers      bool
		physAddr           dg_phys_addr
		buffer             = make([]byte, dskpBytesPerSector)
		tmpWd              dg_word
	)
	cbLength = DSKP_CB_MIN_SIZE + dskpGetCBextendedStatusSize()
	if dskpData.debug {
		debugPrint(dskpLog, "... Processing CB, extended status size is: %d\n", dskpGetCBextendedStatusSize())
	}
	// copy CB contents from host memory
	for w = 0; w < cbLength; w++ {
		cb[w] = memReadWord(addr + dg_phys_addr(w))
		if dskpData.debug {
			debugPrint(dskpLog, "... CB[%d]: %d\n", w, cb[w])
		}
	}

	opCode := cb[DSKP_CB_INA_FLAGS_OPCODE] & 0x03ff
	nextCB = dg_phys_addr(cb[DSKP_CB_LINK_ADDR_HIGH])<<16 | dg_phys_addr(cb[DSKP_CB_LINK_ADDR_LOW])
	if dskpData.debug {
		debugPrint(dskpLog, "... CB OpCode: %d\n", opCode)
		debugPrint(dskpLog, "... .. Next CB Addr: %d\n", nextCB)
	}
	switch opCode {

	case DSKP_CB_OP_RECALIBRATE_DISK:
		if dskpData.debug {
			debugPrint(dskpLog, "... .. RECALIBRATE\n")
		}
		dskpData.cylinder = 0
		dskpData.head = 0
		dskpData.sector = 0
		dskpData.sectorNo = 0
		dskpData.imageFile.Seek(0, 0)
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
		sectorNumber = (dg_dword(cb[DSKP_CB_DEV_ADDR_HIGH]) << 16) | dg_dword(cb[DSKP_CB_DEV_ADDR_LOW])
		dskpData.sectorNo = sectorNumber
		if testWbit(cb[DSKP_CB_PAGENO_LIST_ADDR_HIGH], 0) {
			// logical premapped host address
			physTransfers = false
			log.Fatal("DSKP - CB READ from premapped logical addresses  Not Yet Implemented")
		} else {
			physTransfers = true
			physAddr = dg_phys_addr(cb[DSKP_CB_TXFER_ADDR_HIGH])<<16 | dg_phys_addr(cb[DSKP_CB_TXFER_ADDR_LOW])
		}
		if dskpData.debug {
			debugPrint(dskpLog, "... .. CB READ command, SECCNT: %d\n", cb[DSKP_CB_TXFER_COUNT])
			debugPrint(dskpLog, "... .. .. .... from sector:     %d\n", sectorNumber)
			debugPrint(dskpLog, "... .. .. .... from phys addr:  %d\n", physAddr)
			debugPrint(dskpLog, "... .. .. .... physical txfer?: %d\n", boolToInt(physTransfers))
		}
		for sect = 0; sect < dg_dword(cb[DSKP_CB_TXFER_COUNT]); sect++ {
			dskpPositionDiskImage(sectorNumber + sect)
			dskpData.imageFile.Read(buffer)
			for w = 0; w < dskpWordsPerSector; w++ {
				tmpWd = (dg_word(buffer[w*2]) << 8) | dg_word(buffer[(w*2)+1])
				memWriteWordBmcChan(physAddr+(dg_phys_addr(sect)*dskpWordsPerSector)+dg_phys_addr(w), tmpWd)
			}
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
		dskpData.reads++
		if dskpData.debug {
			debugPrint(dskpLog, "... .. .... READ command finished\n")
			debugPrint(dskpLog, "Last buffer: %X\n", buffer)
		}
		//dskpSetAsyncStatusRegC(STAT_XEC_STATE_MAPPED, STAT_ASYNC_NO_ERRORS)
		dskpSetAsyncStatusRegC(0, STAT_ASYNC_NO_ERRORS)

	case DSKP_CB_OP_WRITE:
		sectorNumber = (dg_dword(cb[DSKP_CB_DEV_ADDR_HIGH]) << 16) | dg_dword(cb[DSKP_CB_DEV_ADDR_LOW])
		dskpData.sectorNo = sectorNumber
		if testWbit(cb[DSKP_CB_PAGENO_LIST_ADDR_HIGH], 0) {
			// logical premapped host address
			physTransfers = false
			log.Fatal("DSKP - CB WRITE from premapped logical addresses  Not Yet Implemented")
		} else {
			physTransfers = true
			physAddr = dg_phys_addr(cb[DSKP_CB_TXFER_ADDR_HIGH])<<16 | dg_phys_addr(cb[DSKP_CB_TXFER_ADDR_LOW])
		}
		if dskpData.debug {
			debugPrint(dskpLog, "... .. CB WRITE command, SECCNT: %d\n", cb[DSKP_CB_TXFER_COUNT])
			debugPrint(dskpLog, "... .. .. ..... to sector:       %d\n", sectorNumber)
			debugPrint(dskpLog, "... .. .. ..... from phys addr:  %d\n", physAddr)
			debugPrint(dskpLog, "... .. .. ..... physical txfer?: %d\n", boolToInt(physTransfers))
		}
		for sect = 0; sect < dg_dword(cb[DSKP_CB_TXFER_COUNT]); sect++ {
			dskpPositionDiskImage(sectorNumber + sect)
			for w = 0; w < dskpWordsPerSector; w++ {
				tmpWd = memReadWordBmcChan(physAddr + (dg_phys_addr(sect) * dskpWordsPerSector) + dg_phys_addr(w))
				buffer[w*2] = byte(tmpWd >> 8)
				buffer[(w*2)+1] = byte(tmpWd & 0x00ff)
			}
			dskpData.imageFile.Write(buffer)
			if dskpData.debug {
				debugPrint(dskpLog, "Wrote buffer: %X\n", buffer)
			}
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
		dskpData.writes++
		//dskpSetAsyncStatusRegC(STAT_XEC_STATE_MAPPED, STAT_ASYNC_NO_ERRORS)
		dskpSetAsyncStatusRegC(0, STAT_ASYNC_NO_ERRORS)

	default:
		log.Fatalf("DSKP CB Command %d not yet implemented\n", opCode)
	}
	// write back CB
	for w = 0; w < cbLength; w++ {
		memWriteWordBmcChan(addr+dg_phys_addr(w), cb[w])
	}
	// chain to next CB?
	if nextCB != 0 {
		dskpProcessCB(nextCB)
	}
}

func dskpReset() {
	dskpResetMapping()
	dskpResetIntInfBlk()
	dskpResetCtrlrInfBlock()
	dskpResetUnitInfBlock()
	dskpData.statusRegB = 0
	dskpSetPioStatusRegC(STAT_XEC_STATE_RESET_DONE, 0, DSKP_RESET, testWbit(dskpData.commandRegC, 15))
	if dskpData.debug {
		debugPrint(dskpLog, "DSKP Reset via call to dskpReset()\n")
	}
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
	dskpData.intInfBlock[3] = 8<<12 | 30
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
	var logicalBlocks dg_dword = dskpLogicalBlocks
	dskpData.unitInfBlock[0] = 0
	dskpData.unitInfBlock[1] = 9<<12 | dskpUcodeRev
	dskpData.unitInfBlock[2] = dwordGetUpperWord(logicalBlocks) // 17.
	dskpData.unitInfBlock[3] = dwordGetLowerWord(logicalBlocks) // 43840.
	dskpData.unitInfBlock[4] = dskpBytesPerSector
	dskpData.unitInfBlock[5] = dskpUserCylinders
	dskpData.unitInfBlock[6] = ((dskpSurfacesPerDisk * dskpHeadsPerSurface) << 8) | (0x00ff & dskpSectorsPerTrack)
}

func dskpSetAsyncStatusRegC(stat byte, asyncIntCode dg_word) {
	dskpData.statusRegC = dg_word(stat) << 12
	dskpData.statusRegC |= (asyncIntCode & 0x03ff)
}

func dskpSetPioStatusRegC(stat byte, ccs byte, cmdEcho dg_word, rr bool) {
	dskpData.statusRegC = dg_word(stat) << 12
	dskpData.statusRegC |= (dg_word(ccs) & 3) << 10
	dskpData.statusRegC |= (cmdEcho & 0x01ff) << 1
	if rr {
		dskpData.statusRegC |= 1
	}
}
