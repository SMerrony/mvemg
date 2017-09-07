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
	"log"
	"mvemg/dg"
	"mvemg/logging"
	"os"

	"github.com/SMerrony/aosvs-tools/simhTape"
)

const maxSimhTapes = 8

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

// Attach associated a SimH tape image file with a virtual tape
func (st *SimhTapesT) Attach(tNum int, imgName string) bool {
	logging.DebugPrint(logging.MtbLog, "INFO: simhTapeAttach called for tape #%d with image <%s>\n", tNum, imgName)
	f, err := os.Open(imgName)
	if err != nil {
		logging.DebugPrint(logging.MtbLog, "ERROR: Could not open simH Tape Image file: %s, due to: %s\n", imgName, err.Error())
		return false
	}
	st[tNum].fileName = imgName
	st[tNum].simhFile = f
	return true
}

// Rewind simulates a tape rewind by seeking to start of SimH tape image file
func (st *SimhTapesT) Rewind(tNum int) bool {
	logging.DebugPrint(logging.MtbLog, "INFO: simhTapeRewind called for tape #%d\n", tNum)
	_, err := st[tNum].simhFile.Seek(0, 0)
	if err != nil {
		logging.DebugPrint(logging.MtbLog, "ERROR: Could not rewind simH Tape Image file: %s, due to: %s\n", st[tNum].fileName, err.Error())
		return false
	}
	return true
}

// ReadRecordHeader reads a 4-byte SimH header/trailer record
func (st *SimhTapesT) ReadRecordHeader(tNum int) (uint32, bool) {
	return simhTape.ReadMetaData(st[tNum].simhFile)
}

// WriteRecordHeader writes a 4-byte header/trailer
func (st *SimhTapesT) WriteRecordHeader(tNum int, hdr dg.DwordT) bool {
	return simhTape.WriteMetaData(st[tNum].simhFile, uint32(hdr))
}

// ReadRecordData attempts to read a data record from SimH tape image, fails if wrong number of bytes read
// N.B. does not read the header and trailer
func (st *SimhTapesT) ReadRecordData(tNum int, byteLen int) ([]byte, bool) {
	return simhTape.ReadRecordData(st[tNum].simhFile, byteLen)
}

// WriteRecordData writes the actual data - not the header/trailer
func (st *SimhTapesT) WriteRecordData(tNum int, rec []byte) bool {
	return simhTape.WriteRecordData(st[tNum].simhFile, rec)
}

// ReadCompleteRecord reads a header-record-trailer sequence and returns the data if OK
func (st *SimhTapesT) ReadCompleteRecord(tNum int) ([]byte, bool) {
	var (
		hdr, trlr       uint32
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
		logging.DebugPrint(logging.MtbLog, "ERROR: Tape record header indicates presence of error in block\n")
		return nil, false
	}
	rec, ok = st.ReadRecordData(tNum, hdrInt)
	if !ok {
		return nil, false
	}
	if hdrInt != len(rec) {
		logging.DebugPrint(logging.MtbLog, "ERROR: Tape record block length does not match header (%d vs. %d)\n", len(rec), hdrInt)
		return nil, false
	}
	trlr, ok = st.ReadRecordHeader(tNum)
	if !ok {
		return nil, false
	}
	trlrInt = int(int32(trlr))
	if hdrInt != trlrInt {
		logging.DebugPrint(logging.MtbLog, "ERROR: Tape record trailer does not match header (%d vs. %d)\n", trlrInt, hdrInt)
		return nil, false
	}
	return rec, true
}

// SpaceFwd advances the virtual tape by the specified amount (0 means 1 whole file)
func (st *SimhTapesT) SpaceFwd(tNum int, recCnt int) bool {

	var hdr, trailer uint32
	done := false
	logging.DebugPrint(logging.MtbLog, "DEBUG: simhTapesTpaceFwd called for %d records\n", recCnt)

	// special case when recCnt == 0 which means space forward one file...
	if recCnt == 0 {
		for !done {
			hdr, _ = st.ReadRecordHeader(tNum)
			if hdr == simhTape.SimhMtrTmk {
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
	return simhTape.ScanImage(st[tNum].fileName, false)
}
