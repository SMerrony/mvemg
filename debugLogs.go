// debugLogs.go
package main

import (
	"fmt"
	//"io"
	//"log"
	"os"
)

const (
	DEBUG_LOGS      = 4
	DEBUG_LOG_LINES = 40000
	DEBUG_LOG       = 0
	DPF_LOG         = 1
	DSKP_LOG        = 2
	MAP_LOG         = 3

	LOG_PERMS = 0644
)

var (
	logArr    [DEBUG_LOGS][DEBUG_LOG_LINES]string
	firstLine [DEBUG_LOGS]int
	lastLine  [DEBUG_LOGS]int

	// DebugLog, DPFlog, DSKPlog, MAPlog *log.Logger
)

//func debugLogsInit() {

//	// general debugging output to STDOUT
//	DebugLog = log.New(os.Stdout, "DEBUG: ", log.Ltime|log.Lshortfile)

//	dpfLogHandle, _ := os.OpenFile("dpf_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, LOG_PERMS)
//	DPFlog = log.New(dpfLogHandle, "DPF: ", log.Ltime)

//	dskpLogHandle, _ := os.OpenFile("dskp_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, LOG_PERMS)
//	DSKPlog = log.New(dskpLogHandle, "DSKP: ", log.Ltime)

//	mapLogHandle, _ := os.OpenFile("map_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, LOG_PERMS)
//	MAPlog = log.New(mapLogHandle, "MAP: ", log.Ltime|log.Lshortfile)
//}

func debugLogsDump() {

	var (
		debugDumpFile *os.File
		//err           error
	)

	for l := range logArr {
		if firstLine[l] != lastLine[l] { // ignore unused or empty logs
			switch l {
			case DEBUG_LOG:
				debugDumpFile, _ = os.OpenFile("mvem_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, LOG_PERMS)
			case DPF_LOG:
				debugDumpFile, _ = os.OpenFile("dpf_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, LOG_PERMS)
			case DSKP_LOG:
				debugDumpFile, _ = os.OpenFile("dskp_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, LOG_PERMS)
			case MAP_LOG:
				debugDumpFile, _ = os.OpenFile("map_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, LOG_PERMS)
			}
			debugDumpFile.WriteString(">>> Dumping Debug Log\n\n")
			thisLine := firstLine[l]
			for thisLine != lastLine[l] {
				debugDumpFile.WriteString(logArr[l][thisLine])
				thisLine++
				if thisLine == DEBUG_LOG_LINES {
					thisLine = 0
				}
			}
			debugDumpFile.WriteString(logArr[l][thisLine])
			debugDumpFile.WriteString("\n>>> Debug Log Ends\n")
			debugDumpFile.Close()
		}
	}
}

func debugPrint(log int, aFmt string, msg ...interface{}) {

	lastLine[log]++
	if lastLine[log] == DEBUG_LOG_LINES {
		lastLine[log] = 0
	}

	if lastLine[log] == firstLine[log] {
		firstLine[log]++
		if firstLine[log] == DEBUG_LOG_LINES {
			firstLine[log] = 0
		}
	}

	logArr[log][lastLine[log]] = fmt.Sprintf(aFmt, msg...)
}
