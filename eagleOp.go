// eagleOp.go
package main

import (
	"log"
)

func eagleOp(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	//var addr dg_phys_addr

	var (
		wd  dg_word
		dwd dg_dword
		res int32
	)

	switch iPtr.mnemonic {

	case "ANDI":
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		cpuPtr.ac[iPtr.acd] = dg_dword(wd&iPtr.imm16b) & 0x0000ffff

	case "CRYTC":
		cpuPtr.carry = !cpuPtr.carry

	case "CRYTO":
		cpuPtr.carry = true

	case "CRYTZ":
		cpuPtr.carry = false

	case "NLDAI":
		cpuPtr.ac[iPtr.acd] = sexWordToDWord(iPtr.imm16b)

	case "WADD":
		res = int32(cpuPtr.ac[iPtr.acs]) + int32(cpuPtr.ac[iPtr.acd])
		cpuPtr.ac[iPtr.acd] = dg_dword(res)
		// FIXME - handle overflow and carry

	case "WADC":
		dwd = ^cpuPtr.ac[iPtr.acs]
		cpuPtr.ac[iPtr.acd] = dg_dword(int32(cpuPtr.ac[iPtr.acd]) + int32(dwd))
		// FIXME - handle overflow and carry

	case "WAND":
		cpuPtr.ac[iPtr.acd] &= cpuPtr.ac[iPtr.acs]

	case "WCOM":
		cpuPtr.ac[iPtr.acd] ^= cpuPtr.ac[iPtr.acs]

	case "WINC":
		cpuPtr.ac[iPtr.acd] = cpuPtr.ac[iPtr.acs] + 1

	case "WMOV":
		cpuPtr.ac[iPtr.acd] = cpuPtr.ac[iPtr.acs]

	case "WSUB":
		res = int32(cpuPtr.ac[iPtr.acd]) - int32(cpuPtr.ac[iPtr.acs])
		cpuPtr.ac[iPtr.acd] = dg_dword(res)
		// FIXME - handle overflow and carry
	default:
		log.Printf("ERROR: EAGLE_OP instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
