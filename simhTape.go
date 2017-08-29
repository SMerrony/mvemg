// simhTape.go

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
	"mvemg/dg"
	"mvemg/logging"
	"os"
)

const (
	simhMtrTmk    = 0          /* tape mark */
	simhMtrEom    = 0xFFFFFFFF /* end of medium */
	simhMtrGap    = 0xFFFFFFFE /* primary gap */
	simhMtrMaxlen = 0x00FFFFFF /* max len is 24b */
	simhMtrErf    = 0x80000000 /* error flag */
	maxSimhTapes  = 8
	maxSimhRecLen = 32768
)

// SimhTapesT contians
type SimhTapesT [maxSimhTapes]struct {
	fileName string
	simhFile *os.File
}

func (st *SimhTapesT) simhTapeInit() {
	for t := range st {
		st[t].simhFile = nil
		st[t].fileName = ""
	}
}

func (st *SimhTapesT) simhTapeAttach(tNum int, imgName string) bool {
	logging.DebugPrint(logging.DebugLog, "INFO: simhTapeAttach called for tape #%d with image <%s>\n", tNum, imgName)
	f, err := os.Open(imgName)
	if err != nil {
		logging.DebugPrint(logging.DebugLog, "ERROR: Could not open simH Tape Image file: %s, due to: %s\n", imgName, err.Error())
		return false
	}
	st[tNum].fileName = imgName
	st[tNum].simhFile = f
	return true
}

// simulate a tape rewind by seeking to start of SimH tape image file
func (st *SimhTapesT) simhTapeRewind(tNum int) bool {
	logging.DebugPrint(logging.DebugLog, "INFO: simhTapeRewind called for tape #%d\n", tNum)
	_, err := st[tNum].simhFile.Seek(0, 0)
	if err != nil {
		logging.DebugPrint(logging.DebugLog, "ERROR: Could not rewind simH Tape Image file: %s, due to: %s\n", st[tNum].fileName, err.Error())
		return false
	}
	return true
}

// read a 4-byte SimH header/trailer record
func (st *SimhTapesT) simhTapeReadRecordHeader(tNum int) (dg.DwordT, bool) {
	hdrBytes := make([]byte, 4)
	nb, err := st[tNum].simhFile.Read(hdrBytes)
	if err != nil {
		logging.DebugPrint(logging.DebugLog, "ERROR: Could not read simH Tape Image record header: %s, due to: %s\n", st[tNum].fileName, err.Error())
		return 0, false
	}
	if nb != 4 {
		logging.DebugPrint(logging.DebugLog, "ERROR: Wrong length simH Tape Image record header: %d\n", nb)
		return 0, false
	}
	//logging.DebugPrint(logging.DEBUG_LOG,"Debug - Header bytes: %d %d %d %d\n", hdrBytes[0], hdrBytes[1], hdrBytes[2], hdrBytes[3])
	var hdr dg.DwordT
	hdr = dg.DwordT(hdrBytes[3]) << 24
	hdr |= dg.DwordT(hdrBytes[2]) << 16
	hdr |= dg.DwordT(hdrBytes[1]) << 8
	hdr |= dg.DwordT(hdrBytes[0])
	return hdr, true
}

func (st *SimhTapesT) simhTapeWriteRecordHeader(tNum int, hdr dg.DwordT) bool {
	hdrBytes := make([]byte, 4)
	hdrBytes[3] = byte(hdr >> 24)
	hdrBytes[2] = byte(hdr >> 16)
	hdrBytes[1] = byte(hdr >> 8)
	hdrBytes[0] = byte(hdr)
	nb, err := st[tNum].simhFile.Write(hdrBytes)
	if err != nil || nb != 4 {
		logging.DebugPrint(logging.DebugLog, "ERROR: Could not write header record due to %s\n", err.Error())
		return false
	}
	return true
}

// attempt to read record from SimH tape image, fail if wrong number of bytes read
func (st *SimhTapesT) simhTapeReadRecord(tNum int, byteLen int) ([]byte, bool) {
	rec := make([]byte, byteLen)
	nb, err := st[tNum].simhFile.Read(rec)
	if err != nil {
		logging.DebugPrint(logging.DebugLog, "ERROR: Could not read simH Tape Image %s record due to: %s\n", st[tNum].fileName, err.Error())
		return nil, false
	}
	if nb != byteLen {
		logging.DebugPrint(logging.DebugLog, "ERROR: Could not read simH Tape Image %s record, got %d bytes, expecting %d\n", st[tNum].fileName, nb, byteLen)
		return nil, false
	}
	return rec, true
}

// SimhTapeSpaceFwd advances the virtual tape by the specifield amount (0 means 1 whole file)
func (st *SimhTapesT) SimhTapeSpaceFwd(tNum int, recCnt int) bool {

	var hdr, trailer dg.DwordT
	done := false
	logging.DebugPrint(logging.DebugLog, "DEBUG: simhTapesTpaceFwd called for %d records\n", recCnt)

	// special case when recCnt == 0 which means space forward one file...
	if recCnt == 0 {
		for !done {
			hdr, _ = st.simhTapeReadRecordHeader(tNum)
			if hdr == simhMtrTmk {
				done = true
			} else {
				// read record and throw it away
				st.simhTapeReadRecord(tNum, int(hdr))
				// read trailer
				trailer, _ = st.simhTapeReadRecordHeader(tNum)
				if hdr != trailer {
					log.Fatal("ERROR: simhTapesTpaceFwd found non-matching header/trailer")
				}
			}
		}
	} else {
		log.Fatal("ERROR: simhTapesTpaceFwd called with record count != 0 - Not Yet Implemented")
	}

	return true
}

// SimhTapeScanImage - This function is available to the SCP emulator so that the user may determine if an
//   attached tape image makes any sense.  It also serves to test the simhTapeReadRecordHeader() and
//   simhTapeReadRecord() functions to some extent.
func (st *SimhTapesT) SimhTapeScanImage(tNum int) string {
	var res string
	if st[tNum].simhFile == nil {
		return "\012 *** No Tape Image Attached ***"
	}
	res = fmt.Sprintf("\012Scanning attached tape image: %s...", st[tNum].fileName)

	res += "\012Rewinding..."
	st.simhTapeRewind(tNum)
	res += "done!"

	var fileSize, markCount, fileCount, recNum int
	var hdr, trailer dg.DwordT
	fileCount = -1

loop:
	for {
		hdr, _ = st.simhTapeReadRecordHeader(tNum)
		//logging.DebugPrint(logging.DEBUG_LOG,"Debug: got header value: %d\n", hdr)
		switch hdr {
		case simhMtrTmk:
			if fileSize > 0 {
				fileCount++
				res += fmt.Sprintf("\012File %d : %14d bytes in %7d block(s)",
					fileCount, fileSize, recNum)
				fileSize = 0
				recNum = 0
			}
			markCount++
			if markCount == 3 {
				res += "\012Triple Mark (old End Of Tape indicator)"
				break loop
			}
		case simhMtrEom:
			res += "\012End of Medium"
			break loop
		case simhMtrGap:
			res += "\012Erase Gap"
			markCount = 0
		default:
			recNum++
			markCount = 0
			st.simhTapeReadRecord(tNum, int(hdr)) // read record and throw away
			trailer, _ = st.simhTapeReadRecordHeader(tNum)
			//logging.DebugPrint(logging.DEBUG_LOG,"Debug: got trailer value: %d\n", trailer)
			if hdr == trailer {
				fileSize += int(hdr)
			} else {
				res += "\012Non-matching trailer found."
			}
		}
	}
	st.simhTapeRewind(tNum)
	return res
}
