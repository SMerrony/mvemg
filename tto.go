// tto - console output
package main

import (
	"log"
	"net"
)

var (
	tto net.Conn
)

func ttoInit(c net.Conn) {
	tto = c
	busAddDevice(DEV_TTO, "TTO", TTO_PMB, true, true, false)
	busSetResetFunc(DEV_TTO, ttoReset)
	busSetDataOutFunc(DEV_TTO, ttoDataOut)
}

func ttoPutChar(c byte) {
	tto.Write([]byte{c})
}

func ttoPutString(s string) {
	tto.Write([]byte(s))
}

func ttoPutStringNL(s string) {
	tto.Write([]byte(s))
	tto.Write([]byte{ASCII_NL})
}

func ttoPutNLString(s string) {
	tto.Write([]byte{ASCII_NL})
	tto.Write([]byte(s))
}

func ttoReset() {
	ttoPutChar(ASCII_FF)
	log.Println("INFO: TTO Reset")
}

// This is called from Bus to implement DOA to the TTO device
func ttoDataOut(cpuPtr *Cpu, iPtr *DecodedInstr, abc byte) {
	var ascii byte
	switch abc {
	case 'A':
		ascii = byte(cpuPtr.ac[iPtr.acd])
		debugPrint(DEBUG_LOG,"ttoDataOut: AC# %d contains %d                                 yielding ASCII char<%c>\n",
			iPtr.acd, cpuPtr.ac[iPtr.acd], ascii)
		if iPtr.f == 'S' {
			busSetBusy(DEV_TTO, true)
			busSetDone(DEV_TTO, false)
		}
		ttoPutChar(ascii)
		busSetBusy(DEV_TTO, false)
		busSetDone(DEV_TTO, true)
	default:
		log.Fatalf("ERROR: unexpected source buffer <%c> for DOx ac,TTO instruction\n", abc)
	}
}
