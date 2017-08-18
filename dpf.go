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
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
	"mvemg/util"
	"os"
	"sync"
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
	dpfCmdRead = iota
	dpfCmdRecal
	dpfCmdSeek
	dpfCmdStop
	dpfCmdOffsetFwd
	dpfCmdOffsetRev
	dpfCmdWriteDisable
	dpfCmdRelease
	dpfCmdTrespass
	dpfCmdSetAltMode_1
	dpfCmdSetAltMode_2
	dpfCmdNoOp
	dpfCmdVerify
	dpfCmdReadBuffs
	dpfCmdWrite
	dpfCmdFormat
)
const (
	dpfInsModeNormal = iota
	dpfInsModeAlt_1
	dpfInsModeAlt_2
)
const (
	// drive statuses
	dpfDrivefault = 1 << iota
	dpfWritefault
	dpfClockfault
	dpfPosnfault
	dpfPackunsafe
	dpfPowerfault
	dpfIllegalcmd
	dpfInvalidaddr
	dpfUnused
	dpfWritedis
	dpfOffset
	dpfBusy
	dpfReady
	dpfTrespassed
	dpfReserved
	dpfInvalid
)
const (
	// R/W statuses
	dpfRwfault = 1 << iota
	dpfDatalate
	dpfRwtimeout
	dpfVerify
	dpfSurfsect
	dpfCylinder
	dpfBadsector
	dpfEcc
	dpfIllegalsector
	dpfParity
	dpfDrive3Done
	dpfDrive2Done
	dpfDrive1Done
	dpfDrive0Done
	dpfRwdone
	dpfControlfull
)

// dpfStatsPeriodMs is the number of milliseconds between sending status updates
const dpfStatsPeriodMs = 500

type dpfDataT struct {
	// MV/Em internals...
	debug               bool
	imageAttached       bool
	dpfMu               sync.RWMutex
	imageFileName       string
	imageFile           *os.File
	reads, writes       uint64
	readBuff, writeBuff []byte
	// DG data...
	cmdDrvAddr      byte // 6-bit?
	command         int8 // 4-bit
	rwCommand       int8
	drive           uint8    // 2-bit
	mapEnabled      bool     // is the BMC addressing physical (0) or Mapped (1)
	memAddr         dg.WordT // self-incrementing on DG
	ema             uint8    // 5-bit
	cylinder        dg.WordT // 10-bit
	surface         uint8    // 5-bit - increments post-op
	sector          uint8    // 5-bit - increments mid-op
	sectCnt         int8     // 5-bit - incrememts mid-op - signed
	ecc             dg.DwordT
	driveStatus     dg.WordT
	rwStatus        dg.WordT
	instructionMode int
	lastDOAwasSeek  bool
}

// DpfStatT holds the data reported to the status collector
type DpfStatT struct {
	imageAttached bool
	cylinder      dg.WordT
	head, sector  uint8
	reads, writes uint64
}

var (
	dpfData   dpfDataT
	err       error
	cmdDecode [dpfCmdFormat + 1]string
)

// DpfInit must be called to initialise the emulated DPF controller
func DpfInit(statsChann chan DpfStatT) {
	dpfData.dpfMu.Lock()
	defer dpfData.dpfMu.Unlock()
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
	dpfData.instructionMode = dpfInsModeNormal
	dpfData.driveStatus = dpfReady
	dpfData.mapEnabled = false
	dpfData.readBuff = make([]byte, dpfBytesPerSect)
	dpfData.writeBuff = make([]byte, dpfBytesPerSect)
}

// attempt to attach an extant MV/Em disk image to the running emulator
func dpfAttach(dNum int, imgName string) bool {
	// TODO Disk Number not currently used
	logging.DebugPrint(logging.DpfLog, "dpfAttach called for disk #%d with image <%s>\n", dNum, imgName)
	dpfData.dpfMu.Lock()
	dpfData.imageFile, err = os.OpenFile(imgName, os.O_RDWR, 0755)
	if err != nil {
		logging.DebugPrint(logging.DpfLog, "Failed to open image for attaching\n")
		logging.DebugPrint(logging.DebugLog, "WARN: Failed to open DPF image <%s> for ATTach\n", imgName)
		dpfData.dpfMu.Unlock()
		return false
	}
	dpfData.imageFileName = imgName
	dpfData.imageAttached = true
	dpfData.dpfMu.Unlock()
	busSetAttached(DEV_DPF)
	return true
}

