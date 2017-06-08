// dskp.go
// Here we are emulating the DSKP device, specifically model 6239/6240
// controller/drive combination which provides 592MB of formatted capacity.
package main

import (
	"bufio"
	//"fmt"
	"log"
	"os"
)

const (
	DSKP_SURFACES_PER_DISK  = 8
	DSKP_HEADS_PER_SURFACE  = 2
	DSKP_SECTORS_PER_TRACK  = 75
	DSKP_WORDS_PER_SECTOR   = 256
	DSKP_BYTES_PER_SECTOR   = 512
	DSKP_PHYSICAL_CYLINDERS = 981
	DSKP_USER_CYLINDERS     = 978
	DSKP_LOGICAL_BLOCKS     = 1157952 // ???
	DSKP_UCODE_REV          = 99

	DSKP_INT_INF_BLK_SIZE   = 8
	DSKP_CTRLR_INF_BLK_SIZE = 2
	DSKP_UNIT_INF_BLK_SIZE  = 7
	DSKP_CB_MAX_SIZE        = 21
	DSKP_CB_MIN_SIZE        = 12 // Was 10

	STAT_XEC_STATE_RESETTING  = 0x00
	STAT_XEC_STATE_RESET_DONE = 0x01
	STAT_XEC_STATE_BEGUN      = 0x08
	STAT_XEC_STATE_MAPPED     = 0x0c
	STAT_XEC_STATE_DIAG_MODE  = 0x04

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
	DSKP_PHYSICAL_BYTE_SIZE  = DSKP_SURFACES_PER_DISK * DSKP_HEADS_PER_SURFACE * DSKP_SECTORS_PER_TRACK * DSKP_BYTES_PER_SECTOR * DSKP_PHYSICAL_CYLINDERS
	DSKP_PHYSICAL_BLOCK_SIZE = DSKP_SURFACES_PER_DISK * DSKP_HEADS_PER_SURFACE * DSKP_SECTORS_PER_TRACK * DSKP_PHYSICAL_CYLINDERS
)

type dskpData_t struct {
	// MV/Em internals...
	debug         bool
	imageAttached bool
	imageFileName string
	imageFile     *os.File
	// DG data...
	commandRegA, commandRegB, commandRegC dg_word
	statusRegA, statusRegB, statusRegC    dg_word
	mappingRegA, mappingRegB              dg_word
	intInfBlock                           [DSKP_INT_INF_BLK_SIZE]dg_word
	ctrlInfBlock                          [DSKP_CTRLR_INF_BLK_SIZE]dg_word
	unitInfBlock                          [DSKP_UNIT_INF_BLK_SIZE]dg_word
	cylinder, head, sector                dg_word
}

var (
	dskpData dskpData_t
)

func dskpInit() {
	debugPrint(DSKP_LOG, "DSKP Initialising via call to dskpInit()...\n")

	dskpData.debug = true

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
	debugPrint(DSKP_LOG, "dskpAttach called for disk #%d with image <%s>\n", dNum, imgName)
	dskpData.imageFile, err = os.OpenFile(imgName, os.O_RDWR, 0755)
	if err != nil {
		debugPrint(DSKP_LOG, "Failed to open image for attaching\n")
		debugPrint(DEBUG_LOG,"WARN: Failed to open dskp image <%s> for ATTach\n", imgName)
		return false
	}
	dskpData.imageFileName = imgName
	dskpData.imageAttached = true
	busSetAttached(DEV_DSKP)
	return true
}

// Create an empty disk file of the correct size for the DSKP emulator to use
func dskpCreateBlank(imgName string) bool {
	newFile, err := os.Create(imgName)
	if err != nil {
		return false
	}
	defer newFile.Close()
	debugPrint(DSKP_LOG, "dskpCreateBlank attempting to write %d bytes\n", DSKP_PHYSICAL_BYTE_SIZE)
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
		debugPrint(DSKP_LOG, "DIA [Read Status A] returning %s for DRV=%d\n", wordToBinStr(dskpData.statusRegA), 0)
	case 'B':
		cpuPtr.ac[iPtr.acd] = dg_dword(dskpData.statusRegB)
		debugPrint(DSKP_LOG, "DIB [Read Status B] returning %s for DRV=%d\n", wordToBinStr(dskpData.statusRegB), 0)
	case 'C':
		cpuPtr.ac[iPtr.acd] = dg_dword(dskpData.statusRegC)
		debugPrint(DSKP_LOG, "DIC [Read Status C] returning %s for DRV=%d\n", wordToBinStr(dskpData.statusRegC), 0)
	}
	dskpHandleFlag(iPtr.f)
}

