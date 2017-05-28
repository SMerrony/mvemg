// eclipseOp.go
package main

import (
	"log"
)

func eclipseOp(cpuPtr *Cpu, iPtr *DecodedInstr) bool {
	var (
		addr dg_phys_addr
		byt  dg_byte
		wd   dg_word
		dwd  dg_dword
	)

	switch iPtr.mnemonic {

	case "ADI": // unsigned
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		wd += dg_word(iPtr.immVal) // TODO this is a signed int, is this OK?
		cpuPtr.ac[iPtr.acd] = dg_dword(wd)

	case "DIV": // unsigned divide
		uw := dwordGetLowerWord(cpuPtr.ac[0])
		lw := dwordGetLowerWord(cpuPtr.ac[1])
		dwd = dg_dword(uw)<<16 | dg_dword(lw)
		quot := dwordGetLowerWord(cpuPtr.ac[2])
		if uw > quot || quot == 0 {
			cpuPtr.carry = true
		} else {
			cpuPtr.ac[0] = dwd % dg_dword(quot)
			cpuPtr.ac[1] = dwd / dg_dword(quot)
		}

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

	case "LDB":
		byt = memReadByteEclipseBA(dwordGetLowerWord(cpuPtr.ac[iPtr.acs]))
		cpuPtr.ac[iPtr.acd] = dg_dword(byt)

	case "SBI": // unsigned
		wd = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
		wd -= dg_word(iPtr.immVal) // TODO this is a signed int, is this OK?
		cpuPtr.ac[iPtr.acd] = dg_dword(wd)

	case "STB":
		hiLo := testDWbit(cpuPtr.ac[iPtr.acs], 31)
		addr = dg_phys_addr(dwordGetLowerWord(cpuPtr.ac[iPtr.acs])) >> 1
		byt = dg_byte(cpuPtr.ac[iPtr.acd])
		memWriteByte(addr, hiLo, byt)

	case "XCH":
		dwd = cpuPtr.ac[iPtr.acs]
		cpuPtr.ac[iPtr.acs] = cpuPtr.ac[iPtr.acd]
		cpuPtr.ac[iPtr.acd] = dwd

	default:
		log.Printf("ERROR: ECLIPSE_OP instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}

	cpuPtr.pc += dg_phys_addr(iPtr.instrLength)
	return true
}
