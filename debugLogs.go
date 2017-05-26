// debugLogs.go
package main

import (
	"os"
)

const (
	DEBUG_LOGS      = 4
	DEBUG_LOG_LINES = 40000
	SYSTEM_LOG      = 0
	DPF_LOG         = 1
	DSKP_LOG        = 2
	MAP_LOG         = 3
)

var (
	logArr    [DEBUG_LOGS][DEBUG_LOG_LINES]string
	firstLine [DEBUG_LOGS]int
	lastLine  [DEBUG_LOGS]int
)

func debugLogsDump() {

	var (
		debugDumpFile *os.File
		//err           error
	)

	for l := range logArr {
		if firstLine[l] != lastLine[l] { // ignore unused or empty logs
			switch l {
			case SYSTEM_LOG:
				debugDumpFile, _ = os.OpenFile("mvem_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
			case DPF_LOG:
				debugDumpFile, _ = os.OpenFile("dpf_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
			case DSKP_LOG:
				debugDumpFile, _ = os.OpenFile("dskp_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
			case MAP_LOG:
				debugDumpFile, _ = os.OpenFile("map_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
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

func debugPrint(log int, msg string) {
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

	logArr[log][lastLine[log]] = msg
}