func dpfStatsSender(sChan chan DpfStatT) {
	var stats DpfStatT
	for {
		dpfData.dpfMu.RLock()
		if dpfData.imageAttached {
			stats.imageAttached = true
			stats.cylinder = dpfData.cylinder
			stats.head = dpfData.surface
			stats.sector = dpfData.sector
			stats.reads = dpfData.reads
			stats.writes = dpfData.writes
		} else {
			stats = DpfStatT{}
		}
		dpfData.dpfMu.RUnlock()
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
func dpfDataIn(cpuPtr *CPU, iPtr *decodedInstrT, abc byte) {
	dpfData.dpfMu.RLock()
	switch abc {
	case 'A':
		switch dpfData.instructionMode {
		case dpfInsModeNormal:
			cpuPtr.ac[iPtr.acd] = dg.DwordT(dpfData.rwStatus)
			logging.DebugPrint(logging.DpfLog, "DIA [Read Data Txfr Status] (Normal mode) returning %s for DRV=%d\n",
				util.WordToBinStr(dpfData.rwStatus), dpfData.drive)
		case dpfInsModeAlt_1:
			log.Fatal("DPF DIA (Alt Mode 1) not yet implemented")
		case dpfInsModeAlt_2:
			log.Fatal("DPF DIA (Alt Mode 2) not yet implemented")
		}
	case 'B':
		switch dpfData.instructionMode {
		case dpfInsModeNormal:
			cpuPtr.ac[iPtr.acd] = dg.DwordT(dpfData.driveStatus & 0xfeff)
			logging.DebugPrint(logging.DpfLog, "DIB [Read Drive Status] (normal mode) DRV=%d, %s to AC%d, PC: %d\n",
				dpfData.drive, util.WordToBinStr(dpfData.driveStatus), iPtr.acd, cpuPtr.pc)
		case dpfInsModeAlt_1:
			cpuPtr.ac[iPtr.acd] = dg.DwordT(0x8000) | (dg.DwordT(dpfData.ema) & 0x01f)
			//			if dpfData.mapEnabled {
			//				cpuPtr.ac[iPtr.acd] = dg_dword(dpfData.ema&0x1f) | 0x8000
			//			} else {
			//				cpuPtr.ac[iPtr.acd] = dg_dword(dpfData.ema & 0x1f)
			//			}
			logging.DebugPrint(logging.DpfLog, "DIB [Read EMA] (Alt Mode 1) returning: %d, PC: %d\n",
				cpuPtr.ac[iPtr.acd], cpuPtr.pc)
		case dpfInsModeAlt_2:
			log.Fatal("DPF DIB (Alt Mode 2) not yet implemented")
		}
	case 'C':
		var ssc dg.WordT
		if dpfData.mapEnabled {
			ssc = 1 << 15
		}
		ssc |= (dg.WordT(dpfData.surface) & 0x1f) << 10
		ssc |= (dg.WordT(dpfData.sector) & 0x1f) << 5
		ssc |= (dg.WordT(dpfData.sectCnt) & 0x1f)
		cpuPtr.ac[iPtr.acd] = dg.DwordT(ssc)
		logging.DebugPrint(logging.DpfLog, "DPF DIC returning: %s\n", util.WordToBinStr(ssc))
	}
	dpfData.dpfMu.RUnlock()

	dpfHandleFlag(iPtr.f)
}

// dpfDataOut implements the DOA/B/C instructions for this device
// NIO is also routed here with a dummy abc flag value of N
func dpfDataOut(cpuPtr *CPU, iPtr *decodedInstrT, abc byte) {
	dpfData.dpfMu.Lock()
	data := util.DWordGetLowerWord(cpuPtr.ac[iPtr.acd])
	switch abc {
	case 'A':
		dpfData.command = extractDpfCommand(data)
		dpfData.drive = extractDpfDriveNo(data)
		dpfData.ema = extractDpfEMA(data)
		if util.TestWbit(data, 0) {
			dpfData.rwStatus &= ^dg.WordT(dpfRwdone)
		}
		if util.TestWbit(data, 1) {
			dpfData.rwStatus &= ^dg.WordT(dpfDrive0Done)
		}
		if util.TestWbit(data, 2) {
			dpfData.rwStatus &= ^dg.WordT(dpfDrive1Done)
		}
		if util.TestWbit(data, 3) {
			dpfData.rwStatus &= ^dg.WordT(dpfDrive2Done)
		}
		if util.TestWbit(data, 4) {
			dpfData.rwStatus &= ^dg.WordT(dpfDrive3Done)
		}
		dpfData.instructionMode = dpfInsModeNormal
		if dpfData.command == dpfCmdSetAltMode_1 {
			dpfData.instructionMode = dpfInsModeAlt_1
		}
		if dpfData.command == dpfCmdSetAltMode_2 {
			dpfData.instructionMode = dpfInsModeAlt_2
		}
		if dpfData.command == dpfCmdNoOp {
			dpfData.instructionMode = dpfInsModeNormal
			dpfData.rwStatus = 0
			dpfData.driveStatus = dpfReady
			if dpfData.debug {
				logging.DebugPrint(logging.DpfLog, "... NO OP command done\n")
			}
		}
		dpfData.lastDOAwasSeek = (dpfData.command == dpfCmdSeek)
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "DOA [Specify Cmd,Drv,EMA] to DRV=%d with data %s at PC: %d\n",
				dpfData.drive, util.WordToBinStr(data), cpuPtr.pc)
			logging.DebugPrint(logging.DpfLog, "... CMD: %s, DRV: %d, EMA: %d\n",
				cmdDecode[dpfData.command], dpfData.drive, dpfData.ema)
		}
	case 'B':
		if util.TestWbit(data, 0) {
			dpfData.ema |= 0x01
		} else {
			dpfData.ema &= 0xfe
		}
		dpfData.memAddr = data & 0x7fff
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "DOB [Specify Memory Addr] with data %s at PC: %d\n",
				util.WordToBinStr(data), cpuPtr.pc)
			logging.DebugPrint(logging.DpfLog, "... MEM Addr: %d\n", dpfData.memAddr)
			logging.DebugPrint(logging.DpfLog, "... EMA: %d\n", dpfData.ema)
		}
	case 'C':
		if dpfData.lastDOAwasSeek {
			dpfData.cylinder = data & 0x03ff // mask off lower 10 bits
			if dpfData.debug {
				logging.DebugPrint(logging.DpfLog, "DOC [Specify Cylinder] after SEEK with data %s at PC: %d\n",
					util.WordToBinStr(data), cpuPtr.pc)
				logging.DebugPrint(logging.DpfLog, "... CYL: %d\n", dpfData.cylinder)
			}
		} else {
			dpfData.mapEnabled = util.TestWbit(data, 0)
			dpfData.surface = extractsurface(data)
			dpfData.sector = extractSector(data)
			dpfData.sectCnt = extractSectCnt(data)
			if dpfData.debug {
				logging.DebugPrint(logging.DpfLog, "DOC [Specify Surf,Sect,Cnt] (not after seek) with data %s at PC: %d\n",
					util.WordToBinStr(data), cpuPtr.pc)
				logging.DebugPrint(logging.DpfLog, "... MAP: %d, SURF: %d, SECT: %d, SECCNT: %d\n",
					util.BoolToInt(dpfData.mapEnabled), dpfData.surface, dpfData.sector, dpfData.sectCnt)
			}
		}
	case 'N': // dummy value for NIO - we just handle the flag below
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "NIO%c received\n", iPtr.f)
		}
	}
	dpfData.dpfMu.Unlock()

	dpfHandleFlag(iPtr.f)
}

