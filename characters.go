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

// ASCII and DASHER special characters
const (
	asciiBEL = 0x07
	asciiBS  = 0x08
	asciiTAB = 0x09
	asciiNL  = 0x0A
	asciiFF  = 0x0C
	asciiCR  = 0x0D
	asciiESC = 0x1B
	asciiSPC = 0x20

	dasherERASEEOL        = 013
	dasherERASEPAGE       = 014
	dasherCURSORLEFT      = 0x19
	dasherWRITEWINDOWADDR = 020 //followed by col then row
	dasherDIMON           = 034
	dasherDIMOFF          = 035
	dasherUNDERLINE       = 024
	dasherNORMAL          = 025
	dasherDELETE          = 0177
)
