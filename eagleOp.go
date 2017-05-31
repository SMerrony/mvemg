// eagleOp.go
package main

import (
	"log"
)

func eagleOp(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	//var addr dg_phys_addr

	var (
		wd       dg_word
		dwd      dg_dword
		i32, res int32
	)

	switch iPtr.mnemonic {

	case "ADDI":
		// signed 16-bit add immediate
		i32 = int32(sexWordToDWord(dwordGetLowerWord(cpuPtr.ac[iPtr.acd])))
		i32 += int32(sexWordToDWord(iPtr.imm16b))
		cpuPtr.ac[iPtr.acd] = dg_dword(i32 & 0x0ffff)

	case "ANDI":
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		cpuPtr.ac[iPtr.acd] = dg_dword(wd&iPtr.imm16b) & 0x0000ffff

	case "CRYTC":
		cpuPtr.carry = !cpuPtr.carry

	case "CRYTO":
		cpuPtr.carry = true

	case "CRYTZ":
		cpuPtr.carry = false

	case "LLEF":
		cpuPtr.ac[iPtr.acd] = dg_dword(resolve32bitEffAddr(cpuPtr, iPtr.ind, iPtr.mode, iPtr.disp))

	case "NADD": // signed add
		res = int32(int16(cpuPtr.ac[iPtr.acd]) + int16(cpuPtr.ac[iPtr.acs]))
		cpuPtr.ac[iPtr.acd] = dg_dword(res)

	case "NLDAI":
		cpuPtr.ac[iPtr.acd] = sexWordToDWord(iPtr.imm16b)

	case "NSUB": // signed subtract
		res = int32(int16(cpuPtr.ac[iPtr.acd]) - int16(cpuPtr.ac[iPtr.acs]))
		cpuPtr.ac[iPtr.acd] = dg_dword(res)

	case "SSPT": /* NO-OP - see p.8-5 of MV/10000 Sys Func Chars */
		log.Println("INFO: SSPT is a No-Op on this machine, continuing")

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

	case "WANDI":
		cpuPtr.ac[iPtr.acd] &= iPtr.imm32b

	case "WCOM":
		cpuPtr.ac[iPtr.acd] ^= cpuPtr.ac[iPtr.acs]

	case "WINC":
		cpuPtr.ac[iPtr.acd] = cpuPtr.ac[iPtr.acs] + 1

	case "WIORI":
		cpuPtr.ac[iPtr.acd] |= iPtr.imm32b

	case "WLDAI":
		cpuPtr.ac[iPtr.acd] = iPtr.imm32b

	case "WLSHI":
		shiftAmt8 := int8(iPtr.imm16b & 0x0ff)
		if shiftAmt8 < 0 { // shift right
			shiftAmt8 *= -1
			dwd = cpuPtr.ac[iPtr.acd] >> uint(shiftAmt8)
			cpuPtr.ac[iPtr.acd] = dwd
		}
		if shiftAmt8 > 0 { // shift left
			dwd = cpuPtr.ac[iPtr.acd] << uint(shiftAmt8)
			cpuPtr.ac[iPtr.acd] = dwd
		}

	case "WMOV":
		cpuPtr.ac[iPtr.acd] = cpuPtr.ac[iPtr.acs]

	case "WNEG":
		cpuPtr.ac[iPtr.acd] = -cpuPtr.ac[iPtr.acs] // FIXME WNEG - handle CARRY/OVR

	case "WSUB":
		res = int32(cpuPtr.ac[iPtr.acd]) - int32(cpuPtr.ac[iPtr.acs])
		cpuPtr.ac[iPtr.acd] = dg_dword(res)
		// FIXME - handle overflow and carry

	case "WSBI":
		res = int32(cpuPtr.ac[iPtr.acd]) - iPtr.immVal
		cpuPtr.ac[iPtr.acd] = dg_dword(res)
		// FIXME - handle overflow and carry

	case "ZEX":
		cpuPtr.ac[iPtr.acd] = 0 | dg_dword(dwordGetLowerWord(cpuPtr.ac[iPtr.acs]))

	default:
		log.Printf("ERROR: EAGLE_OP instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
