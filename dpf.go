// dpf.go

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

// Here we are emulating the DPF device, specifically model 6061
// controller/drive combination which provides c.190MB of formatted capacity.
package main

import (
	"bufio"
	"fmt"
	"log"
	"mvemg/logging"
	"os"
	"time"
)

// Physical characteristics of the emulated disk
const (
	dpfSurfPerDisk  = 19 //5 // 19
	dpfSectPerTrack = 24
	dpfWordsPerSect = 256
	dpfBytesPerSect = dpfWordsPerSect * 2
	dpfPhysCyls     = 815
	dpfPhysByteSize = dpfSurfPerDisk * dpfSectPerTrack * dpfBytesPerSect * dpfPhysCyls
)

const (
	DPF_CMD_READ = iota
	DPF_CMD_RECAL
	DPF_CMD_SEEK
	DPF_CMD_STOP
	DPF_CMD_OFFSET_FWD
	DPF_CMD_OFFSET_REV
	DPF_CMD_WRITE_DISABLE
	DPF_CMD_RELEASE
	DPF_CMD_TRESPASS
	DPF_CMD_SET_ALT_MODE_1
	DPF_CMD_SET_ALT_MODE_2
	DPF_CMD_NO_OP
	DPF_CMD_VERIFY
	DPF_CMD_READ_BUFFS
	DPF_CMD_WRITE
	DPF_CMD_FORMAT
)
const (
	DPF_INS_MODE_NORMAL = iota
	DPF_INS_MODE_ALT_1
	DPF_INS_MODE_ALT_2
)
const (
	// drive statuses
	DPF_DRIVEFAULT = 1 << iota
	DPF_WRITEFAULT
	DPF_CLOCKFAULT
	DPF_POSNFAULT
	DPF_PACKUNSAFE
	DPF_POWERFAULT
	DPF_ILLEGALCMD
	DPF_INVALIDADDR
	DPF_UNUSED
	DPF_WRITEDIS
	DPF_OFFSET
	DPF_BUSY
	DPF_READY
	DPF_TRESPASSED
	DPF_RESERVED
	DPF_INVALID
)
const (
	// R/W statuses
	DPF_RWFAULT = 1 << iota
	DPF_DATALATE
	DPF_RWTIMEOUT
	DPF_VERIFY
	DPF_SURFSECT
	DPF_CYLINDER
	DPF_BADSECTOR
	DPF_ECC
	DPF_ILLEGALSECTOR
	DPF_PARITY
	DPF_DRIVE3DONE
	DPF_DRIVE2DONE
	DPF_DRIVE1DONE
	DPF_DRIVE0DONE
	DPF_RWDONE
	DPF_CONTROLFULL
)

// dpfStatsPeriodMs is the number of milliseconds between sending status updates
const dpfStatsPeriodMs = 500

type dpfDataT struct {
	// MV/Em internals...
	debug         bool
	imageAttached bool
	imageFileName string
	imageFile     *os.File
	// DG data...
	cmdDrvAddr      byte // 6-bit?
	command         int8 // 4-bit
	rwCommand       int8
	drive           uint8   // 2-bit
	mapEnabled      bool    // is the BMC addressing physical (0) or Mapped (1)
	memAddr         dg_word // self-incrementing on DG
	ema             uint8   // 5-bit
	cylAddr         dg_word // 10-bit
	surfAddr        uint8   // 5-bit - increments post-op
	sectAddr        uint8   // 5-bit - increments mid-op
	sectCnt         int8    // 5-bit - incrememts mid-op - signed
	ecc             dg_dword
	driveStatus     dg_word
	rwStatus        dg_word
	instructionMode int
	lastDOAwasSeek  bool
}

type dpfStatT struct {
	imageAttached bool
	cylinder      dg_word
	head, sector  uint8
}

var (
	dpfData   dpfDataT
	err       error
	cmdDecode [DPF_CMD_FORMAT + 1]string
)

