// eclipseOp.go
package main

import (
	//"fmt"
	"log"
)

func eclipseOp(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	var (
		addr, offset dg_phys_addr
		byt          dg_byte
		wd           dg_word
		dwd          dg_dword
		bitNum       uint
	)

	switch iPtr.mnemonic {

	case "ADI": // 16-bit unsigned Add Immediate
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if iPtr.immVal < 1 || iPtr.immVal > 4 {
			log.Fatal("Invalid immediate value in ADI")
		}
		wd += dg_word(iPtr.immVal) // unsigned arithmetic does wraparound in Go
		cpuPtr.ac[iPtr.acd] = dg_dword(wd)

	case "BTO":
		// TODO Handle segment and indirection...
		if iPtr.acd == iPtr.acs {
			addr = 0
		} else {
			addr = dg_phys_addr(cpuPtr.ac[iPtr.acs]) & 0x7fff // mask off lower 15 bits
		}
		offset = (dg_phys_addr(cpuPtr.ac[iPtr.acd]) & 0x0000fff0) >> 4
		addr += offset // add unsigned offset
		bitNum = uint(cpuPtr.ac[iPtr.acd] & 0x000f)
		wd = memReadWord(addr)
		DebugLog.Printf("... BTO Addr: %d, Bit: %d, Before: %s\n",
			addr, bitNum, wordToBinStr(wd))
		wd |= 1 << (15 - bitNum) // set the bit
		memWriteWord(addr, wd)
		DebugLog.Printf("... BTO                     Result: %s\n", wordToBinStr(wd))

	case "BTZ":
		// TODO Handle segment and indirection...
		if iPtr.acd == iPtr.acs {
			addr = 0
		} else {
			addr = dg_phys_addr(cpuPtr.ac[iPtr.acs]) & 0x7fff // mask off lower 15 bits
		}
		offset = (dg_phys_addr(cpuPtr.ac[iPtr.acd]) & 0x0000fff0) >> 4
		addr += offset // add unsigned offset
		bitNum = uint(cpuPtr.ac[iPtr.acd] & 0x000f)
		wd = memReadWord(addr)
		DebugLog.Printf("... BTZ Addr: %d, Bit: %d, Before: %s\n", addr, bitNum, wordToBinStr(wd))
		//wd |= 1 << (15 - bitNum) // set the bit
		if testWbit(wd, int(bitNum)) {
			wd ^= 1 << (15 - bitNum) // clear the bit
		}
		memWriteWord(addr, wd)
		DebugLog.Printf("... BTZ                     Result: %s\n",
			wordToBinStr(wd))

	case "DIV": // unsigned divide
		uw := dwordGetLowerWord(cpuPtr.ac[0])
		lw := dwordGetLowerWord(cpuPtr.ac[1])
		dwd = dg_dword(uw)<<16 | dg_dword(lw)
		quot := dwordGetLowerWord(cpuPtr.ac[2])
		if uw > quot || quot == 0 {
			cpuPtr.carry = true
		} else {
			cpuPtr.carry = false
			cpuPtr.ac[0] = (dwd % dg_dword(quot)) & 0x0ffff
			cpuPtr.ac[1] = (dwd / dg_dword(quot)) & 0x0ffff
		}

	case "DLSH":
		dplus1 := iPtr.acd + 1
		if dplus1 == 4 {
			dplus1 = 0
		}
		dwd = dlsh(cpuPtr.ac[iPtr.acs], cpuPtr.ac[iPtr.acd], cpuPtr.ac[dplus1])
		cpuPtr.ac[iPtr.acd] = dg_dword(dwordGetUpperWord(dwd))
		cpuPtr.ac[dplus1] = dg_dword(dwordGetLowerWord(dwd))

	case "ELEF":
		cpuPtr.ac[iPtr.acd] = dg_dword(resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp))

	case "ESTA":
		addr = resolve16bitEclipseAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp)
		memWriteWord(addr, dwordGetLowerWord(cpuPtr.ac[iPtr.acd]))

	case "HXL":
		dwd = cpuPtr.ac[iPtr.acd] << (uint32(iPtr.immVal) * 4)
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	case "HXR":
		dwd = cpuPtr.ac[iPtr.acd] >> (uint32(iPtr.immVal) * 4)
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	case "IOR":
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd]) | dwordGetLowerWord(cpuPtr.ac[iPtr.acs])
		cpuPtr.ac[iPtr.acd] = dg_dword(wd)

	case "IORI":
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd]) | dg_word(iPtr.immVal)
		cpuPtr.ac[iPtr.acd] = dg_dword(wd)

	case "LDB":
		byt = memReadByteEclipseBA(dwordGetLowerWord(cpuPtr.ac[iPtr.acs]))
		cpuPtr.ac[iPtr.acd] = dg_dword(byt)

	case "LSH":
		cpuPtr.ac[iPtr.acd] = lsh(cpuPtr.ac[iPtr.acs], cpuPtr.ac[iPtr.acd])

	case "MUL": // unsigned 16-bit multiply with add: (AC1 * AC2) + AC0 => AC0(h) and AC1(l)
		ac0 := dwordGetLowerWord(cpuPtr.ac[0])
		ac1 := dwordGetLowerWord(cpuPtr.ac[1])
		ac2 := dwordGetLowerWord(cpuPtr.ac[2])
		dwd := (dg_dword(ac1) * dg_dword(ac2)) + dg_dword(ac0)
		cpuPtr.ac[0] = dg_dword(dwordGetUpperWord(dwd))
		cpuPtr.ac[1] = dg_dword(dwordGetLowerWord(dwd))

	case "SBI": // unsigned
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		if iPtr.immVal < 1 || iPtr.immVal > 4 {
			log.Fatal("Invalid immediate value in SBI")
		}
		wd -= dg_word(iPtr.immVal)
		cpuPtr.ac[iPtr.acd] = dg_dword(wd)

	case "STB":
		hiLo := testDWbit(cpuPtr.ac[iPtr.acs], 31)
		addr = dg_phys_addr(dwordGetLowerWord(cpuPtr.ac[iPtr.acs])) >> 1
		byt = dg_byte(cpuPtr.ac[iPtr.acd])
		memWriteByte(addr, hiLo, byt)

	case "XCH":
		dwd = cpuPtr.ac[iPtr.acs]
		cpuPtr.ac[iPtr.acs] = cpuPtr.ac[iPtr.acd] & 0x0ffff
		cpuPtr.ac[iPtr.acd] = dwd & 0x0ffff

	default:
		log.Printf("ERROR: ECLIPSE_OP instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}

func dlsh(acS, acDh, acDl dg_dword) dg_dword {
	var shft int8 = int8(acS)
	var dwd dg_dword = ((acDh & 0x0ffff) << 16) | (acDl & 0x0ffff)
	if shft != 0 {
		if shft < -31 || shft > 31 {
			dwd = 0
		} else {
			if shft > 0 {
				dwd >>= uint8(shft)
			} else {
				shft *= -1
				dwd >>= uint8(shft)
			}
		}
	}
	return dwd
}

func lsh(acS, acD dg_dword) dg_dword {
	var shft int8 = int8(acS)
	var wd dg_word = dwordGetLowerWord(acD)
	if shft == 0 {
		wd = dwordGetLowerWord(acD) // do nothing
	} else {
		if shft < -15 || shft > 15 {
			wd = 0 // 16+ bit shift clears word
		} else {
			if shft > 0 {
				wd >>= uint8(shft)
			} else {
				shft *= -1
				wd >>= uint8(shft)
			}
		}
	}
	return dg_dword(wd)
}
