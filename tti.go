// tti - console input

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
	"mvemg/dg"
	"net"
	"os"
	"sync"
)

var (
	tti          net.Conn
	oneCharBuf   byte
	oneCharBufMu sync.Mutex
)

func ttiInit(c net.Conn, cpuPtr *CPUT, ch chan<- byte) {
	tti = c
	busAddDevice(DEV_TTI, "TTI", TTI_PMB, true, true, false)
	busSetResetFunc(DEV_TTI, ttiReset)
	busSetDataInFunc(DEV_TTI, ttiDataIn)
	go ttiListener(cpuPtr, ch)
}

func ttiListener(cpuPtr *CPUT, scpChan chan<- byte) {
	b := make([]byte, 80)
	for {
		n, err := tti.Read(b)
		if err != nil || n == 0 {
			log.Println("ERROR: could not read from console port: ", err.Error())
			os.Exit(1)
		}
		//log.Printf("DEBUG: ttiListener() got <%c>\n", b[0])
		for c := 0; c < n; c++ {
			// console ESCape?
			if b[c] == ASCII_ESC || b[c] == 0 {
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
				oneCharBufMu.Lock()
				oneCharBuf = b[c]
				oneCharBufMu.Unlock()
				busSetDone(DEV_TTI, true)
			}
		}
	}
}

func ttiReset() {
	log.Println("INFO: TTI Reset")
}

// This is called from Bus to implement DIA from the TTI DEV_TTIice
func ttiDataIn(cpuPtr *CPUT, iPtr *novaDataIoT, abc byte) {
	oneCharBufMu.Lock()
	cpuPtr.ac[iPtr.acd] = dg.DwordT(oneCharBuf) // grab the char from the buffer
	oneCharBufMu.Unlock()
	switch abc {
	case 'A':
		switch iPtr.f {
		case 'S':
			busSetBusy(DEV_TTI, true)
			busSetDone(DEV_TTI, false)
		case 'C':
			busSetBusy(DEV_TTI, false)
			busSetDone(DEV_TTI, false)
		}

	default:
		log.Fatalf("ERROR: unexpected source buffer <%c> for DOx ac,TTO instruction\n", abc)
	}
}
