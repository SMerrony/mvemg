// mvemg project main.go
package main

import (
	"bufio"
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
	breakpoints []dg_phys_addr
)

func main() {
	p = profile.Start(profile.ProfilePath("."))
	defer p.Stop()
	//	debugLogsInit()
	log.Println("INFO: MV/Em will not start until console connected")

	l, err := net.Listen("tcp", "localhost:"+SCP_PORT)
	if err != nil {
		log.Println("ERROR: Could not listen on console port: ", err.Error())
		os.Exit(1)
	}

	// close the port once we are done
	defer l.Close()
	//defer debugLogsDump()

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
		busInit()
		busAddDevice(DEV_SCP, "SCP", SCP_PMB, true, false, false)
		ttoInit(conn)
		ttiInit(conn)
		instructionsInit()
		cpuInit()
		mtbInit()
		dpfInit()
		dskpInit()

		// say hello...
		ttoPutChar(ASCII_FF)
		ttoPutStringNL(" *** Welcome to the MV/Emulator - Type HE for help ***")

		// the main SCP/console interaction loop
		for {
			ttoPutNLString("SCP-CLI> ")
			command := scpGetLine()
			log.Println("INFO: Got SCP command: " + command)
			doCommand(command)
		}
	}
}

// Get one line from the console - handle DASHER DELete key as corrector
func scpGetLine() string {
	line := []byte{}
	var cc byte
	for cc != ASCII_CR {
		cc = ttiGetChar()
		// handle the DASHER Delete key
		if cc == DASHER_DELETE && len(line) > 0 {
			ttoPutChar(DASHER_CURSOR_LEFT)
			line = line[:len(line)-1]
		} else {
			ttoPutChar(cc)
			line = append(line, cc)
		}
	}
	// we don't want the final CR
	line = line[:len(line)-1]

	return string(line[:])
}

// Exit cleanly, tidying up as much as we can
func cleanExit() {
	ttoPutNLString(" *** MV/Emulator stopping at user request ***")
	p.Stop()
	debugLogsDump()
	os.Exit(0)
}

func doCommand(cmd string) {
	words := strings.Split(strings.TrimSpace(cmd), " ")
	log.Printf("INFO: doCommand parsed command as <%s>\n", words[0])

	switch words[0] {
	// SCP-like commands
	case ".":
		ttoPutString(cpuPrintableStatus())
	case "B":
		boot(words)
	case "CO":
		run()
	case "E":
		ttoPutNLString(CMD_NYI)
	case "HE":
		showHelp()
	case "SS":
		singleStep()
	case "ST":
		ttoPutNLString(CMD_NYI)

	// emulator commands
	case "ATT":
		attach(words)
	case "BREAK":
		breakSet(words)
	case "CHECK":
		ttoPutStringNL(mtbScanImage(0))
	case "CREATE":
		createBlank(words)
	case "DIS":
		disassemble(words)
	case "DO":
		doScript(words)
	case "EXIT":
		cleanExit()
	case "NOBREAK":
		ttoPutNLString(CMD_NYI)
	case "SAVE":
		ttoPutNLString(CMD_NYI)
	case "SHOW":
		show(words)
	default:
		ttoPutNLString(CMD_UNKNOWN)
	}
}

/* Commands are below here... */

// Attach an image file to an emulated device
func attach(cmd []string) {
	if len(cmd) < 3 {
		ttoPutNLString(" *** ATT command requires arguments: <dev> and <image> ***")
		return
	}
	log.Printf("INFO: Attach called  with parms <%s> <%s>\n", cmd[1], cmd[2])
	switch cmd[1] {
	case "MTB":
		if mtbAttach(0, cmd[2]) {
			ttoPutNLString(" *** Tape Image Attached ***")
		} else {
			ttoPutNLString(" *** Could not ATTach Tape Image ***")
		}

	case "DPF":
		if dpfAttach(0, cmd[2]) {
			ttoPutNLString(" *** DPF Disk Image Attached ***")
		} else {
			ttoPutNLString(" *** Could not ATTach DPF Disk Image ***")
		}

	case "DSKP":
		if dskpAttach(0, cmd[2]) {
			ttoPutNLString(" *** DSKP Disk Image Attached ***")
		} else {
			ttoPutNLString(" *** Could not ATTach DSKP Disk Image ***")
		}

	default:
		ttoPutNLString(" *** Unknown or unimplemented Device for ATT command ***")
	}
}