// initialise the emulated DPF controller
func dpfInit(statsChann chan dpfStatT) {
	dpfData.debug = true

	cmdDecode = [...]string{"READ", "RECAL", "SEEK", "STOP", "OFFSET FWD", "OFFSET REV",
		"WRITE DISABLE", "RELEASE", "TRESPASS", "SET ALT MODE 1", "SET ALT MODE 2",
		"NO OP", "VERIFY", "READ BUFFERS", "WRITE", "FORMAT"}

	go dpfStatsSender(statsChann)

	busAddDevice(DEV_DPF, "DPF", DPF_PMB, false, true, true)
	busSetResetFunc(DEV_DPF, dpfReset)
	busSetDataInFunc(DEV_DPF, dpfDataIn)
	busSetDataOutFunc(DEV_DPF, dpfDataOut)
	dpfData.imageAttached = false
	dpfData.instructionMode = DPF_INS_MODE_NORMAL
	dpfData.driveStatus = DPF_READY
	dpfData.mapEnabled = false

}

// attempt to attach an extant MV/Em disk image to the running emulator
func dpfAttach(dNum int, imgName string) bool {
	// TODO Disk Number not currently used
	logging.DebugPrint(logging.DpfLog, "dpfAttach called for disk #%d with image <%s>\n", dNum, imgName)
	dpfData.imageFile, err = os.OpenFile(imgName, os.O_RDWR, 0755)
	if err != nil {
		logging.DebugPrint(logging.DpfLog, "Failed to open image for attaching\n")
		logging.DebugPrint(logging.DebugLog, "WARN: Failed to open DPF image <%s> for ATTach\n", imgName)
		return false
	}
	dpfData.imageFileName = imgName
	dpfData.imageAttached = true
	busSetAttached(DEV_DPF)
	return true
}

func dpfStatsSender(sChan chan dpfStatT) {
	var stats dpfStatT
	for {
		if dpfData.imageAttached {
			stats.imageAttached = true
			stats.cylinder = dpfData.cylAddr
			stats.head = dpfData.surfAddr
			stats.sector = dpfData.sectAddr
		} else {
			stats = dpfStatT{}
		}
		select {
		case sChan <- stats:
		default:
		}
		time.Sleep(time.Millisecond * dpfStatsPeriodMs)
	}
}

// Create an empty disk file of the correct size for the DPF emulator to use
func dpfCreateBlank(imgName string) bool {
	newFile, err := os.Create(imgName)
	if err != nil {
		return false
	}
	defer newFile.Close()
	logging.DebugPrint(logging.DpfLog, "dpfCreateBlank attempting to write %d bytes\n", dpfPhysByteSize)
	w := bufio.NewWriter(newFile)
	for b := 0; b < dpfPhysByteSize; b++ {
		w.WriteByte(0)
	}
	w.Flush()
	return true
}

// dpfDataIn implements the DIA/B/C I/O instructions for this device
func dpfDataIn(cpuPtr *CPU, iPtr *DecodedInstr, abc byte) {
	switch abc {
	case 'A':
		switch dpfData.instructionMode {
		case DPF_INS_MODE_NORMAL:
			cpuPtr.ac[iPtr.acd] = dg_dword(dpfData.rwStatus)
			logging.DebugPrint(logging.DpfLog, "DIA [Read Data Txfr Status] (Normal mode) returning %s for DRV=%d\n",
				wordToBinStr(dpfData.rwStatus), dpfData.drive)
		case DPF_INS_MODE_ALT_1:
			log.Fatal("DPF DIA (Alt Mode 1) not yet implemented")
		case DPF_INS_MODE_ALT_2:
			log.Fatal("DPF DIA (Alt Mode 2) not yet implemented")
		}
	case 'B':
		switch dpfData.instructionMode {
		case DPF_INS_MODE_NORMAL:
			cpuPtr.ac[iPtr.acd] = dg_dword(dpfData.driveStatus & 0xfeff)
			logging.DebugPrint(logging.DpfLog, "DIB [Read Drive Status] (normal mode) DRV=%d, %s to AC%d, PC: %d\n",
				dpfData.drive, wordToBinStr(dpfData.driveStatus), iPtr.acd, cpuPtr.pc)
		case DPF_INS_MODE_ALT_1:
			cpuPtr.ac[iPtr.acd] = dg_dword(0x8000) | (dg_dword(dpfData.ema) & 0x01f)
			//			if dpfData.mapEnabled {
			//				cpuPtr.ac[iPtr.acd] = dg_dword(dpfData.ema&0x1f) | 0x8000
			//			} else {
			//				cpuPtr.ac[iPtr.acd] = dg_dword(dpfData.ema & 0x1f)
			//			}
			logging.DebugPrint(logging.DpfLog, "DIB [Read EMA] (Alt Mode 1) returning: %d, PC: %d\n",
				cpuPtr.ac[iPtr.acd], cpuPtr.pc)
		case DPF_INS_MODE_ALT_2:
			log.Fatal("DPF DIB (Alt Mode 2) not yet implemented")
		}
	case 'C':
		var ssc dg_word = 0
		if dpfData.mapEnabled {
			ssc = 1 << 15
		}
		ssc |= (dg_word(dpfData.surfAddr) & 0x1f) << 10
		ssc |= (dg_word(dpfData.sectAddr) & 0x1f) << 5
		ssc |= (dg_word(dpfData.sectCnt) & 0x1f)
		cpuPtr.ac[iPtr.acd] = dg_dword(ssc)
		logging.DebugPrint(logging.DpfLog, "DPF DIC returning: %s\n", wordToBinStr(ssc))
	}

	dpfHandleFlag(iPtr.f) // TODO Is this go-able?
}

