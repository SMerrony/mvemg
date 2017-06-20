// memory
package main

import (
	"fmt"
	"log"
	"os"
	//"strconv"
)

const (
	MEM_SIZE_WORDS = 8388608
	MEM_SIZE_LCPID = 0x3F

	// Some Page Zero special locations...

	WFP_LOC  = 020
	WSP_LOC  = 022
	WSL_LOC  = 024
	WSB_LOC  = 026
	WPFH_LOC = 030
	CBP_LOC  = 032

	NSP_LOC  = 040 // 32. Narrow Stack Pointer
	NFP_LOC  = 041
	NSL_LOC  = 042
	NSFA_LOC = 043
)

type Memory struct {
	ram                 [MEM_SIZE_WORDS]dg_word
	atuEnabled          bool
	pushCount, popCount int
}

var memory Memory

func memInit() {
	// zero ram?
	memory.atuEnabled = false
	bmcdchInit()
	debugPrint(debugLog, "INFO: Initialised %d words of main memory\n", MEM_SIZE_WORDS)
}

// read a byte from memory using word address and low-byte flag (true => lower (rightmost) byte)
func memReadByte(wordAddr dg_phys_addr, loByte bool) dg_byte {
	if wordAddr >= MEM_SIZE_WORDS {
		log.Fatalf("ERROR: Attempt to read byte beyond end of physical memory using address: %d", wordAddr)
		os.Exit(1)
	}
	var res dg_byte
	wd := memory.ram[wordAddr]
	if loByte {
		res = dg_byte(wd & 0xff)
	} else {
		res = dg_byte(wd >> 8)
	}
	return res
}

func memReadByteEclipseBA(byteAddr16 dg_word) dg_byte {
	var (
		hiLo bool
		addr dg_phys_addr
	)
	hiLo = testWbit(byteAddr16, 15) // determine which byte to get
	addr = dg_phys_addr(byteAddr16) >> 1
	return memReadByte(addr, hiLo)
}

func memWriteByte(wordAddr dg_phys_addr, loByte bool, b dg_byte) {
	wd := memory.ram[wordAddr]
	if loByte {
		wd = (wd & 0xff00) | dg_word(b)
	} else {
		wd = dg_word(b)<<8 | (wd & 0x00ff)
	}
	memWriteWord(wordAddr, wd)
}

func memReadWord(wordAddr dg_phys_addr) dg_word {
	if wordAddr >= MEM_SIZE_WORDS {
		log.Fatalf("ERROR: Attempt to read word beyond end of physical memory using address: %d", wordAddr)
		os.Exit(1)
	}
	return memory.ram[wordAddr]
}

func memReadWordDchChan(addr dg_phys_addr) dg_word {
	pAddr := addr
	if getDchMode() {
		pAddr, _ = getBmcDchMapAddr(addr)
	}
	debugPrint(mapLog, "memReadWordBmcChan got addr: %d, read from addr: %d\n", addr, pAddr)
	return memReadWord(pAddr)
}

func memReadWordBmcChan(addr dg_phys_addr) dg_word {
	var pAddr dg_phys_addr
	decodedAddr := decodeBmcAddr(addr)
	if decodedAddr.isLogical {
		pAddr, _ = getBmcDchMapAddr(addr) // FIXME
	} else {
		pAddr = decodedAddr.ca
	}
	debugPrint(mapLog, "memWriteReadBmcChan got addr: %d, wrote to addr: %d\n", addr, pAddr)
	return memReadWord(pAddr)
}

// memWriteWord - ALL memory-writing should ultimately go through this function
// N.B. minor exceptions may be made for nsPush() and nsPop()
func memWriteWord(wordAddr dg_phys_addr, datum dg_word) {
	if wordAddr >= MEM_SIZE_WORDS {
		log.Fatalf("ERROR: Attempt to write word beyond end of physical memory using address: %d", wordAddr)
		os.Exit(1)
	}
	memory.ram[wordAddr] = datum
}

func memWriteWordDchChan(addr dg_phys_addr, data dg_word) dg_phys_addr {
	pAddr := addr

	if getDchMode() {
		pAddr, _ = getBmcDchMapAddr(addr)
	}
	memWriteWord(pAddr, data)
	debugPrint(mapLog, "memWriteWordDchChan got addr: %d, wrote to addr: %d\n", addr, pAddr)
	return pAddr
}

func memWriteWordBmcChan(addr dg_phys_addr, data dg_word) dg_phys_addr {
	var pAddr dg_phys_addr
	decodedAddr := decodeBmcAddr(addr)
	if decodedAddr.isLogical {
		pAddr, _ = getBmcDchMapAddr(addr) // FIXME
	} else {
		pAddr = decodedAddr.ca
	}
	memWriteWord(pAddr, data)
	debugPrint(mapLog, "memWriteWordBmcChan got addr: %d, wrote to addr: %d\n", addr, pAddr)
	return pAddr
}

