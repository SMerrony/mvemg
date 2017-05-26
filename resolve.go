// resolve.go
package main

import (
	"log"
)

func resolve16bitEclipseAddr(cpuPtr *Cpu, ind byte, mode string, disp int32) dg_phys_addr {

	eff := dg_phys_addr(disp) & 0x0ffff

	// handle addressing mode...
	switch mode {
	case "Absolute":
		// nothing to do
	case "PC":
		eff += dg_phys_addr(cpuPtr.pc) & 0x0ffff
	case "AC2":
		eff += dg_phys_addr(cpuPtr.ac[2]) & 0x0ffff
	case "AC3":
		eff += dg_phys_addr(cpuPtr.ac[3]) & 0x0ffff
	}

	// handle indirection
	if ind == '@' { // down the rabbit hole...
		indAddr := memReadWord(eff)
		for testWbit(indAddr, 0) {
			indAddr = memReadWord(dg_phys_addr(indAddr))
		}
		eff = dg_phys_addr(indAddr)
	}

	// mask off to Eclipse range
	eff &= 0x7fff

	log.Printf("... resolve16bitEclipseAddr got: %d., returning %d.\n", disp, eff)
	return eff
}

func resolve16bitEagleAddr(cpuPtr *Cpu, ind byte, mode string, disp int32) dg_phys_addr {

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

	log.Printf("... resolve16bitEclipseAddr got: %d., returning %d.\n", disp, eff)
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

	log.Printf("... resolve32bitEffAddr got: %d., returning %d.\n", disp, eff)

	return eff
}
