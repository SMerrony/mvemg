// debugLogs.go
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

package logging

import (
	"fmt"
	"os"
	"sync"
)

const (
	numDebugLogs     = 5
	numDebugLogLines = 40000

	// DebugLog is the general-purpose log
	DebugLog = 0
	// MtbLog is for the MTB tape module
	MtbLog = 1
	// DpfLog is for the DPF disk module
	DpfLog = 2
	// DskpLog is for the DSKP disk module
	DskpLog = 3
	// MapLog is for BMC/DCH-related logging
	MapLog = 4

	logPerms = 0644
)

var (
	logArr    [numDebugLogs][numDebugLogLines]string // the stored log messages
	logArrMu  [numDebugLogs]sync.Mutex
	firstLine [numDebugLogs]int // pointer to the first line of each log
	lastLine  [numDebugLogs]int // pointer to the last line of each log
)

// DebugLogsDump can be called to dump out each of the non-empty debug logs to text files
func DebugLogsDump() {

	var (
		debugDumpFile *os.File
		//err           error
	)

	for l := range logArr {
		if firstLine[l] != lastLine[l] { // ignore unused or empty logs
			switch l {
			case DebugLog:
				debugDumpFile, _ = os.OpenFile("mvem_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, logPerms)
			case MtbLog:
				debugDumpFile, _ = os.OpenFile("mtb_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, logPerms)
			case DpfLog:
				debugDumpFile, _ = os.OpenFile("dpf_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, logPerms)
			case DskpLog:
				debugDumpFile, _ = os.OpenFile("dskp_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, logPerms)
			case MapLog:
				debugDumpFile, _ = os.OpenFile("map_debug.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, logPerms)
			}
			debugDumpFile.WriteString(">>> Dumping Debug Log\n\n")
			thisLine := firstLine[l]
			for thisLine != lastLine[l] {
				debugDumpFile.WriteString(logArr[l][thisLine])
				thisLine++
				if thisLine == numDebugLogLines {
					thisLine = 0
				}
			}
			debugDumpFile.WriteString(logArr[l][thisLine])
			debugDumpFile.WriteString("\n>>> Debug Log Ends\n")
			debugDumpFile.Close()
		}
	}
}

// DebugPrint doesn't print anything!  It stores the log message
// in array-backed circular arrays
// for printing when debugLogsDump() is invoked.
// This func can be called very often, KISS...
func DebugPrint(log int, aFmt string, msg ...interface{}) {

	logArrMu[log].Lock()

	lastLine[log]++

	// end of log array?
	if lastLine[log] == numDebugLogLines {
		lastLine[log] = 0 // wrap-around
	}

	// has the tail hit the head of the circular buffer?
	if lastLine[log] == firstLine[log] {
		firstLine[log]++ // advance the head pointer
		if firstLine[log] == numDebugLogLines {
			firstLine[log] = 0 // but reset if at limit
		}
	}

	// sprintf the given message to tail of the specified log
	logArr[log][lastLine[log]] = fmt.Sprintf(aFmt, msg...)
	logArrMu[log].Unlock()
}
