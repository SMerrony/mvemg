// bmcdch.go

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

// Each "slot" contains two 16-bit "registers", it seems the contents can be accessed either
// by slot as a doubleword, OR by slot + high or low word, OR by word directly.
package memory

import (
	"log"

	"mvemg/dg"
	"mvemg/logging"
	"mvemg/util"
)

// See p.8-44 of PoP for meanings of these...
const (
	BMC_REGS           = 1024
	DCH_REGS           = 512
	BMCDCH_REGS        = 4096
	IOCHAN_DEF_REG     = 06000
	IOCHAN_STATUS_REG  = 07700
	IOCHAN_MASK_REG    = 07701
	CPU_DEDICATION_CTL = 07702

	IOCDR_ICE = 1 << 15
	IOCDR_BVE = 1 << 12
	IOCDR_DVE = 1 << 11
	IOCDR_DCH = 1 << 10
	IOCDR_BMC = 1 << 9
	IOCDR_BAP = 1 << 8
	IOCDR_BDP = 1 << 7
	IOCDR_DME = 1 << 1
	IOCDR_1   = 1

	IOCSR_ERR = 1 << 15
	IOCSR_DTO = 1 << 5
	IOCSR_MPE = 1 << 4
	IOCSR_1A  = 1 << 3
	IOCSR_1B  = 1 << 2
	IOCSR_CMB = 1 << 1
	IOCSR_INT = 1

	IOCMR_MK0 = 1 << 7
	IOCMR_MK1 = 1 << 6
	IOCMR_MK2 = 1 << 5
	IOCMR_MK3 = 1 << 4
	IOCMR_MK4 = 1 << 3
	IOCMR_MK5 = 1 << 2
	IOCMR_MK6 = 1 << 1
)

type bmcAddrT struct {
	isLogical bool // is this a Physical(f) or Logical(t) address?

	// physical addresses...
	bk  byte         // bank selection bits (3-bit)
	xca byte         // eXtended Channel Addr bits (3-bit)
	ca  dg.PhysAddrT // Channel Addr (15-bit)

	// logical addresess..
	tt   byte         // Translation Table (5-bit)
	ttr  byte         // TT Register (5-bit)
	plow dg.PhysAddrT // Page Low Order Word (10-bit)
}

var regs [BMCDCH_REGS]dg.WordT

// bmcdchInit is only called by MemInit()...
func bmcdchInit() {
	bmcdchReset()
	log.Println("INFO: BMC/DCH Maps Initialised")
	logging.DebugPrint(logging.MapLog, "BMC/DCH Maps Initialised\n")
}

func bmcdchReset() {
	for r := range regs {
		regs[r] = 0
	}
	regs[IOCHAN_DEF_REG] = IOCDR_1
	regs[IOCHAN_STATUS_REG] = IOCSR_1A | IOCSR_1B
	regs[IOCHAN_MASK_REG] = IOCMR_MK1 | IOCMR_MK2 | IOCMR_MK3 | IOCMR_MK4 | IOCMR_MK5 | IOCMR_MK6
}

func getDchMode() bool {
	logging.DebugPrint(logging.MapLog, "getDchMode returning: %d\n",
		util.BoolToInt(util.TestWbit(regs[IOCHAN_DEF_REG], 14)))
	return util.TestWbit(regs[IOCHAN_DEF_REG], 14)
}

func BmcdchWriteReg(reg int, data dg.WordT) {
	logging.DebugPrint(logging.DebugLog, "bmcdchWriteReg: Reg %d, Data: %d\n", reg, data)
	regs[reg] = data
}

func BmcdchWriteSlot(slot int, data dg.DwordT) {
	logging.DebugPrint(logging.DebugLog, "bmcdch*Write*Slot: Slot %d, Data: %d\n", slot, data)
	regs[slot*2] = util.DWordGetUpperWord(data)
	regs[(slot*2)+1] = util.DWordGetLowerWord(data)
}

func BmcdchReadReg(reg int) dg.WordT {
	return regs[reg]
}

