// bmcdch.go
package main

import (
	"log"
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

var regs [BMCDCH_REGS]dg_word

func bmcdchInit() {
	for r, _ := range regs {
		regs[r] = 0
	}
	bmcdchReset()
	log.Println("INFO: BMC/DCH Maps Initialised")
}

func bmcdchReset() {
	// TODO should we clear the regs?
	regs[IOCHAN_DEF_REG] = IOCDR_1
	regs[IOCHAN_STATUS_REG] = IOCSR_1A | IOCSR_1B
	regs[IOCHAN_MASK_REG] = IOCMR_MK1 | IOCMR_MK2 | IOCMR_MK3 | IOCMR_MK4 | IOCMR_MK5 | IOCMR_MK6
}

func getDchMode() bool {
	log.Printf("DEBUG: getDchMode returning: %d\n", testWbit(regs[IOCHAN_DEF_REG], 14))
	return testWbit(regs[IOCHAN_DEF_REG], 14)
}

func bmcdchWriteSlot(slot int, data dg_word) {
	log.Printf("DEBUG: bmcdchWriteSlot: Slot %d, Data: %d\n", slot, data)
	regs[slot] = data
}

func bmcdchReadSlot(slot int) dg_word {
	return regs[slot]
}

func getBmcDchMapAddr(mAddr dg_phys_addr) dg_phys_addr {
	var page, slot, pAddr dg_phys_addr
	slot = mAddr >> 10
	/*** TODO: at some point between 1980 and 1987 the lower 5 bits of the even word were
	  prepended to the even word to extend the mappable space */
	//page = ((regs[slot] & 0x1f) << 16) + (regs[slot+1] << 10);
	page = dg_phys_addr(regs[slot+1]) << 10
	pAddr = (mAddr & 0x3ff) | page
	log.Printf("DEBUG: getBmcDchMapAddr got: %d, slot: %d, returning: %d\n", mAddr, slot, pAddr)
	return pAddr
}
