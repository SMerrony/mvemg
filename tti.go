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
	"net"
	"os"
	"sync"

	"github.com/SMerrony/dgemug"
)

var (
	tti          net.Conn
	oneCharBuf   byte
	oneCharBufMu sync.Mutex
)

func ttiInit(c net.Conn, cpuPtr *CPUT, ch chan<- byte) {
	tti = c
	busAddDevice(devTTI, "TTI", pmbTTI, true, true, false)
	busSetResetFunc(devTTI, ttiReset)
	busSetDataInFunc(devTTI, ttiDataIn)
	busSetDataOutFunc(devTTI, ttiDataOut)
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
				oneCharBufMu.Lock()
				oneCharBuf = b[c]
				oneCharBufMu.Unlock()
				busSetDone(devTTI, true)
			}
		}
	}
}

func ttiReset() {
	log.Println("INFO: TTI Reset")
}

// This is called from Bus to implement DIA from the TTI devTTIice
func ttiDataIn(cpuPtr *CPUT, iPtr *novaDataIoT, abc byte) {
	oneCharBufMu.Lock()
	cpuPtr.ac[iPtr.acd] = dg.DwordT(oneCharBuf) // grab the char from the buffer
	oneCharBufMu.Unlock()
	switch abc {
	case 'A':
		switch iPtr.f {
		case 'S':
			busSetBusy(devTTI, true)
			busSetDone(devTTI, false)
		case 'C':
			busSetBusy(devTTI, false)
			busSetDone(devTTI, false)
		}

	default:
		log.Fatalf("ERROR: unexpected source buffer <%c> for DOx ac,TTO instruction\n", abc)
	}
}

// this is only here to support NIO commands to TTI
func ttiDataOut(cpuPtr *CPUT, iPtr *novaDataIoT, abc byte) {
	switch abc {
	case 'N':
		switch iPtr.f {
		case 'S':
			busSetBusy(devTTI, true)
			busSetDone(devTTI, false)
		case 'C':
			busSetBusy(devTTI, false)
			busSetDone(devTTI, false)
		}
	default:
		log.Fatalf("ERROR: unexpected call to ttiDataOut with abc(n) flag set to %c\n", abc)
	}
}
