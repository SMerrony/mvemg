// mvemg project main.go

// Copyright (C) 2017,2018,2019  Steve Merrony

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
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SMerrony/dgemug/devices"
	"github.com/SMerrony/dgemug/dg"
	"github.com/SMerrony/dgemug/logging"
	"github.com/SMerrony/dgemug/memory"
)

// import "github.com/pkg/profile"

const (
	// Displayable name of emulator
	appName = "MV/Em"
	// appVersion number
	appVersion = "v0.1.0"
	// appReleaseType - Alpha, Beta, Production etc.
	appReleaseType = "Prerelease"
	// ScpBuffSize is the char buffer length for SCP input lines
	ScpBuffSize = 135

	// MemSizeWords defines the size of MV/Em's emulated RAM in 16-bit words
	MemSizeWords = 8388608 // = 040 000 000 (8) = 0x80 0000
	// MemSizeLCPID is the code returned by the LCPID to indicate the size of RAM in half megabytes
	MemSizeLCPID = ((MemSizeWords * 2) / (256 * 1024)) - 1 // 0x3F
	// MemSizeNCLID is the code returned by NCLID to indicate size of RAM in 32Kb increments
	MemSizeNCLID = ((MemSizeWords * 2) / (32 * 1024)) - 1

	cmdUnknown = " *** UNKNOWN SCP-CLI COMMAND ***"
	cmdNYI     = "Command Not Yet Implemented"

	defaultRadix = 8
)

var (
	// debugLogging - CPU runs about 3x faster without debugLogging
	// (and another 3x faster without disassembly, linked to this)
	debugLogging  = true
	breakpoints   []dg.PhysAddrT
	cpuPtr        *CPUT
	cpuStatsChan  chan cpuStatT
	dpfStatsChan  chan devices.Disk6061StatT
	dskpStatsChan chan devices.Disk6239StatT
	mtbStatsChan  chan devices.MtStatT
	ttiSCPchan    chan byte

	tti  devices.TtiT
	tto  devices.TtoT
	bus  devices.BusT
	dpf  devices.Disk6061T
	dskp devices.Disk6239DataT
	mtb  devices.MagTape6026T

	inputRadix = defaultRadix
)

// flags
var (
	consoleAddrFlag = flag.String("consoleaddr", "localhost:10000", "network interface/port for console")
	doFlag          = flag.String("do", "", "run script `file` at startup")
	statusAddrFlag  = flag.String("statusaddr", "localhost:9999", "network interface/port for status monitoring")
	cpuprofile      = flag.String("cpuprofile", "", "write cpu profile `file`")
	memprofile      = flag.String("memprofile", "", "write memory profile to `file`")
)

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
	}

	log.Printf("INFO: %s will not start until console connected to  %s.\n", appName, *consoleAddrFlag)

	l, err := net.Listen("tcp", *consoleAddrFlag)
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
		dpfStatsChan = make(chan devices.Disk6061StatT, 3)
		dskpStatsChan = make(chan devices.Disk6239StatT, 3)
		mtbStatsChan = make(chan devices.MtStatT, 3)

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

		memory.MemInit(MemSizeWords, debugLogging)
		bus.BusInit()
		bus.AddDevice(deviceMap, devBMC, true)
		bus.SetResetFunc(devBMC, memory.BmcdchReset) // created by memory, needs bus!

		bus.AddDevice(deviceMap, devSCP, true)
		instructionsInit()
		decoderGenAllPossOpcodes()
		cpuPtr = cpuInit(cpuStatsChan)

		bus.AddDevice(deviceMap, devTTO, true)
		tto.Init(devTTO, &bus, conn)

		bus.AddDevice(deviceMap, devTTI, true)
		//ttiInit(conn, cpuPtr, ttiSCPchan)
		tti.Init(devTTI, &bus)
		go consoleListener(conn, cpuPtr, ttiSCPchan, &tti)

		bus.AddDevice(deviceMap, devMTB, false)
		mtb.MtInit(devMTB, &bus, mtbStatsChan, logging.MtLog, debugLogging)

		bus.AddDevice(deviceMap, devDPF, false)
		dpf.Disk6061Init(devDPF, &bus, dpfStatsChan, logging.DpfLog, debugLogging)

		bus.AddDevice(deviceMap, devDSKP, false)
		dskp.Disk6239Init(devDSKP, &bus, dskpStatsChan, logging.DskpLog, debugLogging)

		// say hello...
		tto.PutChar(asciiFF)
		tto.PutStringNL(" *** Welcome to the MV/Emulator - Type HE for help ***")

		// kick off the status monitor routine
		go statusCollector(*statusAddrFlag, cpuStatsChan, dpfStatsChan, dskpStatsChan, mtbStatsChan)

		// run any command specified on the command line
		if *doFlag != "" {
			command := fmt.Sprintf("DO %s", *doFlag)
			log.Printf("INFO: got startup command <%s>\n", command)
			doCommand(command) // N.B. will not pass here until start-up script is complete...
		}

		// the main SCP/console interaction loop
		cpuPtr.cpuMu.Lock()
		cpuPtr.scpIO = true
		cpuPtr.cpuMu.Unlock()
		for {
			tto.PutNLString("SCP-CLI> ")
			command := scpGetLine()
			//log.Println("INFO: Got SCP command: " + command)
			doCommand(command)
		}
	}
}

