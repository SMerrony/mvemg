// statusCollector.go
package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

const (
	// define which screen row each of the monotored data appear on
	STAT_CPU_ROW  = 3
	STAT_CPU_ROW2 = 5
	STAT_DPF_ROW  = 7
	STAT_DSKP_ROW = 9
)

// StatusCollector maintains a near real-time status screen available on STAT_PORT.
// The screen uses DG DASHER control codes for formatting, so a DASHER terminal emulator
// should be attached to it for good results.
// The function (which is intended to be run as a goroutine) listens for status updates
// from known senders and upon receiving an update refreshes the display of that status
// on the monitor page.
func statusCollector(
	cpuChan chan cpuStatT,
	dpfChan chan dpfStatT,
	dskpChan chan dskpStatT) {

	var (
		cpuStats                cpuStatT
		lastIcount, iCount, ips uint64
		dpfStats                dpfStatT
		dskpStats               dskpStatT
	)

	l, err := net.Listen("tcp", "localhost:"+STAT_PORT)
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

		statusSendString(conn, fmt.Sprintf("%c                        MV/Em Status\012", DASHER_ERASE_PAGE))
		statusSendString(conn, "                        ============")

		for {
			// blocking wait for a status update to arrive
			select {
			case cpuStats = <-cpuChan:
				iCount = cpuStats.instrCount - lastIcount
				lastIcount = cpuStats.instrCount
				ips = iCount / 100 // to get kIPS
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", DASHER_WRITE_WINDOW_ADDR, 0, STAT_CPU_ROW, DASHER_ERASE_EOL))
				statusSendString(conn, fmt.Sprintf("PC:  %010d   Interrupts: %d    ATU: %d     IPS: %dk/sec",
					cpuStats.pc,
					boolToInt(cpuStats.ion),
					boolToInt(cpuStats.atu),
					ips))
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", DASHER_WRITE_WINDOW_ADDR, 0, STAT_CPU_ROW2, DASHER_ERASE_EOL))
				statusSendString(conn, fmt.Sprintf("AC0: %010d   AC1: %010d   AC2: %010d   AC3: %010d",
					cpuStats.ac[0],
					cpuStats.ac[1],
					cpuStats.ac[2],
					cpuStats.ac[3]))
			case dpfStats = <-dpfChan:
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", DASHER_WRITE_WINDOW_ADDR, 0, STAT_DPF_ROW, DASHER_ERASE_EOL))
				statusSendString(conn, fmt.Sprintf("DPF  - Attached: %d  CYL: %04d  HD: %02d  SECT: %03d",
					boolToInt(dpfStats.imageAttached),
					dpfStats.cylinder,
					dpfStats.head,
					dpfStats.sector))

			case dskpStats = <-dskpChan:
				statusSendString(conn, fmt.Sprintf("%c%c%c%c", DASHER_WRITE_WINDOW_ADDR, 0, STAT_DSKP_ROW, DASHER_ERASE_EOL))
				statusSendString(conn, fmt.Sprintf("DSKP - Attached: %d  CYL: %04d  HD: %02d  SECT: %03d  SECNUM: %08d",
					boolToInt(dskpStats.imageAttached),
					dskpStats.cylinder,
					dskpStats.head,
					dskpStats.sector,
					dskpStats.sectorNo))
			}
		}
	}
}

func statusSendString(con net.Conn, s string) {
	con.Write([]byte(s))
}