func memReadDWord(wordAddr dg_phys_addr) dg_dword {
	if wordAddr >= MEM_SIZE_WORDS {
		log.Fatalf("ERROR: Attempt to read doubleword beyond end of physical memory using address: %d", wordAddr)
		os.Exit(1)
	}
	var dword dg_dword
	dword = dg_dword(memory.ram[wordAddr]) << 16
	dword = dword | dg_dword(memory.ram[wordAddr+1])
	return dword
}

func memWriteDWord(wordAddr dg_phys_addr, dwd dg_dword) {
	if wordAddr >= MEM_SIZE_WORDS {
		log.Fatalf("ERROR: Attempt to write doubleword beyond end of physical memory using address: %d", wordAddr)
		os.Exit(1)
	}
	memWriteWord(wordAddr, dwordGetUpperWord(dwd))
	memWriteWord(wordAddr+1, dwordGetLowerWord(dwd))
}

// PUSH a word onto the Narrow Stack
func nsPush(seg dg_phys_addr, data dg_word) {
	// TODO segment handling
	// TODO overflow/underflow handling - either here or in instruction?
	memory.ram[NSP_LOC]++ // we allow this direct write to a fixed location for performance
	addr := dg_phys_addr(memory.ram[NSP_LOC])
	memWriteWord(addr, data)
	debugPrint(debugLog, "nsPush pushed %8d onto the Narrow Stack at location: %d\n", data, addr)
}

// POP a word off the Narrow Stack
func nsPop(seg dg_phys_addr) dg_word {
	// TODO segment handling
	// TODO overflow/underflow handling - either here or in instruction?
	addr := dg_phys_addr(memory.ram[NSP_LOC])
	data := memReadWord(addr)
	debugPrint(debugLog, "nsPop  popped %8d off  the Narrow Stack at location: %d\n", data, addr)
	memory.ram[NSP_LOC]-- // we allow this direct write to a fixed location for performance
	return data
}

// PUSH a doubleword onto the Wide Stack
func wsPush(seg dg_phys_addr, data dg_dword) {
	// TODO segment handling
	// TODO overflow/underflow handling - either here or in instruction?
	memory.ram[WSP_LOC] += 2 // we allow this direct write to a fixed location for performance
	addr := dg_phys_addr(memory.ram[WSP_LOC])
	memWriteDWord(addr, data)
	debugPrint(debugLog, "wsPush pushed %8d onto the Wide Stack at location: %d\n", data, addr)
}

// POP a word off the Wide Stack
func wsPop(seg dg_phys_addr) dg_dword {
	// TODO segment handling
	// TODO overflow/underflow handling - either here or in instruction?
	addr := dg_phys_addr(memory.ram[WSP_LOC])
	dword := memReadDWord(addr)
	memory.ram[WSP_LOC] -= 2 // we allow this direct write to a fixed location for performance
	debugPrint(debugLog, "wsPop  popped %8d off  the Wide Stack at location: %d\n", dword, addr)
	return dword
}

// dwordGetLowerWord gets the DG-lower word of a doubleword
// Called VERY often, hopefully inlined!
func dwordGetLowerWord(dwd dg_dword) dg_word {
	return dg_word(dwd) & 0x0000ffff
}

func dwordGetUpperWord(dwd dg_dword) dg_word {
	return dg_word((dwd >> 16) & 0x0000ffff)
}

// in the DG world, the first (leftmost) bit is numbered zero...
// extract nbits from value starting at leftBit
func getWbits(value dg_word, leftBit int, nbits int) dg_word {
	var res dg_word = 0
	rightBit := leftBit + nbits
	for b := leftBit; b < rightBit; b++ {
		res = res << 1
		if testWbit(value, b) {
			res++
		}
	}
	return res
}

// in the DG world, the first (leftmost) bit is numbered zero...
// extract nbits from value starting at leftBit
func getDWbits(value dg_dword, leftBit int, nbits int) dg_dword {
	var res dg_dword = 0
	rightBit := leftBit + nbits
	for b := leftBit; b < rightBit; b++ {
		res = res << 1
		if testDWbit(value, b) {
			res++
		}
	}
	return res
}

// sign-extend a DG word to a DG DoubleWord
func sexWordToDWord(wd dg_word) dg_dword {
	var dwd dg_dword
	if testWbit(wd, 0) {
		dwd = dg_dword(wd) | 0xffff0000
	} else {
		dwd = dg_dword(wd) & 0x0000ffff
	}
	return dwd
}

// swap over the two bytes in a dg_word
func swapBytes(wd dg_word) dg_word {
	var res dg_word
	res = (wd >> 8) | ((wd & 0x00ff) << 8)
	return res
}

// does word w have bit b set?
func testWbit(w dg_word, b int) bool {
	bb := uint8(b)
	return (w & (1 << (15 - bb))) != 0
}

// does dword dw have bit b set?
func testDWbit(dw dg_dword, b int) bool {
	bb := uint8(b)
	return ((dw & (1 << (31 - bb))) != 0)
}

// get a pretty-printable string of a word
func wordToBinStr(w dg_word) string {
	return fmt.Sprintf("%08b %08b", w>>8, w&0x0ff)
}
