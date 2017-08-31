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

// SimhTapesT contains the associated (ATTached) image file
// name and handle for each tape
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

// Attach associated a SimH tape image file with a virutal tape
func (st *SimhTapesT) Attach(tNum int, imgName string) bool {
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

// Rewind simulates a tape rewind by seeking to start of SimH tape image file
func (st *SimhTapesT) Rewind(tNum int) bool {
	logging.DebugPrint(logging.DebugLog, "INFO: simhTapeRewind called for tape #%d\n", tNum)
	_, err := st[tNum].simhFile.Seek(0, 0)
	if err != nil {
		logging.DebugPrint(logging.DebugLog, "ERROR: Could not rewind simH Tape Image file: %s, due to: %s\n", st[tNum].fileName, err.Error())
		return false
	}
	return true
}

// ReadRecordHeader reads a 4-byte SimH header/trailer record
func (st *SimhTapesT) ReadRecordHeader(tNum int) (dg.DwordT, bool) {
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

// WriteRecordHeader writes a 4-byte header/trailer
func (st *SimhTapesT) WriteRecordHeader(tNum int, hdr dg.DwordT) bool {
	hdrBytes := make([]byte, 4)
	hdrBytes[3] = byte(hdr >> 24)
	hdrBytes[2] = byte(hdr >> 16)
	hdrBytes[1] = byte(hdr >> 8)
	hdrBytes[0] = byte(hdr)
	nb, err := st[tNum].simhFile.Write(hdrBytes)
	if err != nil || nb != 4 {
		logging.DebugPrint(logging.DebugLog, "ERROR: Could not write simh tape header record due to %s\n", err.Error())
		return false
	}
	return true
}

// ReadRecordData attempts to read a data record from SimH tape image, fails if wrong number of bytes read
// N.B. does not read the header and trailer
func (st *SimhTapesT) ReadRecordData(tNum int, byteLen int) ([]byte, bool) {
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

// WriteRecordData writes the actual data - not the header/trailer
func (st *SimhTapesT) WriteRecordData(tNum int, rec []byte) bool {
	nb, err := st[tNum].simhFile.Write(rec)
	if err != nil {
		logging.DebugPrint(logging.DebugLog, "ERROR: Could not write simh tape record due to %s\n", err.Error())
		return false
	}
	if nb != len(rec) {
		logging.DebugPrint(logging.DebugLog, "ERROR: Could not write complete header record (Wrote %d of %d bytes)\n", nb, len(rec))
		return false
	}
	return true
}

// ReadCompleteRecord reads a header-record-trailer sequence and returns the data if OK
func (st *SimhTapesT) ReadCompleteRecord(tNum int) ([]byte, bool) {
	var (
		hdr, trlr       dg.DwordT
		hdrInt, trlrInt int
		ok              bool
		rec             []byte
	)
	hdr, ok = st.ReadRecordHeader(tNum)
	if !ok {
		return nil, false
	}
	hdrInt = int(int32(hdr))
	if hdrInt < 0 {
		logging.DebugPrint(logging.DebugLog, "ERROR: Tape record header indicates presence of error in block\n")
		return nil, false
	}
	rec, ok = st.ReadRecordData(tNum, hdrInt)
	if !ok {
		return nil, false
	}
	if hdrInt != len(rec) {
		logging.DebugPrint(logging.DebugLog, "ERROR: Tape record block length does not match header (%d vs. %d)\n", len(rec), hdrInt)
		return nil, false
	}
	trlr, ok = st.ReadRecordHeader(tNum)
	if !ok {
		return nil, false
	}
	trlrInt = int(int32(trlr))
	if hdrInt != trlrInt {
		logging.DebugPrint(logging.DebugLog, "ERROR: Tape record trailer does not match header (%d vs. %d)\n", trlrInt, hdrInt)
		return nil, false
	}
	return rec, true
}

// SpaceFwd advances the virtual tape by the specifield amount (0 means 1 whole file)
func (st *SimhTapesT) SpaceFwd(tNum int, recCnt int) bool {

	var hdr, trailer dg.DwordT
	done := false
	logging.DebugPrint(logging.DebugLog, "DEBUG: simhTapesTpaceFwd called for %d records\n", recCnt)

	// special case when recCnt == 0 which means space forward one file...
	if recCnt == 0 {
		for !done {
			hdr, _ = st.ReadRecordHeader(tNum)
			if hdr == simhMtrTmk {
				done = true
			} else {
				// read record and throw it away
				st.ReadRecordData(tNum, int(hdr))
				// read trailer
				trailer, _ = st.ReadRecordHeader(tNum)
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

// ScanImage - This function is available to the SCP emulator so that the user may determine if an
//   attached tape image makes any sense.  It also serves to test the simhTapeReadRecordHeader() and
//   simhTapeReadRecord() functions to some extent.
func (st *SimhTapesT) ScanImage(tNum int) string {
	var res string
	if st[tNum].simhFile == nil {
		return "\012 *** No Tape Image Attached ***"
	}
	res = fmt.Sprintf("\012Scanning attached tape image: %s...", st[tNum].fileName)

	res += "\012Rewinding..."
	st.Rewind(tNum)
	res += "done!"

	var fileSize, markCount, fileCount, recNum int
	var hdr, trailer dg.DwordT
	fileCount = -1

loop:
	for {
		hdr, _ = st.ReadRecordHeader(tNum)
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
			st.ReadRecordData(tNum, int(hdr)) // read record and throw away
			trailer, _ = st.ReadRecordHeader(tNum)
			//logging.DebugPrint(logging.DEBUG_LOG,"Debug: got trailer value: %d\n", trailer)
			if hdr == trailer {
				fileSize += int(hdr)
			} else {
				res += "\012Non-matching trailer found."
			}
		}
	}
	st.Rewind(tNum)
	return res
}
