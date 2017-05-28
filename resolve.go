// resolve.go
package main

import (
	"fmt"
	//"log"
)

func resolve16bitEclipseAddr(cpuPtr *Cpu, ind byte, mode string, disp int32) dg_phys_addr {

	var (
		eff     dg_phys_addr
		intEff  int32
		indAddr dg_word
	)

	// handle addressing mode...
	switch mode {
	case "Absolute":
		intEff = disp
	case "PC":
		intEff = int32(cpuPtr.pc) + disp
	case "AC2":
		intEff = int32(cpuPtr.ac[2]) + disp
	case "AC3":
		intEff = int32(cpuPtr.ac[3]) + disp
	}

	// handle indirection
	if ind == '@' { // down the rabbit hole...
		indAddr = memReadWord(dg_phys_addr(intEff))
		for testWbit(indAddr, 0) {
			indAddr = memReadWord(dg_phys_addr(indAddr))
		}
		intEff = int32(indAddr)
	}

	// mask off to Eclipse range
	eff = dg_phys_addr(intEff) & 0x7fff

	debugPrint(SYSTEM_LOG, fmt.Sprintf("... resolve16bitEclipseAddr got: %d., returning %d.\n", disp, eff))
	return eff
}

// This is the same as resolve16bitEclipseAddr, but without the range masking at the end
func resolve16bitEagleAddr(cpuPtr *Cpu, ind byte, mode string, disp int32) dg_phys_addr {

	var (
		eff     dg_phys_addr
		intEff  int32
		indAddr dg_word
	)

	// handle addressing mode...
	switch mode {
	case "Absolute":
		intEff = disp
	case "PC":
		intEff = int32(cpuPtr.pc) + disp
	case "AC2":
		intEff = int32(cpuPtr.ac[2]) + disp
	case "AC3":
		intEff = int32(cpuPtr.ac[3]) + disp
	}

	// handle indirection
	if ind == '@' { // down the rabbit hole...
		indAddr = memReadWord(dg_phys_addr(intEff))
		for testWbit(indAddr, 0) {
			indAddr = memReadWord(dg_phys_addr(indAddr))
		}
		intEff = int32(indAddr)
	}

	eff = dg_phys_addr(intEff)

	debugPrint(SYSTEM_LOG, fmt.Sprintf("... resolve16bitEagleAddr got: %d., returning %d.\n", disp, eff))
	return eff
}

func resolve32bitEffAddr(cpuPtr *Cpu, ind byte, mode string, disp int32) dg_phys_addr {

	eff := dg_phys_addr(disp)

	// handle addressing mode...
	switch mode {
	case "Absolute":
		// nothing to do
	case "PC":
		eff += dg_phys_addr(cpuPtr.pc)
	case "AC2":
		eff += dg_phys_addr(cpuPtr.ac[2])
	case "AC3":
		eff += dg_phys_addr(cpuPtr.ac[3])
	}

	// handle indirection
	if ind == '@' { // down the rabbit hole...
		indAddr := memReadWord(eff)
		for testWbit(indAddr, 0) {
			indAddr = memReadWord(dg_phys_addr(indAddr))
		}
		eff = dg_phys_addr(indAddr)
	}

	debugPrint(SYSTEM_LOG, fmt.Sprintf("... resolve32bitEffAddr got: %d., returning %d.\n", disp, eff))
	return eff
}
