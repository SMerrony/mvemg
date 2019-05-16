// deviceInfo.go

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

	"github.com/SMerrony/dgemug/devices"
)

// Device IDs and PMBs
// Standard device codes in octal, Priority Mask Bits in decimal
// as per DG docs!
// N.B. add to deviceToString() when new codes added here
const (
	devMax = 0100

	devPWRFL = 000
	devWCS   = 001
	devMAP   = 003
	devPSC   = 004
	devBMC   = 005
	devTTI   = 010
	devTTO   = 011
	devRTC   = 014
	devLPT   = 017
	devMTB   = 022
	devMTJ   = 023
	devDSKP  = 024
	devDPF   = 027
	devISC   = 034
	devPIT   = 043
	devSCP   = 045
	devIAC1  = 050
	devMTB1  = 062
	devMTJ1  = 063
	devDSKP1 = 064
	devIAC   = 065
	devDPF1  = 067
	devFPU   = 076
	devCPU   = 077

	cpuModelNo = 0x224C
	ucodeRev   = 0x04
)

// deviceToString is used for disassembly
func deviceToString(devNum int) string {
	de, known := deviceMap[devNum]
	if known {
		return de.DgMnemonic
	}
	return fmt.Sprintf("%#o", devNum)
}

var deviceMap = devices.DeviceMapT{
	devPWRFL: {"PWRFL", 0, true, false},
	devWCS:   {"WCS", 99, true, false},
	devMAP:   {"MAP", 99, true, false},
	devPSC:   {"PSC", 13, false, false},
	devBMC:   {"BMC", 99, true, false},
	devTTI:   {"TTI", 14, true, false},
	devTTO:   {"TTO", 15, true, false},
	devRTC:   {"RTC", 13, true, false},
	devLPT:   {"LPT", 12, true, false},
	devMTB:   {"MTB", 10, true, true},
	devMTJ:   {"MTJ", 10, true, true},
	devDSKP:  {"DSKP", 7, true, true},
	devDPF:   {"DPF", 7, true, true},
	devISC:   {"ISC", 4, true, false},
	devPIT:   {"PIT", 11, false, false},
	devSCP:   {"SCP", 15, true, false},
	devIAC1:  {"IAC1", 11, true, false},
	devMTB1:  {"MTB1", 10, true, true},
	devMTJ1:  {"MTJ1", 10, true, true},
	devDSKP1: {"DSKP1", 7, true, true},
	devIAC:   {"IAC", 11, true, false},
	devDPF1:  {"DPF1", 7, true, true},
	devFPU:   {"FPU", 99, true, false},
	devCPU:   {"CPU", 0, true, false},
}
