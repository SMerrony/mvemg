// tti - console input
package main

import (
	"log"
	"net"
	"os"
)

type Tti struct {
	conn net.Conn
}

var (
	tti        Tti
	oneCharBuf byte
)

func ttiInit(c net.Conn) {
	tti.conn = c
	busAddDevice(DEV_TTI, "TTI", TTI_PMB, true, true, false)
	busSetResetFunc(DEV_TTI, ttiReset)
	busSetDataInFunc(DEV_TTI, ttiDataIn)
}

func ttiGetChar() byte {
	b := make([]byte, 80)
	n, err := tti.conn.Read(b)
	if err != nil || n == 0 {
		log.Println("ERROR: could not read from console port: ", err.Error())
		os.Exit(1)
	}
	return b[0]
}

func ttiReset() {
	log.Println("INFO: TTI Reset")
}

// This is called from Bus to implement DIA from the TTI device
func ttiDataIn(cpuPtr *Cpu, iPtr *DecodedInstr, abc byte) {

	cpuPtr.ac[iPtr.acd] = dg_dword(oneCharBuf) // grab the char from the buffer

	switch abc {
	case 'A':
		switch iPtr.f {
		case 'S':
			busSetBusy(DEV_TTO, true)
			busSetDone(DEV_TTO, false)
		case 'C':
			busSetBusy(DEV_TTO, false)
			busSetDone(DEV_TTO, false)
		}

	default:
		log.Fatalf("ERROR: unexpected source buffer <%c> for DOx ac,TTO instruction\n", abc)
	}
}

// FIXME - The ttiTask handling is rather obviously derived from the pthreads version
// in previous versions of the emulator.  It needs to be rethought in an idiomatically Go
// way using channels etc.
func ttiTask(cpuPtr *Cpu) {
	log.Println("INFO: ttiTask starting")
	for {
		if cpuPtr.consoleEsc {
			// this traps setting of the flag by the emulator rather than the user
			return
		}
		oneCharBuf = ttiGetChar()
		if oneCharBuf == ASCII_ESC {
			log.Println("INFO: ttiTask stopping due to console ESCape")
			cpuPtr.consoleEsc = true
			return
		}
		busSetDone(DEV_TTI, true)
	}
}

func ttiStartTask(c *Cpu) {
	c.consoleEsc = false
	go ttiTask(c)
}

func ttiStopThread(c *Cpu) {
	c.consoleEsc = true
	log.Println("INFO: ttiTask being terminated")

}
