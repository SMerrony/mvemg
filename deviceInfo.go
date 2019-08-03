// deviceInfo.go

// Copyright (C) 2017,2019  Steve Merrony

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
// N.B. add to deviceMap below when new codes added here
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
)

var deviceMap = devices.DeviceMapT{
	devPWRFL: {DgMnemonic: "PWRFL", PMB: 0, IsIO: true, IsBootable: false},
	devWCS:   {DgMnemonic: "WCS", PMB: 99, IsIO: true, IsBootable: false},
	devMAP:   {DgMnemonic: "MAP", PMB: 99, IsIO: true, IsBootable: false},
	devPSC:   {DgMnemonic: "PSC", PMB: 13, IsIO: false, IsBootable: false},
	devBMC:   {DgMnemonic: "BMC", PMB: 99, IsIO: true, IsBootable: false},
	devTTI:   {DgMnemonic: "TTI", PMB: 14, IsIO: true, IsBootable: false},
	devTTO:   {DgMnemonic: "TTO", PMB: 15, IsIO: true, IsBootable: false},
	devRTC:   {DgMnemonic: "RTC", PMB: 13, IsIO: true, IsBootable: false},
	devLPT:   {DgMnemonic: "LPT", PMB: 12, IsIO: true, IsBootable: false},
	devMTB:   {DgMnemonic: "MTB", PMB: 10, IsIO: true, IsBootable: true},
	devMTJ:   {DgMnemonic: "MTJ", PMB: 10, IsIO: true, IsBootable: true},
	devDSKP:  {DgMnemonic: "DSKP", PMB: 7, IsIO: true, IsBootable: true},
	devDPF:   {DgMnemonic: "DPF", PMB: 7, IsIO: true, IsBootable: true},
	devISC:   {DgMnemonic: "ISC", PMB: 4, IsIO: true, IsBootable: false},
	devPIT:   {DgMnemonic: "PIT", PMB: 11, IsIO: false, IsBootable: false},
	devSCP:   {DgMnemonic: "SCP", PMB: 15, IsIO: true, IsBootable: false},
	devIAC1:  {DgMnemonic: "IAC1", PMB: 11, IsIO: true, IsBootable: false},
	devMTB1:  {DgMnemonic: "MTB1", PMB: 10, IsIO: true, IsBootable: true},
	devMTJ1:  {DgMnemonic: "MTJ1", PMB: 10, IsIO: true, IsBootable: true},
	devDSKP1: {DgMnemonic: "DSKP1", PMB: 7, IsIO: true, IsBootable: true},
	devIAC:   {DgMnemonic: "IAC", PMB: 11, IsIO: true, IsBootable: false},
	devDPF1:  {DgMnemonic: "DPF1", PMB: 7, IsIO: true, IsBootable: true},
	devFPU:   {DgMnemonic: "FPU", PMB: 99, IsIO: true, IsBootable: false},
	devCPU:   {DgMnemonic: "CPU", PMB: 0, IsIO: true, IsBootable: false},
}

// deviceToString is used for disassembly
// If the device code is known, then an assember mnemonic is returned, otherwise just the code
func deviceToString(devNum int) string {
	de, known := deviceMap[devNum]
	if known {
		return de.DgMnemonic
	}
	return fmt.Sprintf("%#o", devNum)
}