func fmtRadixVerb() string {
	switch inputRadix {
	case 2:
		return "%b"
	case 8:
		return "%#o"
	case 10:
		return "%d."
	case 16:
		return "%#x"
	default:
		log.Fatalf("ERROR: Invalid input radix %d", inputRadix)
		return ""
	}
}

func consoleListener(con net.Conn, cpuPtr *CPUT, scpChan chan<- byte, tti *devices.TtiT) {
	b := make([]byte, 80)
	for {
		n, err := con.Read(b)
		if err != nil || n == 0 {
			log.Println("ERROR: could not read from console port: ", err.Error())
			os.Exit(1)
		}
		//log.Printf("DEBUG: ttiListener() got <%c>\n", b[0])
		for c := 0; c < n; c++ {
			// console ESCape?
			//if b[c] == asciiESC || b[c] == 0 {
			if b[c] == asciiESC {
				cpuPtr.cpuMu.Lock()
				cpuPtr.scpIO = true
				cpuPtr.cpuMu.Unlock()
				break // don't want to send the ESC itself to the SCP
			}
			cpuPtr.cpuMu.RLock()
			scp := cpuPtr.scpIO
			cpuPtr.cpuMu.RUnlock()
			if scp {
				// to the SCP
				scpChan <- b[c]
			} else {
				// to the CPU
				tti.InsertChar(b[c])
				// oneCharBufMu.Lock()
				// oneCharBuf = b[c]
				// oneCharBufMu.Unlock()
				// bus.SetDone(devTTI, true)
				// // send IRQ if not masked out
				// if !bus.IsDevMasked(devTTI) {
				// 	bus.SendInterrupt(devTTI)
				// }
			}
		}
	}
}

// Get one line from the console - handle DASHER DELete key as corrector
func scpGetLine() string {
	line := []byte{}
	var cc byte
	for cc != asciiCR {
		cc = <-ttiSCPchan
		//cc = ttiGetChar()
		// handle the DASHER Delete key
		if cc == dasherDELETE && len(line) > 0 {
			tto.PutChar(dasherCURSORLEFT)
			line = line[:len(line)-1]
		} else {
			tto.PutChar(cc)
			line = append(line, cc)
		}
	}
	// we don't want the final CR
	line = line[:len(line)-1]

	return string(line[:])
}

