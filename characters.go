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
	ASCIIBEL  = 0x07
	ASCII_BS  = 0x08
	ASCII_TAB = 0x09
	ASCII_NL  = 0x0A
	ASCII_FF  = 0x0C
	ASCII_CR  = 0x0D
	ASCII_ESC = 0x1B
	ASCII_SPC = 0x20

	DASHER_ERASE_EOL         = 013
	DASHER_ERASE_PAGE        = 014
	DASHER_CURSOR_LEFT       = 0x19
	DASHER_WRITE_WINDOW_ADDR = 020 //followed by col then row
	DASHER_DIM_ON            = 034
	DASHER_DIM_OFF           = 035
	DASHER_UNDERLINE         = 024
	DASHER_NORMAL            = 025
	DASHER_DELETE            = 0177
)
