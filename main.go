// mvemg project main.go
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	//"mvemg/tto"
)

import "github.com/pkg/profile"

const (
	SCP_PORT    = "10000"
	SCP_BUFSIZE = 135

	CMD_UNKNOWN = " *** UNKNOWN SCP-CLI COMMAND ***"
	CMD_NYI     = "Command Not Yet Implemented"
)

type (
	// a DG Word is 16-bit unsigned
	dg_word uint16
	// a DG Double-Word is 32-bit unsigned
	dg_dword uint32
	// a DG Byte is 8-bit unsigned
	dg_byte byte
	// a physical address is 32-bit unsigned
	dg_phys_addr uint32
)

var p interface {
	Stop()
}

var (
	tto Tto
	tti Tti
	bus Devices
	mtb Mtb
	cpu Cpu
)

func main() {
	p = profile.Start(profile.ProfilePath("."))
	defer p.Stop()
	log.Println("INFO: MV/Em will not start until console connected")

	l, err := net.Listen("tcp", "localhost:"+SCP_PORT)
	if err != nil {
		log.Println("ERROR: Could not listen on console port: ", err.Error())
		os.Exit(1)
	}

	// close the port once we are done
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("ERROR: Could not accept on console port: ", err.Error())
			os.Exit(1)
		}

		/***
		 *  The console is connected, now we can set up our emulated machine
		 *
		 * Here we are defining the hardware in our virtual machine
		 * Initially based on a minimally configured MV/10000 Model I.
		 *
		 *   One CPU
		 *   Console (TTI/TTO)
		 *   One Tape Drive
		 *   One HDD
		 *   A generous(!) 16MB RAM
		 *   NO IACs, LPT or ISC
		 ***/

		memInit()
		bus.busInit()
		bus.busAddDevice(DEV_SCP, "SCP", SCP_PMB, true, false, false)
		tto = Tto{conn}
		tti = Tti{conn}
		instructionsInit()
		cpu.cpuInit()
		mtb.mtbInit()

		// say hello...
		tto.ttoPutChar(ASCII_FF)
		tto.ttoPutStringNL(" *** Welcome to the MV/Emulator - Type HE for help ***")

		// the main SCP/console interaction loop
		for {
			tto.ttoPutNLString("SCP-CLI> ")
			command := scpGetLine(&tti, &tto)
			log.Println("Got SCP command: " + command)
			doCommand(command)
		}
	}
}

// Get one line from the console - handle DASHER DELete key as corrector
func scpGetLine(ti *Tti, to *Tto) string {
	line := []byte{}
	var cc byte
	for cc != ASCII_CR {
		cc = ti.ttiGetChar()
		// handle the DASHER Delete key
		if cc == DASHER_DELETE && len(line) > 0 {
			to.ttoPutChar(DASHER_CURSOR_LEFT)
			line = line[:len(line)-1]
		} else {
			to.ttoPutChar(cc)
			line = append(line, cc)
		}
	}
	// we don't want the final CR
	line = line[:len(line)-1]

	return string(line[:])
}

// Exit cleanly, tidying up as much as we can
func cleanExit() {
	tto.ttoPutNLString(" *** MV/Emulator stopping at user request ***")
	p.Stop()
	os.Exit(0)
}

func doCommand(cmd string) {
	words := strings.Split(strings.TrimSpace(cmd), " ")
	log.Printf("doCommand parsed command as <%s>\n", words[0])

	switch words[0] {
	// SCP-like commands
	case ".":
		tto.ttoPutString(cpu.cpuPrintableStatus())
	case "B":
		boot(words)
	case "CO":
		tto.ttoPutNLString(CMD_NYI)
	case "E":
		tto.ttoPutNLString(CMD_NYI)
	case "HE":
		showHelp()
	case "SS":
		singleStep()
	case "ST":
		tto.ttoPutNLString(CMD_NYI)

	// emulator commands
	case "ATT":
		attach(words)
	case "BREAK":
		tto.ttoPutNLString(CMD_NYI)
	case "CHECK":
		tto.ttoPutStringNL(mtb.mtbScanImage(0))
	case "CREATE":
		tto.ttoPutNLString(CMD_NYI)
	case "DIS":
		disassemble(words)
	case "DO":
		tto.ttoPutNLString(CMD_NYI)
	case "EXIT":
		cleanExit()
	case "NOBREAK":
		tto.ttoPutNLString(CMD_NYI)
	case "SAVE":
		tto.ttoPutNLString(CMD_NYI)
	case "SHOW":
		show(words)
	default:
		tto.ttoPutNLString(CMD_UNKNOWN)
	}
}

// Attach an image file to an emulated device
func attach(cmd []string) {
	if len(cmd) < 3 {
		tto.ttoPutNLString(" *** ATT command requires arguments: <dev> and <image> ***")
		return
	}
	log.Printf("INFO: Attach called  with parms <%s> <%s>\n", cmd[1], cmd[2])
	switch cmd[1] {
	case "MTB":
		if mtb.mtbAttach(0, cmd[2]) {
			tto.ttoPutNLString(" *** Tape Image Attached ***")
		} else {
			tto.ttoPutNLString(" *** Could not ATTach Tape Image ***")
		}

	default:
		tto.ttoPutNLString(" *** Unknown or unimplemented Device for ATT command ***")
	}
}

