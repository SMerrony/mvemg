// utils.go

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

package util

import (
	"fmt"
	"mvemg/dg"
)

// BoolToInt converts a bool to 1 or 0
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// BoolToYN converts a bool to Y or N
func BoolToYN(b bool) byte {
	if b {
		return 'Y'
	}
	return 'N'
}

// BoolToOnOff converts a bool to "On" or "Off"
func BoolToOnOff(b bool) string {
	if b {
		return "On"
	}
	return "Off"
}

// BoolToOZ converts a boolean to a O(ne) or Z(ero) byte
func boolToOZ(b bool) byte {
	if b {
		return 'O'
	}
	return 'Z'
}

// DWordGetLowerWord gets the DG-lower word of a doubleword
// Called VERY often, hopefully inlined!
func DWordGetLowerWord(dwd dg.DwordT) dg.WordT {
	return dg.WordT(dwd) // & 0x0000ffff mask unneccessary
}

// DWordGetUpperWord gets the DG-higher word of a doubleword
// Called VERY often, hopefully inlined!
func DWordGetUpperWord(dwd dg.DwordT) dg.WordT {
	return dg.WordT(dwd >> 16)
}

// DWordFromTwoWords - catenate two DG Words into a DG DoubleWord
func DWordFromTwoWords(hw dg.WordT, lw dg.WordT) dg.DwordT {
	return dg.DwordT(hw)<<16 | dg.DwordT(lw)
}

// GetWbits - in the DG world, the first (leftmost) bit is numbered zero...
// extract nbits from value starting at leftBit
func GetWbits(value dg.WordT, leftBit int, nbits int) dg.WordT {
	var res dg.WordT
	rightBit := leftBit + nbits
	for b := leftBit; b < rightBit; b++ {
		res = res << 1
		if TestWbit(value, b) {
			res++
		}
	}
	return res
}

// SetWbit sets a single bit in a DG word
func SetWbit(word dg.WordT, bitNum uint) dg.WordT {
	return word | 1<<(15-bitNum)
}

// ClearWbit clears a single bit in a DG word
func ClearWbit(word dg.WordT, bitNum uint) dg.WordT {
	return word ^ 1<<(15-bitNum)
}

// GetDWbits - in the DG world, the first (leftmost) bit is numbered zero...
// extract nbits from value starting at leftBit
func GetDWbits(value dg.DwordT, leftBit int, nbits int) dg.DwordT {
	var res dg.DwordT
	rightBit := leftBit + nbits
	for b := leftBit; b < rightBit; b++ {
		res = res << 1
		if TestDWbit(value, b) {
			res++
		}
	}
	return res
}

var bb uint8

// TestWbit - does word w have bit b set?
func TestWbit(w dg.WordT, b int) bool {
	bb = uint8(b)
	return (w & (1 << (15 - bb))) != 0
}

// TestDWbit - does dword dw have bit b set?
func TestDWbit(dw dg.DwordT, b int) bool {
	bb = uint8(b)
	return ((dw & (1 << (31 - bb))) != 0)
}

// WordToBinStr - get a pretty-printable string of a word
func WordToBinStr(w dg.WordT) string {
	return fmt.Sprintf("%08b %08b", w>>8, w&0x0ff)
}

// SexWordToDWord - sign-extend a DG word to a DG DoubleWord
func SexWordToDWord(wd dg.WordT) dg.DwordT {
	var dwd dg.DwordT
	if TestWbit(wd, 0) {
		dwd = dg.DwordT(wd) | 0xffff0000
	} else {
		dwd = dg.DwordT(wd) & 0x0000ffff
	}
	return dwd
}

// SwapBytes - swap over the two bytes in a dg_word
func SwapBytes(wd dg.WordT) dg.WordT {
	var res dg.WordT
	res = (wd >> 8) | ((wd & 0x00ff) << 8)
	return res
}