func dpfDoCommand() {
	var (
		//buffer = make([]byte, dpfBytesPerSect)
		wd dg.WordT
	)
	dpfData.dpfMu.Lock()

	dpfData.instructionMode = dpfInsModeNormal

	switch dpfData.command {

	// RECALibrate (goto pos. 0)
	case dpfCmdRecal:
		dpfData.cylinder = 0
		dpfData.surface = 0
		dpfPositionDiskImage()
		dpfData.driveStatus = dpfReady
		dpfData.rwStatus = dpfRwdone | dpfDrive0Done
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... RECAL done, %s\n", dpfPrintableAddr())
		}

	// SEEK
	case dpfCmdSeek:
		// action the seek
		dpfPositionDiskImage()
		dpfData.driveStatus = dpfReady
		dpfData.rwStatus = dpfRwdone | dpfDrive0Done
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... SEEK done, %s\n", dpfPrintableAddr())
		}

	// ===== READ from DPF =====
	case dpfCmdRead:
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... READ command invoked %s\n", dpfPrintableAddr())
			logging.DebugPrint(logging.DpfLog, "... .... Start Address: %d\n", dpfData.memAddr)
		}
		dpfData.rwStatus = 0

		for dpfData.sectCnt != 0 {
			// check CYL
			if dpfData.cylinder >= dpfPhysCyls {
				dpfData.driveStatus = dpfReady
				dpfData.rwStatus = dpfRwdone | dpfRwfault | dpfCylinder
				dpfData.dpfMu.Unlock()
				return
			}
			// check SECT
			if dpfData.sector >= dpfSectPerTrack {
				dpfData.sector = 0
				dpfData.surface++
				if dpfData.debug {
					logging.DebugPrint(logging.DpfLog, "Sector read overflow, advancing to surface %d",
						dpfData.surface)
				}
				// dpfData.driveStatus = dpfReady
				// dpfData.rwStatus = dpfRwdone | dpfRwfault | DPF_ILLEGALSECTOR
				// dpfData.dpfMu.Unlock()
				// return
			}
			// check SURF (head)
			if dpfData.surface >= dpfSurfPerDisk {
				dpfData.driveStatus = dpfReady
				dpfData.rwStatus = dpfRwdone | dpfRwfault | dpfIllegalsector
				dpfData.dpfMu.Unlock()
				return
			}
			dpfPositionDiskImage()
			br, err := dpfData.imageFile.Read(dpfData.readBuff)

			if br != dpfBytesPerSect || err != nil {
				log.Fatalf("ERROR: unexpected return from DPF Image File Read: %s", err)
			}
			for w := 0; w < dpfWordsPerSect; w++ {
				wd = (dg.WordT(dpfData.readBuff[w*2]) << 8) | dg.WordT(dpfData.readBuff[(w*2)+1])
				memory.MemWriteWordBmcChan(dg.PhysAddrT(dpfData.memAddr), wd)
				dpfData.memAddr++
			}
			dpfData.sector++
			dpfData.sectCnt++
			dpfData.reads++

			if dpfData.debug {
				logging.DebugPrint(logging.DpfLog, "Buffer: %X\n", dpfData.readBuff)
			}

		}
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... .... READ command finished %s\n", dpfPrintableAddr())
			logging.DebugPrint(logging.DpfLog, "\n... .... Last Address: %d\n", dpfData.memAddr)
		}
		dpfData.rwStatus = dpfRwdone //| dpfDrive0Done

	case dpfCmdWrite:
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... WRITE command invoked %s\n", dpfPrintableAddr())
			logging.DebugPrint(logging.DpfLog, "... .....  Start Address: %d\n", dpfData.memAddr)
		}
		dpfData.rwStatus = 0

		for dpfData.sectCnt != 0 {
			// check CYL
			if dpfData.cylinder >= dpfPhysCyls {
				dpfData.driveStatus = dpfReady
				dpfData.rwStatus = dpfRwdone | dpfRwfault | dpfCylinder
				dpfData.dpfMu.Unlock()
				return
			}
			// check SECT
			if dpfData.sector >= dpfSectPerTrack {
				dpfData.sector = 0
				dpfData.surface++
				if dpfData.debug {
					logging.DebugPrint(logging.DpfLog, "Sector write overflow, advancing to surface %d",
						dpfData.surface)
				}
				// dpfData.driveStatus = dpfReady
				// dpfData.rwStatus = dpfRwdone | dpfRwfault | DPF_ILLEGALSECTOR
				// dpfData.dpfMu.Unlock()
				// return
			}
			// check SURF (head)/SECT
			if dpfData.surface >= dpfSurfPerDisk {
				dpfData.driveStatus = dpfReady
				dpfData.rwStatus = dpfRwdone | dpfRwfault | dpfIllegalsector
				dpfData.dpfMu.Unlock()
				return
			}
			dpfPositionDiskImage()
			for w := 0; w < dpfWordsPerSect; w++ {
				wd = memory.MemReadWordBmcChan(dg.PhysAddrT(dpfData.memAddr))
				dpfData.memAddr++
				dpfData.writeBuff[w*2] = byte((wd & 0xff00) >> 8)
				dpfData.writeBuff[(w*2)+1] = byte(wd & 0x00ff)
			}
			bw, err := dpfData.imageFile.Write(dpfData.writeBuff)
			if bw != dpfBytesPerSect || err != nil {
				log.Fatalf("ERROR: unexpected return from DPF Image File Write: %s", err)
			}
			dpfData.sector++
			dpfData.sectCnt++
			dpfData.writes++

			if dpfData.debug {
				logging.DebugPrint(logging.DpfLog, "Buffer: %X\n", dpfData.writeBuff)
			}
		}
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... ..... WRITE command finished %s\n", dpfPrintableAddr())
			logging.DebugPrint(logging.DpfLog, "... ..... Last Address: %d\n", dpfData.memAddr)
		}
		dpfData.driveStatus = dpfReady
		dpfData.rwStatus = dpfRwdone //| dpfDrive0Done

	default:
		log.Fatalf("DPF Disk R/W Command %d not yet implemented\n", dpfData.command)
	}
	dpfData.dpfMu.Unlock()
}

