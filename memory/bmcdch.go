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

// Paraphrase of 1988 PoP...
//
// Eclipse MV/Family systems contain 512 DCH slots and 1024 BMC slots.
// Each 32-bit slot consists of two 16-bit map registers.
// These map registers and the I/O channel registers are numbered 0 thru 07777(8).
// The DCH and BMC registers contain page number and access information.
// The I/O channel registers contain status and control information which affect
// DCH and BMC maps and data transfers.
// For the map slots, the even-numbered registers are the most significant half of each slot
// and the odd-numbered are the least significant.

package memory

import (
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/util"
)

// See p.8-44 of PoP for meanings of these...
const (
	bmcRegs         = 2048
	firstDchSlotReg = bmcRegs
	dchRegs         = 1024
	totalRegs       = 4096  // 010000(8)
	iochanDefReg    = 06000 // 3072.
	// 06001-07677 are reserved
	iochanStatusReg   = 07700 // 4032.
	iochanMaskReg     = 07701 // 4033.
	cpuDedicationCtrl = 07702 // 4034.
	// 07703-07777 are reserved

	ioccdrICE = 1 << 15
	ioccdrBVE = 1 << 12
	ioccdrDVE = 1 << 11
	ioccdrDCH = 1 << 10
	ioccdrBMC = 1 << 9
	ioccdrBAP = 1 << 8
	ioccdrBDP = 1 << 7
	ioccdrDME = 1 << 1
	ioccdr1   = 1

	iocsrERR = 1 << 15
	iocsrDTO = 1 << 5
	iocsrMPE = 1 << 4
	iocsr1A  = 1 << 3
	iocsr1B  = 1 << 2
	iocsrCMB = 1 << 1
	iocsrINT = 1

	iocmrMK0 = 1 << 7
	iocmrMK1 = 1 << 6
	iocmrMK2 = 1 << 5
	iocmrMK3 = 1 << 4
	iocmrMK4 = 1 << 3
	iocmrMK5 = 1 << 2
	iocmrMK6 = 1 << 1
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

var regs [totalRegs]dg.WordT

// bmcdchInit is only called by MemInit()...
func bmcdchInit() {
	bmcdchReset()
	logging.DebugPrint(logging.MapLog, "BMC/DCH Map Registers Initialised\n")
}

func bmcdchReset() {
	for r := range regs {
		regs[r] = 0
	}
	regs[iochanDefReg] = ioccdr1
	regs[iochanStatusReg] = iocsr1A | iocsr1B
	regs[iochanMaskReg] = iocmrMK1 | iocmrMK2 | iocmrMK3 | iocmrMK4 | iocmrMK5 | iocmrMK6
}

func getDchMode() bool {
	logging.DebugPrint(logging.MapLog, "getDchMode returning: %d\n",
		util.BoolToInt(util.TestWbit(regs[iochanDefReg], 14)))
	return util.TestWbit(regs[iochanDefReg], 14)
}

// BmcdchWriteReg populates a given 16-bit register with the supplied data
// N.B. Addressed by REGISTER not slot
func BmcdchWriteReg(reg int, data dg.WordT) {
	logging.DebugPrint(logging.DebugLog, "bmcdchWriteReg: Reg %d, Data: %d\n", reg, data)
	regs[reg] = data
}

// BmcdchWriteSlot populates a whole SLOT (pair of registers) with the supplied doubleword
// N.B. Addressed by SLOT not register
func BmcdchWriteSlot(slot int, data dg.DwordT) {
	logging.DebugPrint(logging.DebugLog, "bmcdch*Write*Slot: Slot %d, Data: %d\n", slot, data)
	regs[slot*2] = util.DWordGetUpperWord(data)
	regs[(slot*2)+1] = util.DWordGetLowerWord(data)
}

// BmcdchReadReg returns the single word contents of the requested register
func BmcdchReadReg(reg int) dg.WordT {
	return regs[reg]
}

// BmcdchReadSlot returns the doubleword contents of the requested SLOT
func BmcdchReadSlot(slot int) dg.DwordT {
	return util.DWordFromTwoWords(regs[slot*2], regs[(slot*2)+1])
}

func getBmcMapAddr(mAddr dg.PhysAddrT) (dg.PhysAddrT, dg.PhysAddrT) {
	var page, slot, pAddr dg.PhysAddrT
	slot = mAddr >> 10
	/*** N.B. at some point between 1980 and 1987 the lower 5 bits of the odd word were
	  prepended to the even word to extend the mappable space */
	page = dg.PhysAddrT((regs[slot*2]&0x1f))<<16 + dg.PhysAddrT(regs[(slot*2)+1])<<10
	//page = dg.PhysAddrT(regs[(slot*2)+1]) << 10
	pAddr = (mAddr & 0x3ff) | page
	logging.DebugPrint(logging.MapLog, "getBmcMapAddr got: %d, slot: %d, regs[slot*2+1]: %d, page: %d, returning: %d\n",
		mAddr, slot, regs[(slot*2)+1], page, pAddr)
	return pAddr, page // TODO page return is just for debugging
}

func getDchMapAddr(mAddr dg.PhysAddrT) (dg.PhysAddrT, dg.PhysAddrT) {
	var page, slot, pAddr dg.PhysAddrT
	slot = mAddr >> 10
	/*** N.B. at some point between 1980 and 1987 the lower 5 bits of the odd word were
	  prepended to the even word to extend the mappable space */
	page = dg.PhysAddrT((regs[slot*2]&0x1f))<<16 + dg.PhysAddrT(regs[(slot*2)+1])<<10
	//page = dg.PhysAddrT(regs[(slot*2)+1]) << 10
	pAddr = (mAddr & 0x3ff) | page
	logging.DebugPrint(logging.MapLog, "getDchMapAddr got: %d, slot: %d, regs[slot*2+1]: %d, page: %d, returning: %d\n",
		mAddr, slot, regs[(slot*2)+1], page, pAddr)
	return pAddr, page // TODO page return is just for debugging
}

func decodeBmcAddr(bmcAddr dg.PhysAddrT) bmcAddrT {
	var (
		inAddr dg.DwordT
		res    bmcAddrT
	)

	inAddr = dg.DwordT(bmcAddr << 10) // shift left so we can use documented 21-bit numbering
	res.isLogical = util.TestDWbit(inAddr, 0)
	if res.isLogical {
		// Logical, or Mapped address...
		res.tt = byte(util.GetDWbits(inAddr, 2, 5))
		res.ttr = byte(util.GetDWbits(inAddr, 7, 5))
		res.plow = dg.PhysAddrT(bmcAddr & 0x3ff) // mask off 10 bits
	} else {
		// Physical, or unmapped address..
		res.bk = byte(util.GetDWbits(inAddr, 1, 3))
		res.xca = byte(util.GetDWbits(inAddr, 4, 3))
		res.ca = dg.PhysAddrT(bmcAddr & 0x7fff) // mask off 15 bits
	}

	return res
}

// ReadWordDchChan - reads a 16-bit word over the virtual DCH channel
func ReadWordDchChan(addr dg.PhysAddrT) dg.WordT {
	pAddr := addr
	if getDchMode() {
		pAddr, _ = getDchMapAddr(addr)
	}
	logging.DebugPrint(logging.MapLog, "ReadWordDchChan got addr: %d, read from addr: %d\n", addr, pAddr)
	return ReadWord(pAddr)
}

// ReadWordBmcChan reads a word from memory over the virtual Burst Multiplex Channel
func ReadWordBmcChan(addr dg.PhysAddrT) (dg.WordT, dg.PhysAddrT) {
	var pAddr dg.PhysAddrT
	decodedAddr := decodeBmcAddr(addr)
	if decodedAddr.isLogical {
		pAddr, _ = getBmcMapAddr(addr) // FIXME
	} else {
		pAddr = decodedAddr.ca
	}
	wd := ReadWord(pAddr)
	pAddr++
	logging.DebugPrint(logging.MapLog, "ReadWordBmcChan got addr: %d, wrote to addr: %d\n", addr, pAddr)
	return wd, pAddr
}

// ReadWordBmcChan16bit reads a word from memory over the virtual Burst Multiplex Channel for 16-bit devices
func ReadWordBmcChan16bit(addr *dg.WordT) dg.WordT {
	var pAddr dg.PhysAddrT
	decodedAddr := decodeBmcAddr(dg.PhysAddrT(*addr))
	if decodedAddr.isLogical {
		pAddr, _ = getBmcMapAddr(dg.PhysAddrT(*addr)) // FIXME
	} else {
		pAddr = decodedAddr.ca
	}
	wd := ReadWord(pAddr)
	*addr = *addr + 1
	logging.DebugPrint(logging.MapLog, "ReadWordBmcChan16bit got addr: %d, wrote to addr: %d\n", addr, pAddr)
	return wd
}

// WriteWordDchChan writes a word to memory over the virtual DCH
// pAddr is returned for debugging purposes only
func WriteWordDchChan(addr *dg.PhysAddrT, data dg.WordT) {
	pAddr := *addr

	if getDchMode() {
		pAddr, _ = getDchMapAddr(*addr)
	}
	WriteWord(pAddr, data)
	*addr = *addr + 1
	logging.DebugPrint(logging.MapLog, "WriteWordDchChan got addr: %d, wrote to addr: %d\n", addr, pAddr)

}

// WriteWordBmcChan writes a word over the virtual Burst Multiplex Channel
func WriteWordBmcChan(addr *dg.PhysAddrT, data dg.WordT) {
	var pAddr dg.PhysAddrT
	decodedAddr := decodeBmcAddr(*addr)
	if decodedAddr.isLogical {
		pAddr, _ = getBmcMapAddr(*addr) // FIXME
	} else {
		pAddr = decodedAddr.ca
	}
	WriteWord(pAddr, data)
	*addr = *addr + 1
	logging.DebugPrint(logging.MapLog, "WriteWordBmcChan got addr: %d, wrote to addr: %d\n", addr, pAddr)
}

// WriteWordBmcChan16bit writes a word over the virtual Burst Multiplex Channel for 16-bit devices
func WriteWordBmcChan16bit(addr *dg.WordT, data dg.WordT) {
	var pAddr dg.PhysAddrT
	decodedAddr := decodeBmcAddr(dg.PhysAddrT(*addr))
	if decodedAddr.isLogical {
		pAddr, _ = getBmcMapAddr(dg.PhysAddrT(*addr)) // FIXME
	} else {
		pAddr = decodedAddr.ca
	}
	WriteWord(pAddr, data)
	*addr = *addr + 1
	logging.DebugPrint(logging.MapLog, "WriteWordBmcChan got addr: %d, wrote to addr: %d\n", addr, pAddr)
}