// dpfDataOut implements the DOA/B/C instructions for this device
// NIO is also routed here with a dummy abc flag value of N
func dpfDataOut(cpuPtr *CPU, iPtr *DecodedInstr, abc byte) {
	data := dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
	switch abc {
	case 'A':
		dpfData.command = extractDpfCommand(data)
		dpfData.drive = extractDpfDriveNo(data)
		dpfData.ema = extractDpfEMA(data)
		if testWbit(data, 0) {
			dpfData.rwStatus &= ^dg_word(DPF_RWDONE)
		}
		if testWbit(data, 1) {
			dpfData.rwStatus &= ^dg_word(DPF_DRIVE0DONE)
		}
		if testWbit(data, 2) {
			dpfData.rwStatus &= ^dg_word(DPF_DRIVE1DONE)
		}
		if testWbit(data, 3) {
			dpfData.rwStatus &= ^dg_word(DPF_DRIVE2DONE)
		}
		if testWbit(data, 4) {
			dpfData.rwStatus &= ^dg_word(DPF_DRIVE3DONE)
		}
		dpfData.instructionMode = DPF_INS_MODE_NORMAL
		if dpfData.command == DPF_CMD_SET_ALT_MODE_1 {
			dpfData.instructionMode = DPF_INS_MODE_ALT_1
		}
		if dpfData.command == DPF_CMD_SET_ALT_MODE_2 {
			dpfData.instructionMode = DPF_INS_MODE_ALT_2
		}
		if dpfData.command == DPF_CMD_NO_OP {
			dpfDoNoOpCommand()
		}
		dpfData.lastDOAwasSeek = (dpfData.command == DPF_CMD_SEEK)
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "DOA [Specify Cmd,Drv,EMA] to DRV=%d with data %s at PC: %d\n",
				dpfData.drive, wordToBinStr(data), cpuPtr.pc)
			logging.DebugPrint(logging.DpfLog, "... CMD: %s, DRV: %d, EMA: %d\n",
				cmdDecode[dpfData.command], dpfData.drive, dpfData.ema)
		}
	case 'B':
		if testWbit(data, 0) {
			dpfData.ema |= 0x01
		} else {
			dpfData.ema &= 0xfe
		}
		dpfData.memAddr = data & 0x7fff
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "DOB [Specify Memory Addr] with data %s at PC: %d\n",
				wordToBinStr(data), cpuPtr.pc)
			logging.DebugPrint(logging.DpfLog, "... MEM Addr: %d\n", dpfData.memAddr)
			logging.DebugPrint(logging.DpfLog, "... EMA: %d\n", dpfData.ema)
		}
	case 'C':
		if dpfData.lastDOAwasSeek {
			dpfData.cylAddr = data & 0x03ff
			if dpfData.debug {
				logging.DebugPrint(logging.DpfLog, "DOC [Specify Cylinder] after SEEK with data %s at PC: %d\n",
					wordToBinStr(data), cpuPtr.pc)
				logging.DebugPrint(logging.DpfLog, "... CYL: %d\n", dpfData.cylAddr)
			}
		} else {
			dpfData.mapEnabled = testWbit(data, 0)
			dpfData.surfAddr = extractSurfAddr(data)
			dpfData.sectAddr = extractSectAddr(data)
			dpfData.sectCnt = extractSectCnt(data)
			if dpfData.debug {
				logging.DebugPrint(logging.DpfLog, "DOC [Specify Surf,Sect,Cnt] (not after seek) with data %s at PC: %d\n",
					wordToBinStr(data), cpuPtr.pc)
				logging.DebugPrint(logging.DpfLog, "... MAP: %d, SURF: %d, SECT: %d, SECCNT: %d\n",
					BoolToInt(dpfData.mapEnabled), dpfData.surfAddr, dpfData.sectAddr, dpfData.sectCnt)
			}
		}
	case 'N': // dummy value for NIO - we just handle the flag below
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "NIO%c received\n", iPtr.f)
		}
	}

	dpfHandleFlag(iPtr.f) // TODO Is this go-able?
}