func boot(cmd []string) {
	if len(cmd) != 2 {
		tto.ttoPutNLString(" *** B command requires <devicenumber> ***")
		return
	}
	log.Printf("INFO: Boot called  with parm <%s>\n", cmd[1])
	dev, err := strconv.ParseInt(cmd[1], 8, 16) // FIXME Input Radix used here
	devNum := int(dev)
	if err != nil {
		tto.ttoPutNLString(" *** Expecting <devicenumber> after B ***")
		return
	}
	if !bus.busIsAttached(devNum) {
		tto.ttoPutNLString(" *** Device is not ATTached ***")
		return
	}
	if !bus.busIsBootable(devNum) {
		tto.ttoPutNLString(" *** Device is not bootable ***")
		return
	}
	switch devNum {
	case DEV_MTB:
		mtb.mtbLoadTBoot(memory)
		cpu.ac[0] = DEV_MTB
		cpu.pc = 10
	default:
		tto.ttoPutNLString(" *** Booting from that device not yet implemented ***")
	}
}

func disassemble(cmd []string) {
	var (
		lowAddr, highAddr dg_phys_addr
		word              dg_word
		byte1, byte2      dg_byte
		display           string
		skipDecode        int
	)
	intVal1, err := strconv.Atoi(cmd[1])
	if err != nil {
		tto.ttoPutNLString(" *** Invalid address ***")
		return
	}
	if cmd[1][:0] == "+" {
		lowAddr = cpu.pc
		highAddr = lowAddr + dg_phys_addr(intVal1)
	} else {
		lowAddr = dg_phys_addr(intVal1)
		if len(cmd) == 2 {
			highAddr = lowAddr
		} else {
			intVal2, err := strconv.Atoi(cmd[2])
			if err != nil {
				tto.ttoPutNLString(" *** Invalid address ***")
				return
			}
			highAddr = dg_phys_addr(intVal2)
		}
	}
	if highAddr < lowAddr {
		tto.ttoPutNLString(" *** Invalid address range ***")
		return
	}
	for addr := lowAddr; addr <= highAddr; addr++ {
		word = memReadWord(addr)
		byte1 = dg_byte(word >> 8)
		byte2 = dg_byte(word & 0x00ff)
		display = fmt.Sprintf("%09d: %02X %02X %03o %03o %s \"", addr, byte1, byte2, byte1, byte2, wordToBinStr(word))
		if byte1 >= ' ' && byte1 <= '~' {
			display += string(byte1)
		} else {
			display += " "
		}
		if byte2 >= ' ' && byte2 <= '~' {
			display += string(byte2)
		} else {
			display += " "
		}
		display += "\" "
		if skipDecode == 0 {
			instrTmp, _ := instructionDecode(word, addr, cpu.sbr[cpu.pc>>29].lef, cpu.sbr[cpu.pc>>29].io, cpu.atu)
			display += instrTmp.disassembly
			if instrTmp.instrLength > 1 {
				skipDecode = instrTmp.instrLength - 1
			}
		} else {
			skipDecode--
		}
		tto.ttoPutNLString(display)
	}
}

// Display SCP and Emulator help on the DASHER-compatible console
func showHelp() {
	tto.ttoPutString("\014                          \024SCP-CLI Commands\025" +
		"                          \034MV/Emulator\035\012" +
		" .                      - Display state of CPU\012" +
		" B #                    - Boot from device #\012" +
		" CO                     - COntinue CPU Processing\012" +
		" E A <#> | M [addr] | P - Examine/Modify Acc/Memory/PC\012" +
		" HE                     - HElp (show this)\012" +
		" RE                     - REset the system\012" +
		" SS                     - Single Step one instruction\012" +
		" ST <addr>              - STart processing at specified address\012")
	tto.ttoPutString("\012                          \024Emulator Commands\025\012" +
		" ATT <dev> <file> [RW]  - ATTach the image file to named device (RO)\012" +
		" BREAK/NOBREAK <addr>   - Set or clear a BREAKpoint\012" +
		" CHECK                  - CHECK validity of attached TAPE image\012" +
		" CREATE DPF|DSKP <file> - CREATE an empty/unformatted disk image\012" +
		" DET <dev>              - DETach any image file from the device\012" +
		" DIS <from> <to>|+<#>   - DISassemble physical memory range or # from PC\012" +
		" DO <file>              - Run (DO) emulator commands from <file>\012" +
		" EXIT                   - EXIT the emulator\012" +
		" GET/SAVE <file>        - GET (restore)/SAVE emulator state from/to file\012" +
		" SHOW BREAK/DEV         - SHOW list of BREAKpoints/DEVices configured\012")
}

// Show various emulator states to the user
func show(cmd []string) {
	if len(cmd) == 1 {
		tto.ttoPutNLString(" *** SHOW requires argument ***")
		return
	}
	switch cmd[1] {
	case "DEV":
		tto.ttoPutNLString(bus.busGetPrintableDevList())
	default:
		tto.ttoPutNLString(" *** Invalid SHOW type ***")
	}
}

// Attempt to execute the opcode at PC
func singleStep() {
	tto.ttoPutString(cpu.cpuPrintableStatus())
	thisOp := memReadWord(cpu.pc)
	if iPtr, ok := instructionDecode(thisOp, cpu.pc, cpu.sbr[cpu.pc>>29].lef, cpu.sbr[cpu.pc>>29].io, cpu.atu); ok {
		tto.ttoPutNLString(iPtr.disassembly)
		if cpu.cpuExecute(iPtr) {
			tto.ttoPutString(cpu.cpuPrintableStatus())
		} else {
			tto.ttoPutNLString(" *** Error: could not execute instruction")
		}
	} else {
		tto.ttoPutNLString(" *** Error: could not decode opcode")
	}
}