// Exit cleanly, tidying up as much as we can
func cleanExit() {
	tto.PutNLString(" *** MV/Emulator stopping at user request ***")
	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}
	logging.DebugLogsDump("logs/")
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
		tto.PutString(cpuPrintableStatus())
	case "B":
		boot(words)
	case "CO":
		run()
	case "E":
		examine(words)
	case "HE":
		showHelp()
	case "RE":
		reset()
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
		tto.PutStringNL(mtb.MtScanImage(0))
	case "CREATE":
		createBlank(words)
	case "DET":
		detach(words)
	case "DIS":
		disassemble(words)
	case "DO":
		doScript(words)
	case "exit", "EXIT", "QUIT":
		cleanExit()
	case "LOAD":
		tto.PutNLString(memory.LoadFromASCIIFile(words[1]))
	case "NOBREAK":
		breakClear(words)
	case "SET":
		set(words)
	case "SH", "SHO", "SHOW":
		show(words)
	default:
		tto.PutNLString(cmdUnknown)
	}
}

/* Commands are below here... */

// Attach an image file to an emulated device
func attach(cmd []string) {
	if len(cmd) < 3 {
		tto.PutNLString(" *** ATT command requires arguments: <dev> and <image> ***")
		return
	}
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "INFO: Attach called  with parms <%s> <%s>\n", cmd[1], cmd[2])
	}
	switch cmd[1] {
	case "MTB":
		if mtb.MtAttach(0, cmd[2]) {
			tto.PutNLString(" *** Tape Image Attached ***")
		} else {
			tto.PutNLString(" *** Could not ATTach Tape Image ***")
		}

	case "DPF":
		if dpf.Disk6061Attach(0, cmd[2]) {
			tto.PutNLString(" *** DPF Disk Image Attached ***")
		} else {
			tto.PutNLString(" *** Could not ATTach DPF Disk Image ***")
		}

	case "DSKP":
		if dskp.Disk6239Attach(0, cmd[2]) {
			tto.PutNLString(" *** DSKP Disk Image Attached ***")
		} else {
			tto.PutNLString(" *** Could not ATTach DSKP Disk Image ***")
		}

	default:
		tto.PutNLString(" *** Unknown or unimplemented Device for ATT command ***")
	}
}

func boot(cmd []string) {
	if len(cmd) != 2 {
		tto.PutNLString(" *** B command requires <devicenumber> ***")
		return
	}
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "INFO: Boot called  with parm <%s>\n", cmd[1])
	}
	dev, err := strconv.ParseInt(cmd[1], inputRadix, 16)
	devNum := int(dev)
	if err != nil {
		tto.PutNLString(" *** Expecting <devicenumber> after B ***")
		return
	}
	if !bus.IsAttached(devNum) {
		tto.PutNLString(" *** Device is not ATTached ***")
		return
	}
	if !bus.IsBootable(devNum) {
		tto.PutNLString(" *** Device is not bootable ***")
		return
	}
	memory.MemInit(MemSizeWords, debugLogging)
	switch devNum {
	case devMTB:
		mtb.MtLoadTBoot()
		cpu.cpuMu.Lock()
		cpu.sr = 0x8000 | devMTB
		cpu.ac[0] = devMTB
		cpu.pc = 10
		cpu.cpuMu.Unlock()
	case devDPF:
		dpf.Disk6061LoadDKBT()
		cpu.cpuMu.Lock()
		cpu.sr = 0x8000 | devDPF
		cpu.ac[0] = devDPF
		cpu.pc = 0377
		cpu.cpuMu.Unlock()
	case devDSKP:
		dskp.Disk6239LoadDKBT()
		cpu.cpuMu.Lock()
		cpu.sr = 0x8000 | devDSKP
		cpu.ac[0] = devDSKP
		cpu.pc = 0377
		cpu.cpuMu.Unlock()
	default:
		tto.PutNLString(" *** Booting from that device not yet implemented ***")
	}
}

func breakSet(cmd []string) {
	if len(cmd) != 2 {
		tto.PutNLString(" *** BREAK command requires a single physical <address> argument ***")
		return
	}
	pAddr, err := strconv.ParseInt(cmd[1], inputRadix, 32)
	if err != nil {
		tto.PutNLString(" *** BREAK command could not parse <address> argument ***")
		return
	}
	breakpoints = append(breakpoints, dg.PhysAddrT(pAddr))

	tto.PutNLString("BREAKpoint set")
}

