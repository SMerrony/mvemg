// instructions.go

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
	"bufio"
	"fmt"
	"mvemg/dg"
	"os"
)

// the characteristics of each instruction
type instrChars struct {
	// mnemonic   string  // DG standard assembler mnemonic for opcode
	bits      dg.WordT // bit-pattern for opcode
	mask      dg.WordT // mask for unique bits in opcode
	instrLen  int      // # of words in opcode and any following args
	instrFmt  int      // opcode layout
	instrType int      // class of opcode (somewhat arbitrary)
	//xeqCounter uint64  // count of # times instruction hit during this run
}

// InstructionSet contains the map of all recognised instruction.
// N.B. Recognised, not implemented necessarily.
type InstructionSet map[string]instrChars

var instructionSet = make(InstructionSet)

var ioFlags = [...]byte{' ', 'S', 'C', 'P'}
var ioTests = [...]string{"BN", "BZ", "DN", "DZ"}
var modes = [...]string{"Absolute", "PC", "AC2", "AC3"}
var skips = [...]string{"NONE", "SKP", "SZC", "SNC", "SZR", "SNR", "SEZ", "SBN"}

// debugging function...
func dumpCSV() {
	csvFilename := "/tmp/mvinstr.csv"
	csvFile, _ := os.Create(csvFilename)
	csvWriter := bufio.NewWriter(csvFile)
	fmt.Fprintf(csvWriter, ";\n;Instructions\n")
	for mnem := range instructionSet {
		fmt.Fprintf(csvWriter, "%s,0x%X,0x%X,%d,%d,%d\n",
			mnem,
			instructionSet[mnem].bits,
			instructionSet[mnem].mask,
			instructionSet[mnem].instrLen,
			instructionSet[mnem].instrFmt,
			instructionSet[mnem].instrType)
	}

	fmt.Fprintf(csvWriter, ";\n")
	csvWriter.Flush()
	csvFile.Close()
}