func boot(cmd []string) {
	if len(cmd) != 2 {
		ttoPutNLString(" *** B command requires <devicenumber> ***")
		return
	}
	log.Printf("INFO: Boot called  with parm <%s>\n", cmd[1])
	dev, err := strconv.ParseInt(cmd[1], 8, 16) // FIXME Input Radix used here
	devNum := int(dev)
	if err != nil {
		ttoPutNLString(" *** Expecting <devicenumber> after B ***")
		return
	}
	if !busIsAttached(devNum) {
		ttoPutNLString(" *** Device is not ATTached ***")
		return
	}
	if !busIsBootable(devNum) {
		ttoPutNLString(" *** Device is not bootable ***")
		return
	}
	switch devNum {
	case DEV_MTB:
		mtbLoadTBoot(memory)
		cpu.ac[0] = DEV_MTB
		cpu.pc = 10
	default:
		ttoPutNLString(" *** Booting from that device not yet implemented ***")
	}
}

func breakSet(cmd []string) {
	if len(cmd) != 2 {
		ttoPutNLString(" *** BREAK command requires a single physical <address> argument ***")
		return
	}
	pAddr, err := strconv.Atoi(cmd[1])
	if err != nil {
		ttoPutNLString(" *** BREAK command could not parse <address> argument ***")
		return
	}
	breakpoints = append(breakpoints, dg_phys_addr(pAddr))
	ttoPutNLString("BREAKpoint set")
}

func createBlank(cmd []string) {
	if len(cmd) < 3 {
		ttoPutNLString(" *** Expecting DPF|DSKP <filename> args for CREATE command ***")
		return
	}
	switch cmd[1] {
	case "DPF":
		ttoPutNLString("Attempting to CREATE new empty DPF-type disk image, please wait...")
		if dpfCreateBlank(cmd[2]) {
			ttoPutNLString("Empty MV/Em DPF-type disk image created")
		} else {
			ttoPutNLString(" *** Error: could not create empty disk image ***")
		}
	case "DSKP":
		ttoPutNLString("Attempting to CREATE new empty DSKP-type disk image, please wait...")
		if dskpCreateBlank(cmd[2]) {
			ttoPutNLString("Empty MV/Em DSKP-type disk image created")
		} else {
			ttoPutNLString(" *** Error: could not create empty disk image ***")
		}
	default:
		ttoPutNLString(" *** CREATE not yet supported for that device type ***")
	}
}

func disassemble(cmd []string) {
	var (
		cmd1              = cmd[1]
		lowAddr, highAddr dg_phys_addr
		word              dg_word
		byte1, byte2      dg_byte
		display           string
		skipDecode        int
	)
	intVal1, err := strconv.Atoi(cmd1)
	if err != nil {
		ttoPutNLString(" *** Invalid address ***")
		return
	}
	if cmd1[0] == '+' {
		lowAddr = cpu.pc
		highAddr = lowAddr + dg_phys_addr(intVal1)
	} else {
		lowAddr = dg_phys_addr(intVal1)
		if len(cmd) == 2 {
			highAddr = lowAddr
		} else {
			intVal2, err := strconv.Atoi(cmd[2])
			if err != nil {
				ttoPutNLString(" *** Invalid address ***")
				return
			}
			highAddr = dg_phys_addr(intVal2)
		}
	}
	if highAddr < lowAddr {
		ttoPutNLString(" *** Invalid address range ***")
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
		ttoPutNLString(display)
	}
}

func doScript(cmd []string) {
	scriptFile, err := os.Open(cmd[1])
	if err != nil {
		ttoPutNLString(" *** Could not open MV/Em command script ***")
		log.Printf("WARN: Could not open MV/Em command script <%s>\n", cmd[1])
		return
	}
	defer scriptFile.Close()

	scanner := bufio.NewScanner(scriptFile)
	for scanner.Scan() {
		doCmd := scanner.Text()
		if doCmd[0] != '#' {
			// DebugLog.Printf("doScript read command <%s> from file\n", doCmd)
			ttoPutNLString(doCmd)
			doCommand(doCmd)
		}
	}

}

