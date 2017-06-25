// mtb.go
package main

import (
	//"bytes"
	"log"
	"mvemg/logging"
	"os"
)

const (
	MTB_MAX_RECORD_SIZE = 16384
	MTB_EOF             = 0
	MTB_LOG_FILE        = "mtb_debug.log"
	MTB_CMD_COUNT       = 11
	MTB_CMD_MASK        = 0x00b8

	CMD_READ_BITS         = 0x0000
	CMD_REWIND_BITS       = 0x0008
	CMD_CTRL_MODE_BITS    = 0x0010
	CMD_SPACE_FWD_BITS    = 0x0018
	CMD_SPACE_REV_BITS    = 0x0020
	CMD_WRITE_BITS        = 0x0028
	CMD_WRITE_EOF_BITS    = 0x0030
	CMD_ERASE_BITS        = 0x0038
	CMD_READ_NONSTOP_BITS = 0x0080
	CMD_UNLOAD_BITS       = 0x0088
	CMD_DRIVE_MODE_BITS   = 0x0090

	CMD_READ         = 0
	CMD_REWIND       = 1
	CMD_CTRL_MODE    = 2
	CMD_SPACE_FWD    = 3
	CMD_SPACE_REV    = 4
	CMD_WRITE        = 5
	CMD_WRITE_EOF    = 6
	CMD_ERASE        = 7
	CMD_READ_NONSTOP = 8
	CMD_UNLOAD       = 9
	CMD_DRIVE_MODE   = 10

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

// const mtbStatsPeriodMs = 500

// type mtbStatT struct {
// 	imageAttached bool

// }

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

	mtb        Mtb
	commandSet [MTB_CMD_COUNT]dg_word

	mtbLog *log.Logger
)

func mtbInit() bool {
	commandSet[CMD_READ] = CMD_READ_BITS
	commandSet[CMD_REWIND] = CMD_REWIND_BITS
	commandSet[CMD_CTRL_MODE] = CMD_CTRL_MODE_BITS
	commandSet[CMD_SPACE_FWD] = CMD_SPACE_FWD_BITS
	commandSet[CMD_SPACE_REV] = CMD_SPACE_REV_BITS
	commandSet[CMD_WRITE] = CMD_WRITE_BITS
	commandSet[CMD_WRITE_EOF] = CMD_WRITE_EOF_BITS
	commandSet[CMD_ERASE] = CMD_ERASE_BITS
	commandSet[CMD_READ_NONSTOP] = CMD_READ_NONSTOP_BITS
	commandSet[CMD_UNLOAD] = CMD_UNLOAD_BITS
	commandSet[CMD_DRIVE_MODE] = CMD_DRIVE_MODE_BITS

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
func mtbLoadTBoot(mem memoryT) {
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
		logging.DebugPrint(logging.DebugLog, "WARN: mtbLoadTBoot found mismatched trailer in TBOOT file\n")
	}
}

// This is called from Bus to implement DIx from the MTB device
func mtbDataIn(cpuPtr *CPU, iPtr *DecodedInstr, abc byte) {

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
func mtbDataOut(cpuPtr *CPU, iPtr *DecodedInstr, abc byte) {

	switch iPtr.f {
	case 'S':
		busSetBusy(DEV_MTB, true)
		busSetDone(DEV_MTB, false)
	case 'C':
		busSetBusy(DEV_MTB, false)
		busSetDone(DEV_MTB, false)
	}

	ac16 := dwordGetLowerWord(cpuPtr.ac[iPtr.acd])

	switch abc {
	case 'A': // Specify Command and Drive - p.IV-17
		// which command?
		for c := 0; c < MTB_CMD_COUNT; c++ {
			if (ac16 & MTB_CMD_MASK) == commandSet[c] {
				mtb.currentCmd = c
				break
			}
		}
		mtbLog.Printf("DOA - Specify Command and Drive - internal cmd #: %d, PC: %d\n",
			mtb.currentCmd, cpuPtr.pc)

	case 'B':
		mtb.memAddrReg = dg_phys_addr(ac16)
		mtbLog.Printf("DOB - Write Memory Address Register from AC%d, Value: %d, PC: %d\n",
			iPtr.acd, ac16, cpuPtr.pc)

	case 'C':
		mtb.negWordCntReg = int(int16(ac16))
		mtbLog.Printf("DOC - Set (neg) Word Count to %d, PC: %d\n",
			mtb.negWordCntReg, cpuPtr.pc)
	}

	if iPtr.f == 'S' {
		mtbDoCommand() // TODO Can this be a goroutine?
	}
}

func mtbDoCommand() {

	switch mtb.currentCmd {
	case CMD_READ:
		mtbLog.Printf("*READ* command\n ==== Word Count: %d Location: %d\n", mtb.negWordCntReg, mtb.memAddrReg)
		hdrLen, _ := simht.simhTapeReadRecordHeader(0)
		mtbLog.Printf(" ----  Header read giving length: %d\n", hdrLen)
		if hdrLen == MTB_EOF {
			mtbLog.Printf(" ----  Header is EOF indicator\n")
			mtb.statusReg1 = SR1_HI_DENSITY | SR1_9TRACK | SR1_UNIT_READY | SR1_EOF | SR1_ERROR
		} else {
			mtbLog.Printf(" ----  Calling simhTapeReadRecord with length: %d\n", hdrLen)
			var w dg_dword
			var wd dg_word
			var pAddr dg_phys_addr
			rec, _ := simht.simhTapeReadRecord(0, int(hdrLen))
			for w = 0; w < hdrLen; w += 2 {
				wd = (dg_word(rec[w]) << 8) | dg_word(rec[w+1])
				pAddr = memWriteWordDchChan(mtb.memAddrReg, wd)
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
			mtb.statusReg1 = SR1_HI_DENSITY | SR1_9TRACK | SR1_UNIT_READY
		}
		busSetBusy(DEV_MTB, false)
		busSetDone(DEV_MTB, true)

	case CMD_REWIND:
		mtbLog.Printf("*REWIND* command\n")
		simht.simhTapeRewind(0)
		mtb.statusReg1 = SR1_HI_DENSITY | SR1_9TRACK | SR1_UNIT_READY | SR1_BOT
		mtb.statusReg2 = SR2_PE_MODE
		// FIXME set flags here?

	case CMD_SPACE_FWD:
		mtbLog.Printf("*SPACE FORWARD* command\n")
		simht.simhTapeSpaceFwd(0, 0)
		mtb.statusReg1 = SR1_HI_DENSITY | SR1_9TRACK | SR1_UNIT_READY | SR1_EOF | SR1_ERROR
		busSetBusy(DEV_MTB, false)
		busSetDone(DEV_MTB, true)

	default:
		log.Fatalln("ERROR: mtbDoCommand - Command Not Yet Implemented")
	}
}
