// novaOp.go

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
)

func novaOp(cpuPtr *CPU, iPtr *decodedInstrT) bool {

	var (
		shifter          dg_word
		wideShifter      dg_dword
		tmpAcS, tmpAcD   dg_word
		savedCry, tmpCry bool
		pcInc            dg_phys_addr
	)

	tmpAcS = dwordGetLowerWord(cpuPtr.ac[iPtr.acs])
	tmpAcD = dwordGetLowerWord(cpuPtr.ac[iPtr.acd])
	savedCry = cpuPtr.carry

	// Preset Carry if required
	switch iPtr.c {
	case 'Z': // zero
		cpuPtr.carry = false
	case 'O': // One
		cpuPtr.carry = true
	case 'C': // Complement
		cpuPtr.carry = !cpuPtr.carry
	}

	// perform the operation
	switch iPtr.mnemonic {
	case "ADC":
		wideShifter = dg_dword(tmpAcD + (^tmpAcS))
		shifter = dwordGetLowerWord(wideShifter)
		if wideShifter > 65535 {
			cpuPtr.carry = !cpuPtr.carry
		} else {
			cpuPtr.carry = false
		}

	case "ADD": // unsigned
		wideShifter = dg_dword(tmpAcD + tmpAcS)
		shifter = dwordGetLowerWord(wideShifter)
		if wideShifter > 65535 {
			cpuPtr.carry = !cpuPtr.carry
		} else {
			cpuPtr.carry = false
		}

	case "AND":
		shifter = tmpAcD & tmpAcS

	case "COM":
		shifter = ^tmpAcS

	case "INC":
		shifter = tmpAcS + 1
		if tmpAcS == 0xffff {
			cpuPtr.carry = !cpuPtr.carry
		}

	case "MOV":
		shifter = tmpAcS

	case "NEG":
		shifter = dg_word(-int16(tmpAcS))
		if tmpAcS == 0 {
			cpuPtr.carry = !cpuPtr.carry
		}

	case "SUB":
		shifter = tmpAcD - tmpAcS
		if tmpAcS <= tmpAcD {
			cpuPtr.carry = !cpuPtr.carry
		}

	default:
		log.Fatalf("ERROR: NOVA_MEMREF instruction <%s> not yet implemented\n", iPtr.mnemonic)
	}

	// shift if required
	switch iPtr.sh {
	case 'L':
		tmpCry = cpuPtr.carry
		cpuPtr.carry = testWbit(shifter, 0)
		shifter <<= 1
		if tmpCry {
			shifter |= 0x0001
		}
	case 'R':
		tmpCry = cpuPtr.carry
		cpuPtr.carry = testWbit(shifter, 15)
		shifter >>= 1
		if tmpCry {
			shifter |= 0x8000
		}
	case 'S':
		shifter = swapBytes(shifter)
	}

	// Skip?
	switch iPtr.skip {
	case "NONE":
		pcInc = 1
	case "SKP":
		pcInc = 2
	case "SZC":
		if !cpuPtr.carry {
			pcInc = 2
		} else {
			pcInc = 1
		}
	case "SNC":
		if cpuPtr.carry {
			pcInc = 2
		} else {
			pcInc = 1
		}
	case "SZR":
		if shifter == 0 {
			pcInc = 2
		} else {
			pcInc = 1
		}
	case "SNR":
		if shifter != 0 {
			pcInc = 2
		} else {
			pcInc = 1
		}
	case "SEZ":
		if !cpuPtr.carry || shifter == 0 {
			pcInc = 2
		} else {
			pcInc = 1
		}
	case "SBN":
		if cpuPtr.carry && shifter != 0 {
			pcInc = 2
		} else {
			pcInc = 1
		}
	default:
		log.Fatalln("ERROR: Invalid skip in novaOp()")
	}

	// No-Load?
	if iPtr.nl != '#' {
		cpuPtr.ac[iPtr.acd] = dg_dword(shifter) & 0x0000ffff
	} else {
		// don't load the result from the shifter, restore the Carry flag
		cpuPtr.carry = savedCry
	}

	cpuPtr.pc += pcInc
	return true
}