func dpfDoDriveOpCommand() {
	dpfData.instructionMode = DPF_INS_MODE_NORMAL
	switch dpfData.command {
	case DPF_CMD_RECAL:
		dpfData.cylAddr = 0
		dpfData.surfAddr = 0
		dpfData.rwStatus = DPF_DRIVE0DONE
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... RECAL done, %s\n", dpfPrintableAddr())
		}

	case DPF_CMD_SEEK:
		// action the seek
		dpfData.rwStatus = DPF_DRIVE0DONE
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... SEEK done, %s\n", dpfPrintableAddr())
		}

	default:
		log.Fatalf("DPF Drive Operation command # %d not yet impelemented", dpfData.command)
	}
}

func dpfDoNoOpCommand() {
	dpfData.instructionMode = DPF_INS_MODE_NORMAL
	dpfData.rwStatus = 0
	dpfData.driveStatus = DPF_READY
	if dpfData.debug {
		logging.DebugPrint(logging.DpfLog, "... NO OP command done\n")
	}
}

func dpfDoRWcommand() {
	var (
		buffer = make([]byte, dpfBytesPerSect)
		wd     dg_word
	)
	dpfData.instructionMode = DPF_INS_MODE_NORMAL

	switch dpfData.command {

	// ===== READ from DPF =====
	case DPF_CMD_READ:
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... READ command invoked %s\n", dpfPrintableAddr())
			logging.DebugPrint(logging.DpfLog, "... .... Start Address: %d\n", dpfData.memAddr)
		}
		dpfData.rwStatus = 0
		for dpfData.sectCnt != 0 {
			dpfCheckSectorPos()
			if !dpfCheckCylPos() {
				break
			}
			dpfPositionDiskImage()
			br, err := dpfData.imageFile.Read(buffer)

			if br != dpfBytesPerSect || err != nil {
				log.Fatalf("ERROR: unexpected return from DPF Image File Read: %s", err)
			}
			for w := 0; w < dpfWordsPerSect; w++ {
				wd = (dg_word(buffer[w*2]) << 8) | dg_word(buffer[(w*2)+1])
				//if dpfData.mapEnabled {
				memWriteWordBmcChan(dg_phys_addr(dpfData.memAddr), wd)
				//} else {
				// memWriteWord(dg_phys_addr(dpfData.memAddr), wd)
				//}
				dpfData.memAddr++
			}
			dpfData.sectAddr++
			dpfData.sectCnt++
			dpfCheckSectorPos()
			if !dpfCheckCylPos() {
				break
			}

		}
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... .... READ command finished %s\n", dpfPrintableAddr())
			logging.DebugPrint(logging.DpfLog, "Last buffer: %X", buffer)
			logging.DebugPrint(logging.DpfLog, "\n... .... Last Address: %d\n", dpfData.memAddr)
		}
		dpfData.rwStatus |= DPF_RWDONE //| DPF_DRIVE0DONE

	case DPF_CMD_WRITE:
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... WRITE command invoked %s\n", dpfPrintableAddr())
			logging.DebugPrint(logging.DpfLog, "... .....  Start Address: %d\n", dpfData.memAddr)
		}
		dpfData.rwStatus = 0
		for dpfData.sectCnt != 0 {
			dpfCheckSectorPos()
			if !dpfCheckCylPos() {
				break
			}
			dpfPositionDiskImage()
			for w := 0; w < dpfWordsPerSect; w++ {
				//if dpfData.mapEnabled {
				wd = memReadWordBmcChan(dg_phys_addr(dpfData.memAddr))
				//} else {
				//wd = memReadWord(dg_phys_addr(dpfData.memAddr))
				//}
				dpfData.memAddr++
				buffer[w*2] = byte((wd & 0xff00) >> 8)
				buffer[(w*2)+1] = byte(wd & 0x00ff)
			}
			bw, err := dpfData.imageFile.Write(buffer)
			if bw != dpfBytesPerSect || err != nil {
				log.Fatalf("ERROR: unexpected return from DPF Image File Write: %s", err)
			}
			dpfData.sectAddr++
			dpfData.sectCnt++
			dpfCheckSectorPos()
			if !dpfCheckCylPos() {
				break
			}

		}
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... ..... WRITE command finished %s\n", dpfPrintableAddr())
			logging.DebugPrint(logging.DpfLog, "%X", buffer)
			logging.DebugPrint(logging.DpfLog, "... ..... Last Address: %d\n", dpfData.memAddr)
		}
		dpfData.rwStatus |= DPF_RWDONE //| DPF_DRIVE0DONE

	default:
		log.Fatalf("DPF Disk R/W Command %d not yet implemented\n", dpfData.command)
	}

}

