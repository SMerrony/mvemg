// tto - console output
package main

import (
	"log"
	"net"
)

type Tto struct {
	conn net.Conn
}

var t Tto

func ttoInit(c net.Conn) {
	t.conn = c
	bus.busAddDevice(DEV_TTO, "TTO", TTO_PMB, true, true, false)
	bus.busSetResetFunc(DEV_TTO, ttoReset)
	bus.busSetDataOutFunc(DEV_TTO, ttoDataOut)
}

func ttoPutChar(c byte) {
	t.conn.Write([]byte{c})
}

func ttoPutString(s string) {
	t.conn.Write([]byte(s))
}

func ttoPutStringNL(s string) {
	t.conn.Write([]byte(s))
	t.conn.Write([]byte{ASCII_NL})
}

func ttoPutNLString(s string) {
	t.conn.Write([]byte{ASCII_NL})
	t.conn.Write([]byte(s))
}

func ttoReset() {
	ttoPutChar(ASCII_FF)
	log.Println("INFO: TTO Reset")
}

func ttoDataOut(cpuPtr *Cpu, iPtr *DecodedInstr, abc byte) {
	var ascii byte
	switch abc {
	case 'A':
		ascii = byte(cpuPtr.ac[iPtr.acd])
		log.Printf("ttoDataOut: AC# %d contains %d                                 yielding ASCII char<%c>\n",
			iPtr.acd, cpuPtr.ac[iPtr.acd], ascii)
		if iPtr.f == 'S' {
			bus.busSetBusy(DEV_TTO, true)
			bus.busSetDone(DEV_TTO, false)
		}
		ttoPutChar(ascii)
		bus.busSetBusy(DEV_TTO, false)
		bus.busSetDone(DEV_TTO, true)
	default:
		log.Fatalf("ERROR: unexpected source buffer <%c> for DOx ac,TTO instruction\n", abc)
	}
}
