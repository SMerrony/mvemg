// bmcdch_test.go
package main

import (
	"testing"
)

func TestBmcdchReset(t *testing.T) {
	var wd DgWordT
	bmcdchInit()
	wd = regs[IOCHAN_DEF_REG]
	if wd != IOCDR_1 {
		t.Error("Got ", wd)
	}
}

func TestWriteReadMapSlot(t *testing.T) {
	var dwd1, dwd2 DgDwordT
	dwd1 = 0x11223344
	bmcdchWriteSlot(17, dwd1)
	dwd2 = bmcdchReadSlot(17)
	if dwd2 != 0x11223344 {
		t.Error("Expected 0x11223344, got ", dwd2)
	}

}

func TestBmcDchMapAddr(t *testing.T) {
	var addr1, addr2, page DgPhysAddrT
	bmcdchWriteSlot(0, 0)
	addr1 = 1
	addr2, page = getBmcDchMapAddr(addr1)
	if addr2 != 1 {
		t.Error("Expected 1, got ", addr2, page)
	}
	bmcdchWriteSlot(0, 3)
	addr1 = 1
	addr2, page = getBmcDchMapAddr(addr1)
	// 3 << 10 is 3072
	if addr2 != 3073 {
		t.Error("Expected 3073, got ", addr2, page)
	}
}
