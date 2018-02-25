// novaMemRef.go

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
	"mvemg/memory"
	"mvemg/util"
)

func novaMemRef(cpuPtr *CPUT, iPtr *decodedInstrT) bool {

	var (
		shifter           dg.WordT
		effAddr           dg.PhysAddrT
		novaNoAccEffAddr  novaNoAccEffAddrT
		novaOneAccEffAddr novaOneAccEffAddrT
	)

	switch iPtr.ix {

	case instrDSZ:
		novaNoAccEffAddr = iPtr.variant.(novaNoAccEffAddrT)
		effAddr = resolve16bitEclipseAddr(cpuPtr, novaNoAccEffAddr.ind, novaNoAccEffAddr.mode, novaNoAccEffAddr.disp15)
		shifter = memory.ReadWord(effAddr)
		shifter--
		memory.WriteWord(effAddr, shifter)
		if shifter == 0 {
			cpuPtr.pc++
		}

	case instrISZ:
		novaNoAccEffAddr = iPtr.variant.(novaNoAccEffAddrT)
		effAddr = resolve16bitEclipseAddr(cpuPtr, novaNoAccEffAddr.ind, novaNoAccEffAddr.mode, novaNoAccEffAddr.disp15)
		shifter = memory.ReadWord(effAddr)
		shifter++
		memory.WriteWord(effAddr, shifter)
		if shifter == 0 {
			cpuPtr.pc++
		}

	case instrLDA:
		novaOneAccEffAddr = iPtr.variant.(novaOneAccEffAddrT)
		effAddr = resolve16bitEclipseAddr(cpuPtr, novaOneAccEffAddr.ind, novaOneAccEffAddr.mode, novaOneAccEffAddr.disp15)
		shifter = memory.ReadWord(effAddr)
		cpuPtr.ac[novaOneAccEffAddr.acd] = 0x0000ffff & dg.DwordT(shifter)

	case instrSTA:
		novaOneAccEffAddr = iPtr.variant.(novaOneAccEffAddrT)
		shifter = util.DwordGetLowerWord(cpuPtr.ac[novaOneAccEffAddr.acd])
		effAddr = resolve16bitEclipseAddr(cpuPtr, novaOneAccEffAddr.ind, novaOneAccEffAddr.mode, novaOneAccEffAddr.disp15)
		memory.WriteWord(effAddr, shifter)

	default:
		log.Fatalf("ERROR: NOVA_MEMREF instruction <%s> not yet implemented\n", iPtr.mnemonic)
		return false
	}
	cpuPtr.pc++
	return true
}