func printableBreakpointList() string {
	if len(breakpoints) == 0 {
		return " *** No BREAKpoints are set ***"
	}
	res := "BREAKpoints at: "
	for _, b := range breakpoints {
		res += fmt.Sprintf("%d. ", b)
	}
	return res
}

// Display SCP and Emulator help on the DASHER-compatible console
func showHelp() {
	ttoPutString("\014                          \024SCP-CLI Commands\025" +
		"                          \034MV/Emulator\035\012" +
		" .                      - Display state of CPU\012" +
		" B #                    - Boot from device #\012" +
		" CO                     - COntinue CPU Processing\012" +
		" E A <#> | M [addr] | P - Examine/Modify Acc/Memory/PC\012" +
		" HE                     - HElp (show this)\012" +
		" RE                     - REset the system\012" +
		" SS                     - Single Step one instruction\012" +
		" ST <addr>              - STart processing at specified address\012")
	ttoPutString("\012                          \024Emulator Commands\025\012" +
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
		ttoPutNLString(" *** SHOW requires argument ***")
		return
	}
	switch cmd[1] {
	case "DEV":
		ttoPutNLString(busGetPrintableDevList())
	case "BREAK":
		ttoPutNLString(printableBreakpointList())
	default:
		ttoPutNLString(" *** Invalid SHOW type ***")
	}
}

// Attempt to execute the opcode at PC
func singleStep() {
	ttoPutString(cpuPrintableStatus())
	// FETCH
	thisOp := memReadWord(cpu.pc)
	// DECODE
	if iPtr, ok := instructionDecode(thisOp, cpu.pc, cpu.sbr[cpu.pc>>29].lef, cpu.sbr[cpu.pc>>29].io, cpu.atu); ok {
		ttoPutNLString(iPtr.disassembly)
		// EXECUTE
		if cpuExecute(iPtr) {
			ttoPutString(cpuPrintableStatus())
		} else {
			ttoPutNLString(" *** Error: could not execute instruction")
		}
	} else {
		ttoPutNLString(" *** Error: could not decode opcode")
	}
}

// The main Emulator running loop...
func run() {
	var (
		thisOp    dg_word
		iPtr      *DecodedInstr
		ok        bool
		errDetail string
	)

	// direct console input to the VM
	ttiStartTask(&cpu)

	for {
		// FETCH
		thisOp = memReadWord(cpu.pc)

		// DECODE
		iPtr, ok = instructionDecode(thisOp, cpu.pc, cpu.sbr[cpu.pc>>29].lef, cpu.sbr[cpu.pc>>29].io, cpu.atu)
		if !ok {
			errDetail = " *** Error: could not decode instruction ***"
			break
		}
		log.Printf("%s\t\t%s\n", iPtr.disassembly, cpuCompactPrintableStatus())

		// EXECUTE
		if !cpuExecute(iPtr) {
			errDetail = " *** Error: could not execute instruction (or CPU HALT encountered) ***"
			break
		}

		// BREAKPOINT?
		if len(breakpoints) > 0 {
			for _, bAddr := range breakpoints {
				if bAddr == cpu.pc {
					ttiStopThread(&cpu)
					msg := fmt.Sprintf(" *** BREAKpoint hit at physical address %d. ***", cpu.pc)
					ttoPutNLString(msg)
					log.Println(msg)
					return
				}
			}
		}

		// Console ESCape?
		if cpu.consoleEsc {
			cpu.consoleEsc = false
			errDetail = " *** Console ESCape ***"
			break
		}
	}

	// run halted due to either error or console escape
	log.Println(errDetail)
	ttoPutNLString(errDetail)
	log.Printf("%s\n", cpuPrintableStatus())
	ttoPutString(cpuPrintableStatus())

	errDetail = " *** CPU halting ***"
	log.Println(errDetail)
	ttoPutNLString(errDetail)

	//debugLogsDump()

	errDetail = fmt.Sprintf(" *** MV/Em executed %d	 instructions ***", cpu.instrCount)
	log.Println(errDetail)
	ttoPutNLString(errDetail)

	ttiStopThread(&cpu)
}
