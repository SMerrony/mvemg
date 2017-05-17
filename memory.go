// memory
package main

import (
	"fmt"
	"log"
	"os"
	//"strconv"
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
	log.Printf("INFO: Initialised %d words of main memory\n", MEM_SIZE_WORDS)
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

// memWriteWord - ALL memory-writing should ultimately go through this function
// N.B. minor exceptions are made for nsPush and nsPop
func memWriteWord(wordAddr dg_phys_addr, datum dg_word) {
	if wordAddr >= MEM_SIZE_WORDS {
		log.Fatalf("ERROR: Attempt to write word beyond end of physical memory using address: %d", wordAddr)
		os.Exit(1)
	}
	memory.ram[wordAddr] = datum
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

func dwordGetLowerWord(dwd dg_dword) dg_word {
	return dg_word(dwd & 0x0000ffff)
}

func dwordGetUpperWord(dwd dg_dword) dg_word {
	return dg_word((dwd >> 16) & 0x0000ffff)
}

// in the DG world, the first (leftmost) bit is numbered zero...
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

// sign-extend a DG word to a DG DoubleWord
func sexWordToDWord(wd dg_word) dg_dword {
	var dwd dg_dword
	if testWbit(wd, 0) {
		dwd = dg_dword(wd) | 0xffff0000
	} else {
		dwd = dg_dword(wd)
	}
	return dwd
}

// swap over the two bytes in a dg_word
func swapBytes(wd dg_word) dg_word {
	var res dg_word
	res = (wd >> 8) | ((wd & 0x00ff) << 8)
	return res
}

// does word w have bit b set? //
func testWbit(w dg_word, b int) bool {
	bb := uint8(b)
	return (w & (1 << (15 - bb))) != 0
}

// get a pretty-printable string of a word
func wordToBinStr(w dg_word) string {
	return fmt.Sprintf("%08b %08b", w>>8, w&0x0ff)
}
