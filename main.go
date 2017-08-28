// mvemg project main.go

// Copyright (C) 2017  Steve Merrony

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"bufio"
	"fmt"
	"log"
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
	"mvemg/util"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

// import "github.com/pkg/profile"

const (
	// StatPort is the port for the real-time status monitor
	StatPort = "9999"
	// ScpPort is  the port for the SCP master console
	ScpPort = "10000"
	// ScpBuffSize is the char buffer length for SCP input lines
	ScpBuffSize = 135

	cmdUnknown = " *** UNKNOWN SCP-CLI COMMAND ***"
	cmdNYI     = "Command Not Yet Implemented"
)

var p interface {
	Stop()
}

var (
	// debugLogging - CPU runs about 3x faster without debugLogging
	// (and another 3x faster without disassembly, linked to this)
	debugLogging  = true
	breakpoints   []dg.PhysAddrT
	cpuStatsChan  chan cpuStatT
	dpfStatsChan  chan DpfStatT
	dskpStatsChan chan dskpStatT
	ttiSCPchan    chan byte
)

func main() {
	//p = profile.Start(profile.ProfilePath("."))
	//defer p.Stop()
	//	debugLogsInit()
	log.Println("INFO: MV/Em will not start until console connected")

	l, err := net.Listen("tcp", "localhost:"+ScpPort)
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

		// create the channels used for near-real-time status monitoring
		// See statusCollector.go for details
		cpuStatsChan = make(chan cpuStatT, 3)
		dpfStatsChan = make(chan DpfStatT, 3)
		dskpStatsChan = make(chan dskpStatT, 3)

		ttiSCPchan = make(chan byte, ScpBuffSize)

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

		memory.MemInit()
		busInit()
		busAddDevice(DEV_SCP, "SCP", SCP_PMB, true, false, false)
		instructionsInit()
		decoderGenAllPossOpcodes()
		cpuPtr := cpuInit(cpuStatsChan)
		ttoInit(conn)
		ttiInit(conn, cpuPtr, ttiSCPchan)
		mtbInit()
		DpfInit(dpfStatsChan)
		dskpInit(dskpStatsChan)

		// say hello...
		ttoPutChar(ASCII_FF)
		ttoPutStringNL(" *** Welcome to the MV/Emulator - Type HE for help ***")

		// kick off the status monitor routine
		go statusCollector(cpuStatsChan, dpfStatsChan, dskpStatsChan)

		// run any command specified on the command line
		args := os.Args[1:]
		if len(args) == 1 {
			command := fmt.Sprintf("DO %s", args[0])
			log.Printf("INFO: got startup command <%s>\n", command)
			doCommand(command) // N.B. will not pass here until start-up script is complete...
		}

		// the main SCP/console interaction loop
		cpu.cpuMu.Lock()
		cpu.scpIO = true
		cpu.cpuMu.Unlock()
		for {
			ttoPutNLString("SCP-CLI> ")
			command := scpGetLine()
			//log.Println("INFO: Got SCP command: " + command)
			doCommand(command)
		}
	}
}

// Get one line from the console - handle DASHER DELete key as corrector
func scpGetLine() string {
	line := []byte{}
	var cc byte
	for cc != ASCII_CR {
		cc = <-ttiSCPchan
		//cc = ttiGetChar()
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
	//p.Stop()
	logging.DebugLogsDump()
	os.Exit(0)
}

func doCommand(cmd string) {
	words := strings.Split(strings.TrimSpace(cmd), " ")
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "INFO: doCommand parsed command as <%s>\n", words[0])
	}

	switch words[0] {
	// SCP-like commands
	case ".":
		ttoPutString(cpuPrintableStatus())
	case "B":
		boot(words)
	case "CO":
		run()
	case "E":
		examine(words)
	case "HE":
		showHelp()
	case "SS":
		singleStep()
	case "ST":
		start(words)

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
	// for convenience...
	case "exit":
		fallthrough
	case "EXIT":
		cleanExit()
	case "NOBREAK":
		ttoPutNLString(cmdNYI)
	case "SAVE":
		ttoPutNLString(cmdNYI)
	case "SHOW":
		show(words)
	default:
		ttoPutNLString(cmdUnknown)
	}
}

/* Commands are below here... */

