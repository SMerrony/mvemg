// statusCollector.go

// Copyright (C) 2017,2018  Steve Merrony

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

	"github.com/SMerrony/dgemug/devices"
	"github.com/SMerrony/dgemug/memory"
)

const (
	// define which screen row each of the monitored data appear on
	statCPUrow        = 3
	statCPUrow2       = 5
	statDPFrow        = 7
	statDSKProw       = 9
	statMTrow         = 11
	statMTrow2        = 12
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
	statusPort string,
	cpuChan chan cpuStatT,
	dpfChan chan devices.Disk6061StatT,
	dskpChan chan devices.Disk6239StatT,
	mtbChan chan devices.MtStatT) {

	var (
		cpuStats                               cpuStatT
		lastIcount, iCount                     uint64
		ips, dpfIops, dskpIops                 float64
		lastCPUtime, lastDpfTime, lastDskpTime time.Time
		thisDpfIOcnt, lastDpfIOcnt             uint64
		thisDskpIOcnt, lastDskpIOcnt           uint64
		dpfStats                               devices.Disk6061StatT
		dskpStats                              devices.Disk6239StatT
		mtStats                                devices.MtStatT
	)

	l, err := net.Listen("tcp", "localhost:"+statusPort)
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

		statusSendString(conn, fmt.Sprintf("%c                             %c%s Status%c\012", dasherERASEPAGE, dasherUNDERLINE, appName, dasherNORMAL))

		for {
			// blocking wait for a status update to arrive
			select {
			case cpuStats = <-cpuChan:
				iCount = cpuStats.instrCount - lastIcount
				lastIcount = cpuStats.instrCount
				ips = float64(iCount) / (time.Since(lastCPUtime).Seconds() * 1000)
				lastCPUtime = time.Now()
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statCPUrow, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("PC:  %011o   Interrupts: %s    ATU: %s     IPS: %.fk/sec",
					cpuStats.pc,
					memory.BoolToOnOff(cpuStats.ion),
					memory.BoolToOnOff(cpuStats.atu),
					ips))
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statCPUrow2, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("AC0: %011o   AC1: %011o   AC2: %011o   AC3: %011o",
					cpuStats.ac[0],
					cpuStats.ac[1],
					cpuStats.ac[2],
					cpuStats.ac[3]))
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statInternalsRow, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("        Version: %s (%s) built with %s",
					Version, ReleaseType,
					cpuStats.goVersion))
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statInternalsRow2, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("        Host CPUs: %d  Goroutines: %d  Heap: %dMB",
					cpuStats.hostCPUCount,
					cpuStats.goroutineCount,
					cpuStats.heapSizeMB))

			case dpfStats = <-dpfChan:
				thisDpfIOcnt = dpfStats.Writes + dpfStats.Reads
				dpfIops = float64(thisDpfIOcnt-lastDpfIOcnt) / time.Since(lastDpfTime).Seconds()
				lastDpfIOcnt = thisDpfIOcnt
				lastDpfTime = time.Now()
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statDPFrow, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("DPF  (DPF0) - Attached: %c  IOPS: %.f CYL: %04d.  HD: %02d.  SECT: %03d.",
					memory.BoolToYN(dpfStats.ImageAttached),
					dpfIops,
					dpfStats.Cylinder,
					dpfStats.Head,
					dpfStats.Sector))

			case dskpStats = <-dskpChan:
				thisDskpIOcnt = dskpStats.Writes + dskpStats.Reads
				dskpIops = float64(thisDskpIOcnt-lastDskpIOcnt) / time.Since(lastDskpTime).Seconds()
				lastDskpIOcnt = thisDskpIOcnt
				lastDskpTime = time.Now()
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statDSKProw, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("DSKP (DPJ0) - Attached: %c  IOPS: %.f  SECNUM: %08d.",
					memory.BoolToYN(dskpStats.ImageAttached),
					dskpIops,
					//dskpStats.cylinder,
					//dskpStats.head,
					//dskpStats.sector,
					dskpStats.SectorNo))

			case mtStats = <-mtbChan:
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statMTrow, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("MTA  (MTC0) - Attached: %c  Mem Addr: %06o  Curr Cmd: %d",
					memory.BoolToYN(mtStats.ImageAttached[0]),
					mtStats.MemAddrReg,
					mtStats.CurrentCmd))
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", dasherWRITEWINDOWADDR, 0, statMTrow2, dasherERASEEOL))
				statusSendString(conn, fmt.Sprintf("              Image file: %s", mtStats.FileName[0]))
			}
		}
	}
}

func statusSendString(con net.Conn, s string) {
	con.Write([]byte(s))
}
