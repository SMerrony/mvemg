// statusCollector.go

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
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/SMerrony/dgemug/util"
)

const (
	// define which screen row each of the monitored data appear on
	statCPUrow        = 3
	statCPUrow2       = 5
	statDPFrow        = 7
	statDSKProw       = 9
	statMTBrow        = 11
	statInternalsRow  = 20
	statInternalsRow2 = 21
)

// StatusCollector maintains a near real-time status screen available on STAT_PORT.
//
// The screen uses DG DASHER control codes for formatting, so a DASHER terminal emulator
// should be attached to it for good results.
//
// The function (which is intended to be run as a goroutine) listens for status updates
// from known senders' channels and upon receiving an update refreshes the display of that status
// on the monitor page.  It is therefore the responsibility of the sender to update the
// status as often as it sees fit.
func statusCollector(
	cpuChan chan cpuStatT,
	dpfChan chan DpfStatT,
	dskpChan chan dskpStatT,
	mtbChan chan mtbStatT) {

	var (
		cpuStats                               cpuStatT
		lastIcount, iCount                     uint64
		ips, dpfIops, dskpIops                 float64
		lastCPUtime, lastDpfTime, lastDskpTime time.Time
		thisDpfIOcnt, lastDpfIOcnt             uint64
		thisDskpIOcnt, lastDskpIOcnt           uint64
		dpfStats                               DpfStatT
		dskpStats                              dskpStatT
		mtbStats                               mtbStatT
	)

	l, err := net.Listen("tcp", "localhost:"+StatPort)
	if err != nil {
		log.Println("ERROR: Could not listen on stats port: ", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {

		conn, err := l.Accept()
		if err != nil {
			log.Println("ERROR: Could not accept on stats port: ", err.Error())
			os.Exit(1)
		}

		statusSendString(conn, fmt.Sprintf("%c                             MV/Em Status\012", dasherERASEPAGE))
		statusSendString(conn, "                             ============")

		for {
			// blocking wait for a status update to arrive
			select {
			case cpuStats = <-cpuChan:
				iCount = cpuStats.instrCount - lastIcount
				lastIcount = cpuStats.instrCount
				ips = float64(iCount) / (time.Since(lastCPUtime).Seconds() * 1000)
				lastCPUtime = time.Now()
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statCPUrow, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("PC:  %010d   Interrupts: %s    ATU: %s     IPS: %.fk/sec",
					cpuStats.pc,
					util.BoolToOnOff(cpuStats.ion),
					util.BoolToOnOff(cpuStats.atu),
					ips))
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statCPUrow2, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("AC0: %010d   AC1: %010d   AC2: %010d   AC3: %010d",
					cpuStats.ac[0],
					cpuStats.ac[1],
					cpuStats.ac[2],
					cpuStats.ac[3]))
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statInternalsRow, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("MV/Em - Version: %s (%s) built with %s",
					Version, ReleaseType,
					cpuStats.goVersion))
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statInternalsRow2, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("        Host CPUs: %d  Goroutines: %d  Heap: %dMB",
					cpuStats.hostCPUCount,
					cpuStats.goroutineCount,
					cpuStats.heapSizeMB))

			case dpfStats = <-dpfChan:
				thisDpfIOcnt = dpfStats.writes + dpfStats.reads
				dpfIops = float64(thisDpfIOcnt-lastDpfIOcnt) / time.Since(lastDpfTime).Seconds()
				lastDpfIOcnt = thisDpfIOcnt
				lastDpfTime = time.Now()
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statDPFrow, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("DPF  (DPF0) - Attached: %c  IOPS: %.f CYL: %04d  HD: %02d  SECT: %03d",
					util.BoolToYN(dpfStats.imageAttached),
					dpfIops,
					dpfStats.cylinder,
					dpfStats.head,
					dpfStats.sector))

			case dskpStats = <-dskpChan:
				thisDskpIOcnt = dskpStats.writes + dskpStats.reads
				dskpIops = float64(thisDskpIOcnt-lastDskpIOcnt) / time.Since(lastDskpTime).Seconds()
				lastDskpIOcnt = thisDskpIOcnt
				lastDskpTime = time.Now()
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statDSKProw, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("DSKP (DPJ0) - Attached: %c  IOPS: %.f  SECNUM: %08d",
					util.BoolToYN(dskpStats.imageAttached),
					dskpIops,
					//dskpStats.cylinder,
					//dskpStats.head,
					//dskpStats.sector,
					dskpStats.sectorNo))

			case mtbStats = <-mtbChan:
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statMTBrow, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("MTB  (MTC0) - Attached: %c  File: %s  Mem Addr: %010d  Curr Cmd: %d",
					util.BoolToYN(mtbStats.imageAttached[0]),
					mtbStats.fileName[0],
					mtbStats.memAddrReg,
					mtbStats.currentCmd))
			}
		}
	}
}

func statusSendString(con net.Conn, s string) {
	con.Write([]byte(s))
}
