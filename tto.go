// tto - console output

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
	"log"
	"mvemg/logging"
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
func ttoDataOut(cpuPtr *CPUT, iPtr *novaDataIoT, abc byte) {
	var ascii byte
	switch abc {
	case 'A':
		ascii = byte(cpuPtr.ac[iPtr.acd])
		logging.DebugPrint(logging.DebugLog, "ttoDataOut: AC# %d contains %d                                 yielding ASCII char<%c>\n",
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