// Handle the DOA/B/C PIO commands
func dskpDataOut(cpuPtr *Cpu, iPtr *DecodedInstr, abc byte) {
	switch abc {
	case 'A':
		dskpData.commandRegA = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if dskpData.debug {
			debugPrint(DSKP_LOG, "DOA [Load Cmd Reg A] from AC%d containing %s, PC: %d\n",
				iPtr.acd, wordToBinStr(dskpData.commandRegA), cpuPtr.pc)
		}
	case 'B':
		dskpData.commandRegB = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if dskpData.debug {
			debugPrint(DSKP_LOG, "DOB [Load Cmd Reg B] from AC%d containing %s, PC: %d\n",
				iPtr.acd, wordToBinStr(dskpData.commandRegB), cpuPtr.pc)
		}
	case 'C':
		dskpData.commandRegC = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if dskpData.debug {
			debugPrint(DSKP_LOG, "DOC [Load Cmd Reg C] from AC%d containing %s, PC: %d\n",
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
			debugPrint(DSKP_LOG, "... BEGIN command, unit # %d\n", dskpData.commandRegA)
		}
		// pretend we have succesfully booted ourself
		dskpData.statusRegC = STAT_XEC_STATE_BEGUN << 12
		if testWbit(dskpData.commandRegC, 15) {
			dskpData.statusRegC |= (3 << 10)
		}
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... ..... returning %s\n", wordToBinStr(dskpData.statusRegC))
		}

	case DSKP_GET_MAPPING:
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... GET MAPPING command\n")
		}
		dskpData.statusRegA = dskpData.mappingRegA
		dskpData.statusRegB = dskpData.mappingRegB
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... ... Status Reg A set to %s\n", wordToBinStr(dskpData.statusRegA))
			debugPrint(DSKP_LOG, "... ... Status Reg B set to %s\n", wordToBinStr(dskpData.statusRegB))
		}

	case DSKP_SET_MAPPING:
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... SET MAPPING command\n")
		}
		dskpData.mappingRegA = dskpData.commandRegA
		dskpData.mappingRegB = dskpData.commandRegB
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... ... Mapping Reg A set to %s\n", wordToBinStr(dskpData.commandRegA))
			debugPrint(DSKP_LOG, "... ... Mapping Reg B set to %s\n", wordToBinStr(dskpData.commandRegB))
		}

	case DSKP_GET_INTERFACE:
		addr = dg_phys_addr(dskpData.commandRegA)<<16 | dg_phys_addr(dskpData.commandRegB)
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... GET INTERFACE INFO command\n")
			debugPrint(DSKP_LOG, "... ... Destination Start Address: %d\n", addr)
		}
		for w = 0; w < DSKP_INT_INF_BLK_SIZE; w++ {
			memWriteWord(addr+w, dskpData.intInfBlock[w])
			if dskpData.debug {
				debugPrint(DSKP_LOG, "... ... Word %d: %s\n", w, wordToBinStr(dskpData.intInfBlock[w]))
			}
		}

	case DSKP_SET_INTERFACE:
		addr = dg_phys_addr(dskpData.commandRegA)<<16 | dg_phys_addr(dskpData.commandRegB)
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... SET INTERFACE INFO command\n")
			debugPrint(DSKP_LOG, "... ... Origin Start Address: %d\n", addr)
		}
		// only a few fields can be changed...
		w = 5
		dskpData.intInfBlock[w] = memReadWord(addr+w) & 0xff00
		w = 6
		dskpData.intInfBlock[w] = memReadWord(addr + w)
		w = 7
		dskpData.intInfBlock[w] = memReadWord(addr + w)

	case DSKP_GET_UNIT:
		addr = dg_phys_addr(dskpData.commandRegA)<<16 | dg_phys_addr(dskpData.commandRegB)
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... GET UNIT INFO command\n")
			debugPrint(DSKP_LOG, "... ... Destination Start Address: %d\n", addr)
		}
		for w = 0; w < DSKP_UNIT_INF_BLK_SIZE; w++ {
			memWriteWord(addr+w, dskpData.unitInfBlock[w])
			if dskpData.debug {
				debugPrint(DSKP_LOG, "... ... Word %d: %s\n", w, wordToBinStr(dskpData.unitInfBlock[w]))
			}
		}

	case DSKP_SET_UNIT:
		addr = dg_phys_addr(dskpData.commandRegA)<<16 | dg_phys_addr(dskpData.commandRegB)
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... SET UNIT INFO command\n")
			debugPrint(DSKP_LOG, "... ... Origin Start Address: %d\n", addr)
		}
		// only the first word is writable according to p.2-16
		// TODO check no active CBs first
		dskpData.unitInfBlock[0] = memReadWord(addr)

	case DSKP_RESET:
		dskpReset()

	case DSKP_SET_CONTROLLER:
		addr = dg_phys_addr(dskpData.commandRegA)<<16 | dg_phys_addr(dskpData.commandRegB)
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... SET CONTROLLER INFO command\n")
			debugPrint(DSKP_LOG, "... ... Origin Start Address: %d\n", addr)
		}
		dskpData.ctrlInfBlock[0] = memReadWord(addr)
		dskpData.ctrlInfBlock[1] = memReadWord(addr + 1)

	case DSKP_START_LIST:
		addr = dg_phys_addr(dskpData.commandRegA)<<16 | dg_phys_addr(dskpData.commandRegB)
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... START LIST command\n")
			debugPrint(DSKP_LOG, "... ..... First CB Address: %d\n", addr)
		}
		// TODO should check addr validity before starting processing
		dskpProcessCB(addr)
		dskpData.statusRegA = dwordGetUpperWord(dg_dword(addr)) // return address of 1st CB processed
		dskpData.statusRegB = dwordGetLowerWord(dg_dword(addr))

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
			debugPrint(DSKP_LOG, "... S flag set\n")
		}
		dskpDoPioCommand()
		busSetBusy(DEV_DSKP, false)
		busSetDone(DEV_DSKP, true)
	case 'C':
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... C flag set, clearing DONE flag\n")
		}
		busSetDone(DEV_DSKP, false)
		// TODO clear pending interrupt
	case 'P':
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... P flag set\n")
		}
		log.Fatalln("P flag not yet implemented in DSKP")
	default:
		// no/empty flag - nothing to do
	}
}