// Attach an image file to an emulated device
func attach(cmd []string) {
	if len(cmd) < 3 {
		ttoPutNLString(" *** ATT command requires arguments: <dev> and <image> ***")
		return
	}
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "INFO: Attach called  with parms <%s> <%s>\n", cmd[1], cmd[2])
	}
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
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "INFO: Boot called  with parm <%s>\n", cmd[1])
	}
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
		mtbLoadTBoot()
		cpu.cpuMu.Lock()
		cpu.ac[0] = DEV_MTB
		cpu.pc = 10
		cpu.cpuMu.Unlock()
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
	breakpoints = append(breakpoints, dg.PhysAddrT(pAddr))
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
		lowAddr, highAddr dg.PhysAddrT
		word              dg.WordT
		byte1, byte2      dg.ByteT
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
		highAddr = lowAddr + dg.PhysAddrT(intVal1)
	} else {
		lowAddr = dg.PhysAddrT(intVal1)
		if len(cmd) == 2 {
			highAddr = lowAddr
		} else {
			intVal2, err := strconv.Atoi(cmd[2])
			if err != nil {
				ttoPutNLString(" *** Invalid address ***")
				return
			}
			highAddr = dg.PhysAddrT(intVal2)
		}
	}
	if highAddr < lowAddr {
		ttoPutNLString(" *** Invalid address range ***")
		return
	}
	for addr := lowAddr; addr <= highAddr; addr++ {
		word = memory.ReadWord(addr)
		byte1 = dg.ByteT(word >> 8)
		byte2 = dg.ByteT(word & 0x00ff)
		display = fmt.Sprintf("%09d: %02X %02X %03o %03o %s \"", addr, byte1, byte2, byte1, byte2,
			util.WordToBinStr(word))
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
			instrTmp, _ := instructionDecode(word, addr, cpu.sbr[cpu.pc>>29].lef, cpu.sbr[cpu.pc>>29].io, cpu.atu, true)
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
		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "WARN: Could not open MV/Em command script <%s>\n", cmd[1])
		}
		return
	}
	defer scriptFile.Close()

	scanner := bufio.NewScanner(scriptFile)
	for scanner.Scan() {
		doCmd := scanner.Text()
		if doCmd[0] != '#' {
			ttoPutNLString(doCmd)
			doCommand(doCmd)
		}
	}

}

