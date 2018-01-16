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

import "fmt"

// Device IDs and PMBs
// Standard device codes in octal, Priority Mask Bits in decimal
// as per DG docs!
// N.B. add to deviceToString() when new codes added here
const (
	devMax  = 0100
	devPSC  = 004
	pscPMB  = 13
	devTTI  = 010
	ttiPMB  = 14
	devTTO  = 011
	ttoPMB  = 15
	devRTC  = 014
	rtcPMB  = 13
	devMTB  = 022
	mtbPMB  = 10
	devDSKP = 024
	dskpPMB = 7
	devDPF  = 027
	dpfPMB  = 7
	devSCP  = 045
	scpPMB  = 15
	devCPU  = 077
	cpuPMB  = 0 // kinda!

	cpuModelNo = 0x224C
	ucodeRev   = 0x04
)

func deviceToString(devNum int) string {
	var ds string
	switch devNum {
	case devCPU:
		return "CPU"
	case devDSKP:
		return "DSKP"
	case devDPF:
		return "DPF"
	case devMTB:
		return "MTB"
		//	case devPIT:
		//		return "PIT"
	case devPSC:
		return "PSC"
	case devRTC:
		return "RTC"
	case devSCP:
		return "SCP"
	case devTTI:
		return "TTI"
	case devTTO:
		return "TTO"
	default:
		ds = fmt.Sprintf("%#o", devNum)
	}
	return ds
}