func dpfCheckSectorPos() bool {
	ok := true
	// end of track?
	if dpfData.sectAddr >= dpfSectPerTrack {
		dpfData.surfAddr++
		dpfData.sectAddr = 0
		ok = false
	}
	return ok
}

func dpfCheckCylPos() bool {
	ok := true
	// end of cylinder?
	if dpfData.surfAddr >= dpfSurfPerDisk {
		dpfData.surfAddr = 0
		dpfData.sectAddr = 0
		ok = false
	}
	return ok
}

func dpfHandleFlag(f byte) {
	switch f {
	case 'S':
		busSetBusy(DEV_DPF, true)
		busSetDone(DEV_DPF, false)
		// TODO stop any I/O
		dpfData.rwStatus = 0
		// TODO start I/O timeout
		dpfData.rwCommand = dpfData.command
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... S flag set\n")
		}
		dpfDoRWcommand()
		busSetBusy(DEV_DPF, false)
		busSetDone(DEV_DPF, true)

	case 'C':
		log.Fatal("DPF C flag not yet implemented")

	case 'P':
		busSetBusy(DEV_DPF, false)
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... P flag set\n")
		}
		dpfData.rwStatus = 0
		dpfDoDriveOpCommand()
		dpfData.rwStatus = DPF_DRIVE0DONE

	default:
		// no/empty flag - nothing to do
	}
}

// set the disk image file postion according to current C/H/S
func dpfPositionDiskImage() {
	var offset, r int64
	offset = int64(dpfData.cylAddr) * int64(dpfData.surfAddr) * int64(dpfData.sectAddr)
	r, err = dpfData.imageFile.Seek(offset, 0)
	if r != offset || err != nil {
		log.Fatal("DPF could not postition disk image via seek()")
	}
}

func dpfPrintableAddr() string {
	var pa string
	pa = fmt.Sprintf("DRV: %d, CYL: %d, SURF: %d, SECT: %d, SECCNT: %d",
		dpfData.drive, dpfData.cylAddr,
		dpfData.surfAddr, dpfData.sectAddr, dpfData.sectCnt)
	return pa
}

// reset the DPF controller
func dpfReset() {
	dpfData.instructionMode = DPF_INS_MODE_NORMAL
	dpfData.rwStatus = 0
	dpfData.driveStatus = DPF_READY
	if dpfData.debug {
		logging.DebugPrint(logging.DpfLog, "DPF Reset\n")
	}
}

func extractDpfCommand(word dg_word) int8 {
	return int8((word & 0x0780) >> 7)
}

func extractDpfDriveNo(word dg_word) uint8 {
	return uint8((word & 0x60) >> 5)
}

func extractDpfEMA(word dg_word) uint8 {
	return uint8(word & 0x1f)
}

func extractSectAddr(word dg_word) uint8 {
	return uint8((word & 0x03e0) >> 5)
}

func extractSectCnt(word dg_word) int8 {
	tmpWd := word & 0x01f
	if tmpWd != 0 { // sign-extend
		tmpWd |= 0xe0
	}
	return int8(tmpWd)
}

func extractSurfAddr(word dg_word) uint8 {
	return uint8((word & 0x7c00) >> 10)
}
