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
	devMax = 0100

	devBMC   = 005
	devCPU   = 077
	devDPF   = 027
	devDSKP  = 024
	devFPU   = 076
	devMAP   = 003
	devMTB   = 022
	devPIT   = 043
	devPSC   = 004
	devPWRFL = 0
	devRTC   = 014
	devSCP   = 045
	devTTI   = 010
	devTTO   = 011
	devWCS   = 001
	pmbCPU   = 0 // kinda!
	pmbDPF   = 7
	pmbDSKP  = 7
	pmbMTB   = 10
	pmbPSC   = 13
	pmbRTC   = 13
	pmbSCP   = 15
	pmbTTI   = 14
	pmbTTO   = 15

	cpuModelNo = 0x224C
	ucodeRev   = 0x04
)

// deviceToString is used for disassembly
func deviceToString(devNum int) string {
	var ds string
	switch devNum {
	case devBMC:
		return "BMC"
	case devCPU:
		return "CPU"
	case devDSKP:
		return "DSKP" // This is the Assembler mnemonic, the AOS/VS Mnemonic is "DPJ"
	case devDPF:
		return "DPF"
	case devFPU:
		return "FPU"
	case devMAP:
		return "MAP"
	case devMTB:
		return "MTB"
	case devPIT:
		return "PIT"
	case devPSC:
		return "PSC"
	case devPWRFL:
		return "PWRFAIL"
	case devRTC:
		return "RTC"
	case devSCP:
		return "SCP"
	case devTTI:
		return "TTI"
	case devTTO:
		return "TTO"
	case devWCS:
		return "WCS"
	default:
		ds = fmt.Sprintf("%#o", devNum)
	}
	return ds
}