func breakClear(cmd []string) {
	if len(cmd) != 2 {
		tto.PutNLString(" *** NOBREAK command requires a single physical <address> argument ***")
		return
	}
	pAddr, err := strconv.ParseInt(cmd[1], inputRadix, 16)
	if err != nil {
		tto.PutNLString(" *** NOBREAK command could not parse <address> argument ***")
		return
	}
	cAddr := dg.PhysAddrT(pAddr)
	for ix, addr := range breakpoints {
		if addr == cAddr {
			breakpoints[ix] = breakpoints[len(breakpoints)-1]
			breakpoints = breakpoints[:len(breakpoints)-1]
			tto.PutNLString(" *** Cleared breakpoint ***")
		}
	}
}

func createBlank(cmd []string) {
	if len(cmd) < 3 {
		tto.PutNLString(" *** Expecting DPF|DSKP <filename> args for CREATE command ***")
		return
	}
	switch cmd[1] {
	case "DPF":
		tto.PutNLString("Attempting to CREATE new empty DPF-type disk image, please wait...")
		if dpf.Disk6061CreateBlank(cmd[2]) {
			tto.PutNLString("Empty MV/Em DPF-type disk image created")
		} else {
			tto.PutNLString(" *** Error: could not create empty disk image ***")
		}
	case "DSKP":
		tto.PutNLString("Attempting to CREATE new empty DSKP-type disk image, please wait...")
		if dskp.Disk6239CreateBlank(cmd[2]) {
			tto.PutNLString("Empty MV/Em DSKP-type disk image created")
		} else {
			tto.PutNLString(" *** Error: could not create empty disk image ***")
		}
	default:
		tto.PutNLString(" *** CREATE not yet supported for that device type ***")
	}
}

func detach(cmd []string) {
	if len(cmd) < 2 {
		tto.PutNLString(" *** DET command requires argument: <dev> ***")
		return
	}
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "INFO: Detach called  with parm <%s> \n", cmd[1])
	}
	switch cmd[1] {
	case "MTB":
		if mtb.MtDetach(0) {
			tto.PutNLString(" *** Tape Image Detached ***")
		} else {
			tto.PutNLString(" *** Could not DETach Tape Image ***")
		}
	default:
		tto.PutNLString(" *** Unknown or unimplemented Device for DET command ***")
	}
}

func disassemble(cmd []string) {
	var (
		lowAddr, highAddr dg.PhysAddrT
		word              dg.WordT
		byte1, byte2      dg.ByteT
		display           string
		skipDecode        int
	)
	if len(cmd) == 1 {
		tto.PutNLString(" *** DIS command requires an address ***")
		return
	}
	cmd1 := cmd[1]
	intVal1, err := strconv.ParseInt(cmd[1], inputRadix, 32)
	if err != nil {
		tto.PutNLString(" *** Invalid address ***")
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
			intVal2, err := strconv.ParseInt(cmd[2], inputRadix, 32)
			if err != nil {
				tto.PutNLString(" *** Invalid address ***")
				return
			}
			highAddr = dg.PhysAddrT(intVal2)
		}
	}
	if highAddr < lowAddr {
		tto.PutNLString(" *** Invalid address range ***")
		return
	}
	for addr := lowAddr; addr <= highAddr; addr++ {
		word = memory.ReadWord(addr)
		byte1 = dg.ByteT(word >> 8)
		byte2 = dg.ByteT(word & 0x00ff)
		display = fmt.Sprintf(fmtRadixVerb()+": %02X %02X %03o %03o %s \"", addr, byte1, byte2, byte1, byte2,
			memory.WordToBinStr(word))
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
			instrTmp, ok := instructionDecode(word, addr, cpu.sbr[cpu.pc>>29].lef, cpu.sbr[cpu.pc>>29].io, cpu.atu, true)
			if ok {
				display += instrTmp.disassembly
				if instrTmp.instrLength > 1 {
					skipDecode = instrTmp.instrLength - 1
				}
			} else {
				display += " *** Could not decode ***"
			}
		} else {
			skipDecode--
		}
		tto.PutNLString(display)
	}
}