// examine mimics the E command from later SCP-CLIs
func examine(cmd []string) {
	switch cmd[1] {
	case "A":
		exAc, err := strconv.Atoi(cmd[2])
		if err != nil || exAc < 0 || exAc > 3 {
			ttoPutNLString(" *** Examine Accumulator - invalid AC number ***")
			return
		}
		prompt := fmt.Sprintf("AC%d = %d - Enter new val or just ENTER> ", exAc, cpu.ac[exAc])
		ttoPutNLString(prompt)
		resp := scpGetLine()
		if len(resp) > 0 {
			newVal, err := strconv.Atoi(resp)
			if err != nil {
				ttoPutNLString(" *** Could not parse new AC value ***")
				return
			}
			cpu.ac[exAc] = dg.DwordT(newVal)
			prompt = fmt.Sprintf("AC%d = %d", exAc, cpu.ac[exAc])
			ttoPutNLString(prompt)
		}
	case "M":
		exMem, err := strconv.Atoi(cmd[2])
		if err != nil || exMem < 0 || exMem > memory.MemSizeWords {
			ttoPutNLString(" *** Examine Memory - invalid address ***")
			return
		}
		prompt := fmt.Sprintf("Location %d contains %d - Enter new val or just ENTER> ", exMem, memory.ReadWord(dg.PhysAddrT(exMem)))
		ttoPutNLString(prompt)
		resp := scpGetLine()
		if len(resp) > 0 {
			newVal, err := strconv.Atoi(resp)
			if err != nil {
				ttoPutNLString(" *** Could not parse new value ***")
				return
			}
			memory.WriteWord(dg.PhysAddrT(exMem), dg.WordT(newVal))
			prompt = fmt.Sprintf("Location %d = %d", exMem, memory.ReadWord(dg.PhysAddrT(exMem)))
			ttoPutNLString(prompt)
		}
	case "P":
		prompt := fmt.Sprintf("PC = %d - Enter new val or just ENTER> ", cpu.pc)
		ttoPutNLString(prompt)
		resp := scpGetLine()
		if len(resp) > 0 {
			newVal, err := strconv.Atoi(resp)
			if err != nil {
				ttoPutNLString(" *** Could not parse new PC value ***")
				return
			}
			cpu.pc = dg.PhysAddrT(newVal)
			prompt = fmt.Sprintf("PC = %d", cpu.pc)
			ttoPutNLString(prompt)
		}
	default:
		ttoPutNLString(" *** Expecting A, M, or P for E(xamine) command ***")
		return
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
	thisOp := memory.ReadWord(cpu.pc)
	// DECODE
	if iPtr, ok := instructionDecode(thisOp, cpu.pc, cpu.sbr[cpu.pc>>29].lef, cpu.sbr[cpu.pc>>29].io, cpu.atu, true); ok {
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

// start running at user-provided PC
func start(cmd []string) {
	newPc, err := strconv.Atoi(cmd[1])
	if err != nil || newPc < 0 {
		ttoPutNLString(" *** Could not parse new PC value ***")
		return
	}
	cpu.pc = dg.PhysAddrT(newPc)
	run()
}

// The main Emulator running loop...
func run() {
	var (
		thisOp    dg.WordT
		iPtr      *decodedInstrT
		ok        bool
		errDetail string
		//b         byte
	)

	// instruction counting...
	instrCounts := make(map[string]uint64)

	// instruction disassembly slows CPU down by about 3x, for the moment it seems to make sense
	// for it to follow the debugLogging setting...
	disassembly := debugLogging

	// direct console input to the VM
	cpu.cpuMu.Lock()
	cpu.scpIO = false
	cpu.cpuMu.Unlock()

	// initial read lock taken before loop starts to eliminate one lock/unlock per cycle
	cpu.cpuMu.RLock()

RunLoop: // performance-critical section starts here
	for {
		// FETCH
		thisOp = memory.ReadWord(cpu.pc)

		// DECODE
		iPtr, ok = instructionDecode(thisOp, cpu.pc, cpu.sbr[cpu.pc>>29].lef, cpu.sbr[cpu.pc>>29].io, cpu.atu, disassembly)
		cpu.cpuMu.RUnlock()
		if !ok {
			errDetail = " *** Error: could not decode instruction ***"
			break
		}

		if debugLogging {
			logging.DebugPrint(logging.DebugLog, "%s\t\t%s\n", iPtr.disassembly, cpuCompactPrintableStatus())
		}

		// EXECUTE
		if !cpuExecute(iPtr) {
			errDetail = " *** Error: could not execute instruction (or CPU HALT encountered) ***"
			break
		}

		// BREAKPOINT?
		if len(breakpoints) > 0 {
			cpu.cpuMu.Lock()
			for _, bAddr := range breakpoints {
				if bAddr == cpu.pc {
					cpu.scpIO = true
					cpu.cpuMu.Unlock()
					msg := fmt.Sprintf(" *** BREAKpoint hit at physical address %d. ***", cpu.pc)
					ttoPutNLString(msg)
					log.Println(msg)

					break RunLoop
				}
			}
			cpu.cpuMu.Unlock()
		}

		// Console interrupt?
		cpu.cpuMu.RLock()
		if cpu.scpIO {
			cpu.cpuMu.RUnlock()
			errDetail = " *** Console ESCape ***"
			break
		}

		// instruction counting
		instrCounts[iPtr.mnemonic]++

		// N.B. RLock still in effect as we loop around
	}

	// end of performance-critical section

	cpu.cpuMu.Lock()
	cpu.scpIO = true
	cpu.cpuMu.Unlock()

	// run halted due to either error or console escape
	log.Println(errDetail)
	ttoPutNLString(errDetail)
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "%s\n", cpuPrintableStatus())
	}
	ttoPutString(cpuPrintableStatus())

	errDetail = " *** CPU halting ***"
	log.Println(errDetail)
	ttoPutNLString(errDetail)

	errDetail = fmt.Sprintf(" *** MV/Em executed %d	 instructions ***", cpu.instrCount)
	log.Println(errDetail)
	ttoPutNLString(errDetail)

	// instruction counts, first by Mnemonic, then by count
	var mnems []string
	for m := range instrCounts {
		mnems = append(mnems, m)
	}
	sort.Strings(mnems)
	log.Println("Instruction Execution Count by Mnemonic")
	for _, m := range mnems {
		log.Printf("%s\t%d\n", m, instrCounts[m])
	}

	n := map[int][]string{}
	var a []int
	for mn, v := range instrCounts {
		n[int(v)] = append(n[int(v)], mn)
	}
	for k := range n {
		a = append(a, int(k))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(a)))
	log.Println("Instruction Execution Count by Count")
	for _, k := range a {
		for _, s := range n[k] {
			log.Printf("%d\t%s\n", k, s)
		}
	}

}
