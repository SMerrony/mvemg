// dpf.go
// Here we are emulating the DPF device, specifically model 6061
// controller/drive combination which provides c.190MB of formatted capacity.
package main

import (
	"bufio"
	"fmt"
	"os"
)

const (
	DPF_SURFACES_PER_DISK  = 19
	DPF_SECTORS_PER_TRACK  = 24
	DPF_BYTES_PER_SECTOR   = 512
	DPF_WORDS_PER_SECTOR   = 256
	DPF_PHYSICAL_CYLINDERS = 815

	DPF_PHYSICAL_BYTE_SIZE = DPF_SURFACES_PER_DISK * DPF_SECTORS_PER_TRACK * DPF_BYTES_PER_SECTOR * DPF_PHYSICAL_CYLINDERS
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

type dpfData_t struct {
	// MV/Em internals...
	debug         bool
	imageAttached bool
	imageFileName string
	imageFile     *os.File
	// DG data...
	cmdDrvAddr      byte // 6-bit?
	command         int8 // 4-bit
	rwCommand       int8
	drive           uint8 // 2-bit
	mapEnabled      bool
	memAddr         dg_word // self-incrementing on DG
	ema             uint8   // 5-bit
	clyAddr         dg_word // 10-bit
	surfAddr        uint8   // 5-bit - increments post-op
	sectAddr        uint8   // 5-bit - increments mid-op
	sectCnt         int8    // 5-bit - incrememts mid-op - signed
	ecc             dg_dword
	driveStatus     dg_word
	rwStatus        dg_word
	instructionMode int
	lastDOAwasSeek  bool
}

var (
	dpfData dpfData_t
)

// initialise the emulated DPF controller
func dpfInit() {
	dpfData.debug = true

	busAddDevice(DEV_DPF, "DPF", DPF_PMB, false, true, true)
	busSetResetFunc(DEV_DPF, dpfReset)

	dpfData.imageAttached = false
	dpfData.instructionMode = DPF_INS_MODE_NORMAL
	dpfData.driveStatus = DPF_READY

}

// Create an empty disk file of the correct size for the DPF emulator to use
func dpfCreateBlank(imgName string) bool {
	newFile, err := os.Create(imgName)
	if err != nil {
		return false
	}
	defer newFile.Close()
	debugPrint(DPF_LOG, fmt.Sprintf("dpfCreateBlank attempting to write %d bytes\n", DPF_PHYSICAL_BYTE_SIZE))
	w := bufio.NewWriter(newFile)
	for b := 0; b < DPF_PHYSICAL_BYTE_SIZE; b++ {
		w.WriteByte(0)
	}
	w.Flush()
	return true
}

// reset the DPF controller
func dpfReset() {
	dpfData.instructionMode = DPF_INS_MODE_NORMAL
	dpfData.rwStatus = 0
	dpfData.driveStatus = DPF_READY
	if dpfData.debug {
		debugPrint(DPF_LOG, "DPF Reset\n")
	}
}