func doScript(cmd []string) {
	if len(cmd) < 2 {
		tto.PutNLString(" *** DO command required <scriptfile> ***")
		return
	}
	scriptFile, err := os.Open(cmd[1])
	if err != nil {
		tto.PutNLString(" *** Could not open MV/Em command script ***")
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
			tto.PutNLString(doCmd)
			doCommand(doCmd)
		}
	}

}

// examine mimics the E command from later SCP-CLIs
func examine(cmd []string) {
	if len(cmd) < 2 {
		tto.PutNLString(" *** Examine - missing parameter ***")
		return
	}
	switch cmd[1] {
	case "A":
		if len(cmd) < 3 {
			tto.PutNLString(" *** Examine Accumulator - invalid AC number ***")
			return
		}
		exAc, err := strconv.ParseInt(cmd[2], inputRadix, 16)
		if err != nil || exAc < 0 || exAc > 3 {
			tto.PutNLString(" *** Examine Accumulator - invalid AC number ***")
			return
		}
		prompt := fmt.Sprintf("AC%d = "+fmtRadixVerb()+" - Enter new val or just ENTER> ", exAc, cpu.ac[exAc])
		tto.PutNLString(prompt)
		resp := scpGetLine()
		if len(resp) > 0 {
			newVal, err := strconv.ParseInt(resp, inputRadix, 16)
			if err != nil {
				tto.PutNLString(" *** Could not parse new AC value ***")
				return
			}
			cpu.ac[exAc] = dg.DwordT(newVal)
			prompt = fmt.Sprintf("AC%d = "+fmtRadixVerb(), exAc, cpu.ac[exAc])
			tto.PutNLString(prompt)
		}
	case "M":
		if len(cmd) < 3 {
			tto.PutNLString(" *** Examine Memory - invalid address ***")
			return
		}
		exMem, err := strconv.ParseInt(cmd[2], inputRadix, 16)
		if err != nil || exMem < 0 || exMem > MemSizeWords {
			tto.PutNLString(" *** Examine Memory - invalid address ***")
			return
		}
		prompt := fmt.Sprintf("Location "+fmtRadixVerb()+" contains "+fmtRadixVerb()+" - Enter new val or just ENTER> ", exMem, memory.ReadWord(dg.PhysAddrT(exMem)))
		tto.PutNLString(prompt)
		resp := scpGetLine()
		if len(resp) > 0 {
			newVal, err := strconv.ParseInt(resp, inputRadix, 16)
			if err != nil {
				tto.PutNLString(" *** Could not parse new value ***")
				return
			}
			memory.WriteWord(dg.PhysAddrT(exMem), dg.WordT(newVal))
			prompt = fmt.Sprintf("Location "+fmtRadixVerb()+" = "+fmtRadixVerb()+"", exMem, memory.ReadWord(dg.PhysAddrT(exMem)))
			tto.PutNLString(prompt)
		}
	case "P":
		prompt := fmt.Sprintf("PC = "+fmtRadixVerb()+" - Enter new val or just ENTER> ", cpu.pc)
		tto.PutNLString(prompt)
		resp := scpGetLine()
		if len(resp) > 0 {
			newVal, err := strconv.ParseInt(resp, inputRadix, 16)
			if err != nil {
				tto.PutNLString(" *** Could not parse new PC value ***")
				return
			}
			cpu.pc = dg.PhysAddrT(newVal)
			prompt = fmt.Sprintf("PC = "+fmtRadixVerb(), cpu.pc)
			tto.PutNLString(prompt)
		}
	default:
		tto.PutNLString(" *** Expecting A, M, or P for E(xamine) command ***")
		return
	}
}