func dpfHandleFlag(f byte) {
	switch f {
	case 'S':
		busSetBusy(DEV_DPF, true)
		busSetDone(DEV_DPF, false)
		// TODO stop any I/O
		dpfData.dpfMu.Lock()
		dpfData.rwStatus = 0
		// TODO start I/O timeout
		dpfData.rwCommand = dpfData.command
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... S flag set\n")
		}
		dpfData.dpfMu.Unlock()
		dpfDoCommand()
		busSetBusy(DEV_DPF, false)
		busSetDone(DEV_DPF, true)

	case 'C':
		log.Fatal("DPF C flag not yet implemented")

	case 'P':
		busSetBusy(DEV_DPF, false)
		dpfData.dpfMu.Lock()
		if dpfData.debug {
			logging.DebugPrint(logging.DpfLog, "... P flag set\n")
		}
		dpfData.rwStatus = 0
		dpfData.dpfMu.Unlock()
		dpfDoCommand()
		//dpfData.rwStatus = dpfDrive0Done

	default:
		// no/empty flag - nothing to do
	}
}

// set the MV/Em disk image file postion according to current C/H/S
func dpfPositionDiskImage() {
	var offset, r int64
	offset = int64(dpfData.cylinder) * int64(dpfData.surface) * int64(dpfData.sector) * dpfBytesPerSect
	r, err = dpfData.imageFile.Seek(offset, 0)
	if r != offset || err != nil {
		log.Fatal("DPF could not postition disk image via seek()")
	}
}