func BmcdchReadSlot(slot int) dg.DwordT {
	return util.DWordFromTwoWords(regs[slot*2], regs[(slot*2)+1])
}

func getBmcDchMapAddr(mAddr dg.PhysAddrT) (dg.PhysAddrT, dg.PhysAddrT) {
	var page, slot, pAddr dg.PhysAddrT
	slot = mAddr >> 10
	/*** TODO: at some point between 1980 and 1987 the lower 5 bits of the even word were
	  prepended to the even word to extend the mappable space */
	//page = ((regs[slot] & 0x1f) << 16) + (regs[slot+1] << 10);
	page = dg.PhysAddrT(regs[(slot*2)+1]) << 10
	pAddr = (mAddr & 0x3ff) | page
	logging.DebugPrint(logging.MapLog, "getBmcDchMapAddr got: %d, slot: %d, regs[slot*2+1]: %d, page: %d, returning: %d\n",
		mAddr, slot, regs[(slot*2)+1], page, pAddr)
	return pAddr, page // TODO page return is just for debugging
}

func decodeBmcAddr(bmcAddr dg.PhysAddrT) bmcAddrT {
	var (
		inAddr dg.DwordT
		res    bmcAddrT
	)

	inAddr = dg.DwordT(bmcAddr << 10) // shift lest so we can use documented 21-bit numbering
	res.isLogical = util.TestDWbit(inAddr, 0)
	if res.isLogical {
		// Logical, or Mapped address...
		res.tt = byte(util.GetDWbits(inAddr, 2, 5))
		res.ttr = byte(util.GetDWbits(inAddr, 7, 5))
		// for performance use the parm directly here...
		res.plow = dg.PhysAddrT(bmcAddr & 0x3ff) // mask off 10 bits
	} else {
		// Physical, or unmapped address..
		res.bk = byte(util.GetDWbits(inAddr, 1, 3))
		res.xca = byte(util.GetDWbits(inAddr, 4, 3))
		// for performance use the parm directly here...
		res.ca = dg.PhysAddrT(bmcAddr & 0x7fff) // mask off 15 bits
	}

	return res
}

func MemReadWordDchChan(addr dg.PhysAddrT) dg.WordT {
	pAddr := addr
	if getDchMode() {
		pAddr, _ = getBmcDchMapAddr(addr)
	}
	logging.DebugPrint(logging.MapLog, "memReadWordBmcChan got addr: %d, read from addr: %d\n", addr, pAddr)
	return ReadWord(pAddr)
}

func MemReadWordBmcChan(addr dg.PhysAddrT) dg.WordT {
	var pAddr dg.PhysAddrT
	decodedAddr := decodeBmcAddr(addr)
	if decodedAddr.isLogical {
		pAddr, _ = getBmcDchMapAddr(addr) // FIXME
	} else {
		pAddr = decodedAddr.ca
	}
	logging.DebugPrint(logging.MapLog, "memWriteReadBmcChan got addr: %d, wrote to addr: %d\n", addr, pAddr)
	return ReadWord(pAddr)
}

func MemWriteWordDchChan(addr dg.PhysAddrT, data dg.WordT) dg.PhysAddrT {
	pAddr := addr

	if getDchMode() {
		pAddr, _ = getBmcDchMapAddr(addr)
	}
	WriteWord(pAddr, data)
	logging.DebugPrint(logging.MapLog, "memWriteWordDchChan got addr: %d, wrote to addr: %d\n", addr, pAddr)
	return pAddr
}

func MemWriteWordBmcChan(addr dg.PhysAddrT, data dg.WordT) dg.PhysAddrT {
	var pAddr dg.PhysAddrT
	decodedAddr := decodeBmcAddr(addr)
	if decodedAddr.isLogical {
		pAddr, _ = getBmcDchMapAddr(addr) // FIXME
	} else {
		pAddr = decodedAddr.ca
	}
	WriteWord(pAddr, data)
	logging.DebugPrint(logging.MapLog, "memWriteWordBmcChan got addr: %d, wrote to addr: %d\n", addr, pAddr)
	return pAddr
}