func printableBreakpointList() string {
	if len(breakpoints) == 0 {
		return " *** No BREAKpoints are set ***"
	}
	res := "BREAKpoint(s) at: "
	for _, b := range breakpoints {
		res += fmt.Sprintf(fmtRadixVerb()+" ", b)
	}
	return res
}

// reset should bring the emulator back to its initial state
func reset() {
	memory.MemInit(MemSizeWords, debugLogging)
	bus.ResetAllIODevices()
	cpuReset()
	// mtbReset() // Not Init
	// dpfReset()
	// dskpReset()
}

func set(cmd []string) {
	if len(cmd) < 3 {
		tto.PutNLString(" *** Expecting SET subcommand ***")
		return
	}
	switch cmd[1] {
	case "LOGGING":
		switch cmd[2] {
		case "ON":
			debugLogging = true
		case "OFF":
			debugLogging = false
		}

	default:
		tto.PutNLString(" *** Unknown SET subcommand ***")
	}
}

// showHelp - Display SCP and Emulator help on the DASHER-compatible console
// N.B. Ensure this fits on a 24x80 screen
func showHelp() {
	tto.PutString("\014                          \024SCP-CLI Commands\025" +
		"                               \034MV/EMG\035\012" +
		" .                      - Display state of CPU\012" +
		" B #                    - Boot from device #\012" +
		" CO                     - COntinue CPU Processing\012" +
		" E A <#> | M [addr] | P - Examine/Modify Acc/Memory/PC\012" +
		" HE                     - HElp (show this)\012" +
		" RE                     - REset the system\012" +
		" SS                     - Single Step one instruction\012" +
		" ST <addr>              - STart processing at specified address\012")
	tto.PutString("\012                          \024Emulator Commands\025\012" +
		" ATT <dev> <file> [RW]  - ATTach the image file to named device (RO)\012" +
		" BREAK/NOBREAK <addr>   - Set or clear a BREAKpoint\012" +
		" CHECK                  - CHECK validity of attached TAPE image\012" +
		" CREATE DPF|DSKP <file> - CREATE an empty/unformatted disk image\012" +
		" DET <dev>              - DETach any image file from the device\012" +
		" DIS <from> <to>|+<#>   - DISassemble physical memory range or # from PC\012" +
		" DO <file>              - DO (i.e. run) emulator commands from script <file>\012" +
		" EXIT                   - EXIT the emulator\012" +
		" LOAD <file>            - Load ASCII octal file directly into memory\012" +
		" SET LOGGING ON|OFF     - Turn on or off debug logging (logs dumped end of run)\012" +
		" SHOW BREAK/DEV/LOGGING - SHOW list of BREAKpoints/DEVices configured\012")
}

// Show various emulator states to the user
func show(cmd []string) {
	if len(cmd) == 1 {
		tto.PutNLString(" *** SHOW requires argument ***")
		return
	}
	switch cmd[1] {
	case "DEV":
		tto.PutNLString(bus.GetPrintableDevList())
	case "BREAK":
		tto.PutNLString(printableBreakpointList())
	case "LOGGING":
		resp := fmt.Sprintf("Logging is currently turned %s", memory.BoolToOnOff(debugLogging))
		tto.PutNLString(resp)
	default:
		tto.PutNLString(" *** Invalid SHOW type ***")
	}
}

// Attempt to execute the opcode at PC
func singleStep() {
	tto.PutString(cpuPrintableStatus())
	// FETCH
	thisOp := memory.ReadWord(cpu.pc)
	// DECODE
	if iPtr, ok := instructionDecode(thisOp, cpu.pc, cpu.sbr[cpu.pc>>29].lef, cpu.sbr[cpu.pc>>29].io, cpu.atu, true); ok {
		tto.PutNLString(iPtr.disassembly)
		// EXECUTE
		if cpuExecute(iPtr) {
			tto.PutString(cpuPrintableStatus())
		} else {
			tto.PutNLString(" *** Error: could not execute instruction")
		}
	} else {
		tto.PutNLString(" *** Error: could not decode opcode")
	}
}