// seek to the disk position according to sector number
func dskpPositionDiskImage(sectorNum dg_dword) {
	var offset int64 = int64(sectorNum) * DSKP_BYTES_PER_SECTOR
	_, err := dskpData.imageFile.Seek(offset, 0)
	if err != nil {
		log.Fatalln("DSKP could not position disk image")
	}
	// TODO Set C/H/S???
}

func dskpProcessCB(addr dg_phys_addr) {
	var (
		cb                 [DSKP_CB_MAX_SIZE]dg_word
		w                  int
		nextCB             dg_phys_addr
		sect, sectorNumber dg_dword
		physTransfers      bool
		physAddr           dg_phys_addr
		buffer             = make([]byte, DSKP_BYTES_PER_SECTOR)
		tmpWd              dg_word
	)

	// copy CB contents from host memory
	for w = 0; w < DSKP_CB_MIN_SIZE; w++ {
		cb[w] = memReadWord(addr + dg_phys_addr(w))
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... CB[%d]: %d\n", w, cb[w])
		}
	}
	for w = DSKP_CB_MIN_SIZE; w < DSKP_CB_MAX_SIZE; w++ {
		cb[w] = 0
	}
	opCode := cb[DSKP_CB_INA_FLAGS_OPCODE] & 0x03ff
	nextCB = dg_phys_addr(cb[DSKP_CB_LINK_ADDR_HIGH])<<16 | dg_phys_addr(cb[DSKP_CB_LINK_ADDR_LOW])
	if dskpData.debug {
		debugPrint(DSKP_LOG, "... CB OpCode: %d\n", opCode)
		debugPrint(DSKP_LOG, "... .. Next CB Addr: %d\n", nextCB)
	}
	switch opCode {

	case DSKP_CB_OP_RECALIBRATE_DISK:
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... .. RECALIBRATE\n")
		}
		dskpData.cylinder = 0
		dskpData.head = 0
		dskpData.sector = 0
		dskpData.imageFile.Seek(0, 0)
		cb[DSKP_CB_ERR_STATUS] = 0
		cb[DSKP_CB_UNIT_STATUS] = 1 << 13 // b0010000000000000; // Ready
		dskpData.statusRegC = 5           // 005;     Not full, no errors
		cb[DSKP_CB_CB_STATUS] = 1         // finally, set Done bit

	case DSKP_CB_OP_READ:
		sectorNumber = (dg_dword(cb[DSKP_CB_DEV_ADDR_HIGH]) << 16) | dg_dword(cb[DSKP_CB_DEV_ADDR_LOW])
		if testWbit(cb[DSKP_CB_PAGENO_LIST_ADDR_HIGH], 0) {
			// logical premapped host address
			physTransfers = false
			log.Fatal("DSKP - CB READ from premapped logical addresses  Not Yet Implemented")
		} else {
			physTransfers = true
			physAddr = dg_phys_addr(cb[DSKP_CB_TXFER_ADDR_HIGH])<<16 | dg_phys_addr(cb[DSKP_CB_TXFER_ADDR_LOW])
		}
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... .. CB READ command, SECCNT: %d\n", cb[DSKP_CB_TXFER_COUNT])
			debugPrint(DSKP_LOG, "... .. .. .... to sector:       %d\n", sectorNumber)
			debugPrint(DSKP_LOG, "... .. .. .... from phys addr:  %d\n", physAddr)
			debugPrint(DSKP_LOG, "... .. .. .... physical txfer?: %d\n", boolToInt(physTransfers))
		}
		for sect = 0; sect < dg_dword(cb[DSKP_CB_TXFER_COUNT]); sect++ {
			dskpPositionDiskImage(sectorNumber + sect)
			dskpData.imageFile.Read(buffer)
			for w = 0; w < DSKP_WORDS_PER_SECTOR; w++ {
				tmpWd = (dg_word(buffer[w*2]) << 8) | dg_word(buffer[(w*2)+1])
				memWriteWordBmcChan(physAddr+(dg_phys_addr(sect)*DSKP_WORDS_PER_SECTOR)+dg_phys_addr(w), tmpWd)
			}
		}
		cb[DSKP_CB_ERR_STATUS] = 0
		cb[DSKP_CB_UNIT_STATUS] = 1 << 13 // Ready
		dskpData.statusRegC = 5           // not full, no errrors
		cb[DSKP_CB_CB_STATUS] = 1         // finally, set done bit

	case DSKP_CB_OP_WRITE:
		sectorNumber = (dg_dword(cb[DSKP_CB_DEV_ADDR_HIGH]) << 16) | dg_dword(cb[DSKP_CB_DEV_ADDR_LOW])
		if testWbit(cb[DSKP_CB_PAGENO_LIST_ADDR_HIGH], 0) {
			// logical premapped host address
			physTransfers = false
			log.Fatal("DSKP - CB WRITE from premapped logical addresses  Not Yet Implemented")
		} else {
			physTransfers = true
			physAddr = dg_phys_addr(cb[DSKP_CB_TXFER_ADDR_HIGH])<<16 | dg_phys_addr(cb[DSKP_CB_TXFER_ADDR_LOW])
		}
		if dskpData.debug {
			debugPrint(DSKP_LOG, "... .. CB WRITE command, SECCNT: %d\n", cb[DSKP_CB_TXFER_COUNT])
			debugPrint(DSKP_LOG, "... .. .. ..... to sector:       %d\n", sectorNumber)
			debugPrint(DSKP_LOG, "... .. .. ..... from phys addr:  %d\n", physAddr)
			debugPrint(DSKP_LOG, "... .. .. ..... physical txfer?: %d\n", boolToInt(physTransfers))
		}
		for sect = 0; sect < dg_dword(cb[DSKP_CB_TXFER_COUNT]); sect++ {
			dskpPositionDiskImage(sectorNumber + sect)
			for w = 0; w < DSKP_WORDS_PER_SECTOR; w++ {
				tmpWd = memReadWordBmcChan(physAddr + (dg_phys_addr(sect) * DSKP_WORDS_PER_SECTOR) + dg_phys_addr(w))
				buffer[w*2] = byte(tmpWd >> 8)
				buffer[(w*2)+1] = byte(tmpWd & 0x00ff)
			}
			dskpData.imageFile.Write(buffer)
		}
		cb[DSKP_CB_ERR_STATUS] = 0
		cb[DSKP_CB_UNIT_STATUS] = 1 << 13 // Ready
		dskpData.statusRegC = 5           // not full, no errrors
		cb[DSKP_CB_CB_STATUS] = 1         // finally, set done bit

	default:
		log.Fatalf("DSKP CB Command %d not yet implemented\n", opCode)
	}
	// write back CB
	for w = 0; w < (DSKP_CB_MIN_SIZE + dskpGetCBextendedStatusSize()); w++ {
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
	dskpData.statusRegC = STAT_XEC_STATE_RESET_DONE << 12
	if dskpData.debug {
		debugPrint(DSKP_LOG, "DSKP Reset via call to dskpReset()\n")
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
	dskpData.intInfBlock[1] = DSKP_UCODE_REV
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
	var logicalBlocks dg_dword = DSKP_LOGICAL_BLOCKS
	dskpData.unitInfBlock[0] = 0
	dskpData.unitInfBlock[1] = 9<<8 | DSKP_UCODE_REV
	dskpData.unitInfBlock[2] = dwordGetUpperWord(logicalBlocks)
	dskpData.unitInfBlock[3] = dwordGetLowerWord(logicalBlocks)
	dskpData.unitInfBlock[4] = DSKP_BYTES_PER_SECTOR
	dskpData.unitInfBlock[5] = DSKP_USER_CYLINDERS
	dskpData.unitInfBlock[6] = ((DSKP_SURFACES_PER_DISK * DSKP_HEADS_PER_SURFACE) << 8) | (0x00ff & DSKP_SECTORS_PER_TRACK)
}
