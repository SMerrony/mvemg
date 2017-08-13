// simhTape.go
package main

//"io/ioutil"
//"os"

import (
	"fmt"
	"log"
	"mvemg/dg"
	"mvemg/logging"
	"os"
)

const (
	MTR_TMK          = 0          /* tape mark */
	MTR_EOM          = 0xFFFFFFFF /* end of medium */
	MTR_GAP          = 0xFFFFFFFE /* primary gap */
	MTR_MAXLEN       = 0x00FFFFFF /* max len is 24b */
	MTR_ERF          = 0x80000000 /* error flag */
	MAX_SIMH_TAPES   = 8
	MAX_SIMH_REC_LEN = 32768
)

// SimhTapes contians
type SimhTapes [MAX_SIMH_TAPES]struct {
	fileName string
	simhFile *os.File
}

func (st *SimhTapes) simhTapeInit() {
	for t := range st {
		st[t].simhFile = nil
		st[t].fileName = ""
	}
}

func (st *SimhTapes) simhTapeAttach(tNum int, imgName string) bool {
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
func (st *SimhTapes) simhTapeRewind(tNum int) bool {
	logging.DebugPrint(logging.DebugLog, "INFO: simhTapeRewind called for tape #%d\n", tNum)
	_, err := st[tNum].simhFile.Seek(0, 0)
	if err != nil {
		logging.DebugPrint(logging.DebugLog, "ERROR: Could not rewind simH Tape Image file: %s, due to: %s\n", st[tNum].fileName, err.Error())
		return false
	}
	return true
}

// read a 4-byte SimH header/trailer record
func (st *SimhTapes) simhTapeReadRecordHeader(tNum int) (dg.DwordT, bool) {
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

// attempt to read record from SimH tape image, fail if wrong number of bytes read
func (st *SimhTapes) simhTapeReadRecord(tNum int, byteLen int) ([]byte, bool) {
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

func (st *SimhTapes) simhTapeSpaceFwd(tNum int, recCnt int) bool {

	var hdr, trailer dg.DwordT
	done := false
	logging.DebugPrint(logging.DebugLog, "DEBUG: simhTapeSpaceFwd called for %d records\n", recCnt)

	// special case when recCnt == 0 which means space forward one file...
	if recCnt == 0 {
		for !done {
			hdr, _ = st.simhTapeReadRecordHeader(tNum)
			if hdr == MTR_TMK {
				done = true
			} else {
				// read record and throw it away
				st.simhTapeReadRecord(tNum, int(hdr))
				// read trailer
				trailer, _ = st.simhTapeReadRecordHeader(tNum)
				if hdr != trailer {
					log.Fatal("ERROR: simhTapeSpaceFwd found non-matching header/trailer")
				}
			}
		}
	} else {
		log.Fatal("ERROR: simhTapeSpaceFwd called with record count != 0 - Not Yet Implemented")
	}

	return true
}

/* This function is available to the SCP emulator so that the user may determine if an
   attached tape image makes any sense.  It also serves to test the simhTapeReadRecordHeader() and
   simhTapeReadRecord() functions to some extent.
*/
func (st *SimhTapes) simhTapeScanImage(tNum int) string {
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
		case MTR_TMK:
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
		case MTR_EOM:
			res += "\012End of Medium"
			break loop
		case MTR_GAP:
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