// start running at user-provided PC
func start(cmd []string) {
	if len(cmd) < 2 {
		tto.PutNLString(" *** ST command requires start address ***")
		return
	}
	newPc, err := strconv.ParseInt(cmd[1], inputRadix, 16)
	if err != nil || newPc < 0 {
		tto.PutNLString(" *** Could not parse new PC value ***")
		return
	}
	cpu.pc = dg.PhysAddrT(newPc)
	run()
}

// The main Emulator running loop...
func run() {
	var (
		thisOp      dg.WordT
		prevPC      dg.PhysAddrT
		iPtr        *decodedInstrT
		ok          bool
		indIrq      byte
		errDetail   string
		instrCounts [maxInstrs]int
	)

	// instruction disassembly slows CPU down by about 3x, for the moment it seems to make sense
	// for it to follow the debugLogging setting...
	disassembly := debugLogging

	cpu.cpuMu.Lock()

	cpu.instrCount = 0 // reset instruction counter each run for MIPS calc.
	cpu.scpIO = false  // direct console input to the VM
	cpu.cpuMu.Unlock()

	startTime := time.Now()

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

		// INTERRUPT?
		if cpu.ion && bus.GetIRQ() {
			if debugLogging {
				logging.DebugPrint(logging.DebugLog, "<<< Interrupt >>>\n")
			}
			// disable further interrupts, reset the irq
			cpu.ion = false
			bus.SetIRQ(false)
			// TODO - disable User MAP
			// store PC in location zero
			memory.WriteWord(0, dg.WordT(cpu.pc))
			// fetch service routine address from location one
			if memory.TestWbit(memory.ReadWord(1), 0) {
				indIrq = '@'
			} else {
				indIrq = ' '
			}
			cpu.pc = resolve16bitEclipseAddr(&cpu, indIrq, absoluteMode, int16(memory.ReadWord(1)))
			// next time round RunLoop the interrupt service routine will be started...
		}

		// BREAKPOINT?
		if len(breakpoints) > 0 {
			cpu.cpuMu.Lock()
			for _, bAddr := range breakpoints {
				if bAddr == cpu.pc {
					cpu.scpIO = true
					cpu.cpuMu.Unlock()
					msg := fmt.Sprintf(" *** BREAKpoint hit at physical address "+fmtRadixVerb()+" (previous PC "+fmtRadixVerb()+") ***", cpu.pc, prevPC)
					tto.PutNLString(msg)
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
		instrCounts[iPtr.ix]++

		prevPC = cpu.pc

		// N.B. RLock still in effect as we loop around
	}

	// end of performance-critical section

	cpu.cpuMu.Lock()
	cpu.scpIO = true
	cpu.cpuMu.Unlock()

	runTime := time.Since(startTime).Seconds()
	avgMips := float64(cpu.instrCount/1000000) / runTime

	// run halted due to either error or console escape
	log.Println(errDetail)
	tto.PutNLString(errDetail)
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "%s\n", cpuPrintableStatus())
	}
	tto.PutString(cpuPrintableStatus())

	errDetail = " *** CPU halting ***"
	log.Println(errDetail)
	tto.PutNLString(errDetail)

	errDetail = fmt.Sprintf(" *** MV/Em executed %d instructions, average MIPS: %.1f ***", cpu.instrCount, avgMips)
	log.Println(errDetail)
	tto.PutNLString(errDetail)

	// instruction counts, first by Mnemonic, then by count
	m := make(map[int]string)
	keys := make([]int, 0)

	log.Println("Instruction Execution Count by Mnemonic")
	for i, c := range instrCounts {
		if instrCounts[i] > 0 {
			log.Printf("%s\t%d\n", instructionSet[i].mnemonic, c)
			if m[c] == "" {
				m[c] = instructionSet[i].mnemonic
				keys = append(keys, c)
			} else {
				m[c] += ", " + instructionSet[i].mnemonic
			}
		}
	}
	log.Println("instructions by Count")
	sort.Ints(keys)
	for _, c := range keys {
		log.Printf("%d\t%s\n", c, m[c])
	}
}