func dpfPrintableAddr() string {
	var pa string
	// MUST BE LOCKED BY CALLER
	pa = fmt.Sprintf("DRV: %d, CYL: %d, SURF: %d, SECT: %d, SECCNT: %d",
		dpfData.drive, dpfData.cylinder,
		dpfData.surface, dpfData.sector, dpfData.sectCnt)
	return pa
}

// reset the DPF controller
func dpfReset() {
	dpfData.dpfMu.Lock()
	dpfData.instructionMode = dpfInsModeNormal
	dpfData.rwStatus = 0
	dpfData.driveStatus = dpfReady
	if dpfData.debug {
		logging.DebugPrint(logging.DpfLog, "DPF Reset\n")
	}
	dpfData.dpfMu.Unlock()
}

func extractDpfCommand(word dg.WordT) int8 {
	return int8((word & 0x0780) >> 7)
}

func extractDpfDriveNo(word dg.WordT) uint8 {
	return uint8((word & 0x60) >> 5)
}

func extractDpfEMA(word dg.WordT) uint8 {
	return uint8(word & 0x1f)
}

func extractSector(word dg.WordT) uint8 {
	return uint8((word & 0x03e0) >> 5)
}

func extractSectCnt(word dg.WordT) int8 {
	tmpWd := word & 0x01f
	if tmpWd != 0 { // sign-extend
		tmpWd |= 0xe0
	}
	return int8(tmpWd)
}

func extractsurface(word dg.WordT) uint8 {
	return uint8((word & 0x7c00) >> 10)
}
