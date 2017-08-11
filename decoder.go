// decoder.go

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
	"mvemg/logging"
)

// decodedInstrT defines the MV/Em internal decode of an opcode and any
// parameters.
type decodedInstrT struct {
	mnemonic          string
	instrFmt          int
	instrType         int
	instrLength       int
	c, ind, sh, nl, f byte
	t                 string
	ioDev             int
	acs, acd          int
	skip              string
	mode              string
	disp8             int8
	disp15            int16
	disp16            int16
	disp31            int32
	offsetU16         uint16
	bitLow            bool
	immU16            uint16
	immS16            int16
	immU32            uint32
	immS32            int32
	immWord           DgWordT
	immDword          DgDwordT
	argCount          int
	bitNum            int
	disassembly       string
}

const numPosOpcodes = 65536

var opCodeLookup [numPosOpcodes]string

func decoderGenAllPossOpcodes() {
	for opcode := 0; opcode < numPosOpcodes; opcode++ {
		mnem, found := instructionFind2(DgWordT(opcode), false, false, false)
		if found {
			opCodeLookup[opcode] = mnem
		} else {
			opCodeLookup[opcode] = ""
		}
	}
}

// InsrtructionFind looks for the opcode in the instruction map and returns
// the corresponding mnemonic
func instructionFind(opcode DgWordT, lefMode bool, ioOn bool, atuOn bool) (string, bool) {
	if opCodeLookup[opcode] != "" {
		if opCodeLookup[opcode] == "LEF" && lefMode {
			return "", false
		}
		return opCodeLookup[opcode], true
	}
	return "", false
}

// InsrtructionFind2 looks for the opcode in the instruction map and returns
// the corresponding mnemonic
func instructionFind2(opcode DgWordT, lefMode bool, ioOn bool, atuOn bool) (string, bool) {
	var tail DgWordT
	for mnem, insChar := range instructionSet {
		if (opcode & insChar.mask) == insChar.bits {
			// there are some exceptions to the normal decoding...
			switch mnem {
			case "LEF":
				if lefMode {
					return "", false
				}
			case "ADC", "ADD", "AND", "COM", "INC", "MOV", "NEG", "SUB":
				tail = opcode & 0x000f
				if tail < 8 || tail > 9 {
					return mnem, true
				}
			default:
				return mnem, true

			}
		}
	}
	return "", false
}

// InstructionDecode decodes an opcode
//
// N.B. For the moment this function both decodes and disassembles the given instruction,
// for performance in the future these two tasks should probably either be separated or
// controlled by flags passed into the function.
func instructionDecode(opcode DgWordT, pc DgPhysAddrT, lefMode bool, ioOn bool, autOn bool) (*decodedInstrT, bool) {
	var decodedInstrT decodedInstrT
	var secondWord, thirdWord, fourthWord DgWordT

	decodedInstrT.disassembly = "; Unknown instruction"

	mnem, found := instructionFind(opcode, lefMode, ioOn, autOn)
	if !found {
		logging.DebugPrint(logging.DebugLog, "INFO: instructionFind failed to return anything to instructionDecode for location %d.\n", pc)
		return &decodedInstrT, false
	}
	decodedInstrT.mnemonic = mnem
	decodedInstrT.disassembly = mnem
	decodedInstrT.instrFmt = instructionSet[mnem].instrFmt
	decodedInstrT.instrType = instructionSet[mnem].instrType
	decodedInstrT.instrLength = instructionSet[mnem].instrLen

	switch decodedInstrT.instrFmt {

	case IMM_MODE_2_WORD_FMT: // eg. XNADI, XWADI
		decodedInstrT.immU16 = decode2bitImm(getWbits(opcode, 1, 2))
		decodedInstrT.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstrT.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstrT.disp15 = decode15bitDisp(secondWord, decodedInstrT.mode)
		decodedInstrT.disassembly += fmt.Sprintf(" %d.,%d.%s [2-Word OpCode]",
			decodedInstrT.immU16, decodedInstrT.disp15, modeToString(decodedInstrT.mode))

	case IMM_ONEACC_FMT: // eg. ADI, HXL, NADI, SBI, WADI, WLSI, WSBI
		// N.B. Immediate value is encoded by assembler to be one less than required
		//      This is handled by decode2bitImm()
		decodedInstrT.immU16 = decode2bitImm(getWbits(opcode, 1, 2))
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		decodedInstrT.disassembly += fmt.Sprintf(" %d.,%d", decodedInstrT.immU16, decodedInstrT.acd)

	case IO_FLAGS_DEV_FMT:
		decodedInstrT.f = decodeIOFlags(getWbits(opcode, 8, 2))
		decodedInstrT.ioDev = int(getWbits(opcode, 10, 6))
		decodedInstrT.disassembly += fmt.Sprintf("%c %s",
			decodedInstrT.f, deviceToString(decodedInstrT.ioDev))

	case IO_RESET_FMT:
		decodedInstrT.acd = int(getWbits(opcode, 3, 2)) // TODO is this needed/used?

	case IO_TEST_DEV_FMT:
		decodedInstrT.t = decodeIOTest(getWbits(opcode, 8, 2))
		decodedInstrT.ioDev = int(getWbits(opcode, 10, 6))
		decodedInstrT.disassembly += fmt.Sprintf("%s %s", decodedInstrT.t, deviceToString(decodedInstrT.ioDev))

	case LNDO_4_WORD_FMT:
		decodedInstrT.acd = int(getWbits(opcode, 1, 2))
		decodedInstrT.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		thirdWord = memReadWord(pc + 2)
		fourthWord = memReadWord(pc + 3)
		decodedInstrT.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstrT.disp31 = decode31bitDisp(secondWord, thirdWord, decodedInstrT.mode)
		decodedInstrT.offsetU16 = uint16(fourthWord)
		decodedInstrT.disassembly += fmt.Sprintf(" %d,%d.,%c%d%s [4-Word OpCode]",
			decodedInstrT.acd, decodedInstrT.offsetU16, decodedInstrT.ind, decodedInstrT.disp31, modeToString(decodedInstrT.mode))

	case NOACC_MODE_3_WORD_FMT: // eg. LPEFB,
		decodedInstrT.mode = decodeMode(getWbits(opcode, 3, 2))
		decodedInstrT.immU32 = uint32(memReadDWord(pc + 1))
		decodedInstrT.disassembly += fmt.Sprintf(" %d.,%s [3-Word OpCode]", decodedInstrT.immU32, modeToString(decodedInstrT.mode))

	case NOACC_MODE_IND_2_WORD_E_FMT, NOACC_MODE_IND_2_WORD_X_FMT:
		logging.DebugPrint(logging.DebugLog, "X_FMT: Mnemonic is <%s>\n", decodedInstrT.mnemonic)
		switch decodedInstrT.mnemonic {
		case "XJMP", "XJSR", "XNDSZ", "XNISZ", "XPEF", "XWDSZ":
			decodedInstrT.mode = decodeMode(getWbits(opcode, 3, 2))
		case "EDSZ", "EISZ", "EJMP", "EJSR", "PSHJ":
			decodedInstrT.mode = decodeMode(getWbits(opcode, 6, 2))
		}
		secondWord = memReadWord(pc + 1)
		decodedInstrT.ind = decodeIndirect(testWbit(secondWord, 0))
		switch decodedInstrT.mnemonic {
		case "EJSR", "EJMP": // FIXME - maybe more exceptions needed here
			decodedInstrT.disp15 = decode15bitEclipseDisp(secondWord, decodedInstrT.mode)
		default:
			decodedInstrT.disp15 = decode15bitDisp(secondWord, decodedInstrT.mode)
		}
		decodedInstrT.disassembly += fmt.Sprintf(" %c%d.%s [2-Word OpCode]",
			decodedInstrT.ind, decodedInstrT.disp15, modeToString(decodedInstrT.mode))

	case NOACC_MODE_IND_3_WORD_FMT: // eg. LJMP/LJSR, LNISZ, LNDSZ, LWDS
		decodedInstrT.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		thirdWord = memReadWord(pc + 2)
		decodedInstrT.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstrT.disp31 = decode31bitDisp(secondWord, thirdWord, decodedInstrT.mode)
		decodedInstrT.disassembly += fmt.Sprintf(" %c%d.%s [3-Word OpCode]",
			decodedInstrT.ind, decodedInstrT.disp31, modeToString(decodedInstrT.mode))

	case NOACC_MODE_IND_3_WORD_XCALL_FMT: // XCALL
		decodedInstrT.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		thirdWord = memReadWord(pc + 2)
		decodedInstrT.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstrT.disp15 = decode15bitDisp(secondWord, decodedInstrT.mode)
		decodedInstrT.argCount = int(thirdWord)
		decodedInstrT.disassembly += fmt.Sprintf(" %c%d.%s, %d [3-Word OpCode]",
			decodedInstrT.ind, decodedInstrT.disp15, modeToString(decodedInstrT.mode), decodedInstrT.argCount)

	case NOACC_MODE_IMM_IND_3_WORD_FMT: // eg. LNADI, LNSBI
		decodedInstrT.immU16 = decode2bitImm(getWbits(opcode, 1, 2))
		decodedInstrT.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		thirdWord = memReadWord(pc + 2)
		decodedInstrT.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstrT.disp31 = decode31bitDisp(secondWord, thirdWord, decodedInstrT.mode)
		decodedInstrT.disassembly += fmt.Sprintf(" %d.,%c%d.%s [3-Word OpCode]",
			decodedInstrT.immU16, decodedInstrT.ind, decodedInstrT.disp31, modeToString(decodedInstrT.mode))

	case NOACC_MODE_IND_4_WORD_FMT: // eg. LCALL
		decodedInstrT.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		thirdWord = memReadWord(pc + 2)
		fourthWord = memReadWord(pc + 3)
		decodedInstrT.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstrT.disp31 = decode31bitDisp(secondWord, thirdWord, decodedInstrT.mode)
		decodedInstrT.argCount = int(fourthWord)
		decodedInstrT.disassembly += fmt.Sprintf(" %c%d.%s,%d. [4-Word OpCode]",
			decodedInstrT.ind, decodedInstrT.disp31, modeToString(decodedInstrT.mode), decodedInstrT.argCount)

	case NOVA_DATA_IO_FMT:
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		decodedInstrT.f = decodeIOFlags(getWbits(opcode, 8, 2))
		decodedInstrT.ioDev = int(getWbits(opcode, 10, 6))
		decodedInstrT.disassembly += fmt.Sprintf("%c %d,%s",
			decodedInstrT.f, decodedInstrT.acd, deviceToString(decodedInstrT.ioDev))

	case NOVA_NOACC_EFF_ADDR_FMT:
		decodedInstrT.ind = decodeIndirect(testWbit(opcode, 5))
		decodedInstrT.mode = decodeMode(getWbits(opcode, 6, 2))
		decodedInstrT.disp15 = decode8bitDisp(DgByteT(opcode&0x00ff), decodedInstrT.mode) // NB
		decodedInstrT.disassembly += fmt.Sprintf(" %c%d.%s",
			decodedInstrT.ind, decodedInstrT.disp15, modeToString(decodedInstrT.mode))

	case NOVA_ONEACC_EFF_ADDR_FMT:
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		decodedInstrT.ind = decodeIndirect(testWbit(opcode, 5))
		decodedInstrT.mode = decodeMode(getWbits(opcode, 6, 2))
		decodedInstrT.disp15 = decode8bitDisp(DgByteT(opcode&0x00ff), decodedInstrT.mode) // NB
		decodedInstrT.disassembly += fmt.Sprintf(" %d,%c%d.%s",
			decodedInstrT.acd, decodedInstrT.ind, decodedInstrT.disp15, modeToString(decodedInstrT.mode))

	case NOVA_TWOACC_MULT_OP_FMT:
		decodedInstrT.acs = int(getWbits(opcode, 1, 2))
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		decodedInstrT.sh = decodeShift(getWbits(opcode, 8, 2))
		decodedInstrT.c = decodeCarry(getWbits(opcode, 10, 2))
		decodedInstrT.nl = decodeNoLoad(testWbit(opcode, 12))
		decodedInstrT.skip = decodeSkip(getWbits(opcode, 13, 3))
		decodedInstrT.disassembly += fmt.Sprintf("%c%c%c %d,%d %s",
			decodedInstrT.c, decodedInstrT.sh, decodedInstrT.nl, decodedInstrT.acs, decodedInstrT.acd, skipToString(decodedInstrT.skip))

	case ONEACC_1_WORD_FMT:
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		decodedInstrT.disassembly += fmt.Sprintf(" %d", decodedInstrT.acd)

	case ONEACC_IMM_2_WORD_FMT: // eg. ADDI, NADDI, NLDAI, , WSEQI, WLSHI, WNADI
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstrT.immS16 = int16(secondWord)
		decodedInstrT.disassembly += fmt.Sprintf(" %d,%d. [2-Word OpCode]", decodedInstrT.immS16, decodedInstrT.acd)

	case ONEACC_IMMWD_2_WORD_FMT: // eg. ANDI, IORI
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstrT.immWord = secondWord
		decodedInstrT.disassembly += fmt.Sprintf(" %d.,%d [2-Word OpCode]", decodedInstrT.immWord, decodedInstrT.acd)

	case ONEACC_IMM_3_WORD_FMT: // eg. WADDI
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		decodedInstrT.immS32 = int32(memReadDWord(pc + 1))
		decodedInstrT.disassembly += fmt.Sprintf(" %d.,%d [3-Word OpCode]", decodedInstrT.immS32, decodedInstrT.acd)

	case ONEACC_IMMDWD_3_WORD_FMT: // eg. WANDI, WIORI, WLDAI
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		decodedInstrT.immDword = memReadDWord(pc + 1)
		decodedInstrT.disassembly += fmt.Sprintf(" %d.,%d [3-Word OpCode]", decodedInstrT.immDword, decodedInstrT.acd)

	case ONEACC_MODE_2_WORD_X_B_FMT: // eg. XLDB, XLEFB, XSTB
		decodedInstrT.mode = decodeMode(getWbits(opcode, 1, 2))
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstrT.disp16, decodedInstrT.bitLow = decode16bitByteDisp(secondWord)
		decodedInstrT.disassembly += fmt.Sprintf(" %d,%d.+%c%s [2-Word OpCode]",
			decodedInstrT.acd, decodedInstrT.disp16*2, loHiToByte(decodedInstrT.bitLow), modeToString(decodedInstrT.mode))

	case ONEACC_MODE_2_WORD_E_FMT: // eg. ELDB, ESTB
		decodedInstrT.mode = decodeMode(getWbits(opcode, 6, 2))
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstrT.disp16, decodedInstrT.bitLow = decode16bitByteDisp(secondWord)
		decodedInstrT.disassembly += fmt.Sprintf(" %d,%d.+%c%s [2-Word OpCode]",
			decodedInstrT.acd, decodedInstrT.disp16*2, loHiToByte(decodedInstrT.bitLow), modeToString(decodedInstrT.mode))

	case ONEACC_MODE_3_WORD_FMT: // eg. LLDB, LLEFB
		decodedInstrT.mode = decodeMode(getWbits(opcode, 1, 2))
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		thirdWord = memReadWord(pc + 2)
		decodedInstrT.disp31 = decode31bitDisp(secondWord, thirdWord, decodedInstrT.mode)
		decodedInstrT.disassembly += fmt.Sprintf(" %d,%d.%s [3-Word OpCode]",
			decodedInstrT.acd, decodedInstrT.disp31, modeToString(decodedInstrT.mode))

	case ONEACC_MODE_IND_2_WORD_E_FMT: // eg. ELDA
		decodedInstrT.mode = decodeMode(getWbits(opcode, 6, 2))
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstrT.ind = decodeIndirect(testWbit(secondWord, 0))
		switch decodedInstrT.mnemonic {
		case "ELEF": // FIXME - more exceptions needed here...
			decodedInstrT.disp15 = decode15bitEclipseDisp(secondWord, decodedInstrT.mode)
		default:
			decodedInstrT.disp15 = decode15bitDisp(secondWord, decodedInstrT.mode)
		}
		decodedInstrT.disassembly += fmt.Sprintf(" %d,%c%d.%s [2-Word OpCode]",
			decodedInstrT.acd, decodedInstrT.ind, decodedInstrT.disp15, modeToString(decodedInstrT.mode))

	case ONEACC_MODE_IND_2_WORD_X_FMT: // eg. XNLDA/XWSTA, XLEF
		decodedInstrT.mode = decodeMode(getWbits(opcode, 1, 2))
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstrT.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstrT.disp15 = decode15bitDisp(secondWord, decodedInstrT.mode)
		decodedInstrT.disassembly += fmt.Sprintf(" %d,%c%d.%s [2-Word OpCode]",
			decodedInstrT.acd, decodedInstrT.ind, decodedInstrT.disp15, modeToString(decodedInstrT.mode))

	case ONEACC_MODE_IND_3_WORD_FMT: // eg. LWLDA/LWSTA,LNLDA
		decodedInstrT.mode = decodeMode(getWbits(opcode, 1, 2))
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstrT.ind = decodeIndirect(testWbit(secondWord, 0))
		thirdWord = memReadWord(pc + 2)
		decodedInstrT.disp31 = decode31bitDisp(secondWord, thirdWord, decodedInstrT.mode)
		decodedInstrT.disassembly += fmt.Sprintf(" %d,%c%d.%s [3-Word OpCode]",
			decodedInstrT.acd, decodedInstrT.ind, decodedInstrT.disp31, modeToString(decodedInstrT.mode))

	case TWOACC_1_WORD_FMT: // eg. WSUB
		decodedInstrT.acs = int(getWbits(opcode, 1, 2))
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		decodedInstrT.disassembly += fmt.Sprintf(" %d,%d", decodedInstrT.acs, decodedInstrT.acd)

	case SPLIT_8BIT_DISP_FMT: // eg. WBR, always a signed disp
		tmp8bit := DgByteT(getWbits(opcode, 1, 4) & 0xff)
		tmp8bit = tmp8bit << 4
		tmp8bit |= DgByteT(getWbits(opcode, 6, 4) & 0xff)
		decodedInstrT.disp8 = int8(decode8bitDisp(tmp8bit, "PC"))
		decodedInstrT.disassembly += fmt.Sprintf(" %d.", int32(decodedInstrT.disp8))

	case THREE_WORD_DO_FMT: // eg. XNDO
		decodedInstrT.acd = int(getWbits(opcode, 1, 2))
		decodedInstrT.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstrT.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstrT.disp15 = decode15bitDisp(secondWord, decodedInstrT.mode)
		thirdWord = memReadWord(pc + 2)
		decodedInstrT.offsetU16 = uint16(thirdWord)
		decodedInstrT.disassembly += fmt.Sprintf(" %d,%d. %c%d.%s [3-Word OpCode]",
			decodedInstrT.acd, decodedInstrT.offsetU16, decodedInstrT.ind, decodedInstrT.disp15, modeToString(decodedInstrT.mode))

	case TWOACC_IMM_2_WORD_FMT: // eg. CIOI
		decodedInstrT.acs = int(getWbits(opcode, 1, 2))
		decodedInstrT.acd = int(getWbits(opcode, 3, 2))
		decodedInstrT.immWord = memReadWord(pc + 1)
		decodedInstrT.disassembly += fmt.Sprintf(" %d.,%d,%d", decodedInstrT.immWord, decodedInstrT.acs,
			decodedInstrT.acd)

	case UNIQUE_1_WORD_FMT:
		// nothing to do in this case

	case UNIQUE_2_WORD_FMT: // eg.SAVE, WSAVR, WSAVS
		decodedInstrT.immU16 = uint16(memReadWord(pc + 1))
		decodedInstrT.disassembly += fmt.Sprintf(" %d. [2-Word OpCode]", decodedInstrT.immU16)

	case WSKB_FMT:
		tmp8bit := DgByteT(getWbits(opcode, 1, 3) & 0xff)
		tmp8bit = tmp8bit << 2
		tmp8bit |= DgByteT(getWbits(opcode, 10, 2) & 0xff)
		decodedInstrT.bitNum = int(uint8(tmp8bit))
		decodedInstrT.disassembly += fmt.Sprintf(" %d.", decodedInstrT.bitNum)

	default:
		logging.DebugPrint(logging.DebugLog, "ERROR: Invalid instruction format (%d) for instruction %s",
			decodedInstrT.instrFmt, decodedInstrT.mnemonic)
		return nil, false
	}

	return &decodedInstrT, true
}

/* decoders for (parts of) operands below here... */

var disp16 int16

func decode2bitImm(i DgWordT) uint16 {
	// to expand range (by 1!) 1 is subtracted from operand
	return uint16(i + 1)
}

// Decode8BitDisp must return signed 16-bit as the result could be
// either 8-bit signed or 8-bit unsigned
func decode8bitDisp(d8 DgByteT, mode string) int16 {
	if mode == "Absolute" {
		disp16 = int16(d8) & 0x00ff // unsigned offset
	} else {
		// signed offset...
		disp16 = int16(int8(d8)) // this should sign-extend
	}
	return disp16
}

func decode15bitDisp(d15 DgWordT, mode string) int16 {
	if mode == "Absolute" {
		disp16 = int16(d15 & 0x7fff) // zero extend
	} else {
		if testWbit(d15, 1) {
			disp16 = int16(d15 | 0x8000) // sign extend
		} else {
			disp16 = int16(d15 & 0x7fff) // zero extend
		}
		if mode == "PC" {
			disp16++ // see p.1-12 of PoP
		}
	}
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... decode15bitDisp got: %d, returning: %d\n", d15, disp16)
	}
	return disp16
}

func decode15bitEclipseDisp(d15 DgWordT, mode string) int16 {
	if mode == "Absolute" {
		disp16 = int16(d15 & 0x7fff) // zero extend
	} else {
		if testWbit(d15, 1) {
			disp16 = int16(d15 | 0xc000) // sign extend
		} else {
			disp16 = int16(d15 & 0x3fff) // zero extend
		}
		if mode == "PC" {
			disp16++ // see p.1-12 of PoP
		}
	}
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... decode15bitEclispeDisp got: %d, returning: %d\n", d15, disp16)
	}
	return disp16
}

func decode16bitByteDisp(d16 DgWordT) (int16, bool) {
	loHi := testWbit(d16, 15)
	disp16 = int16(d16 >> 1)
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... decode16bitByteDisp got: %d, returning %d\n", d16, disp16)
	}
	return disp16, loHi
}

func decode31bitDisp(d1, d2 DgWordT, mode string) int32 {
	// FIXME Test this!
	var disp32 int32
	if testWbit(d1, 1) {
		disp32 = int32(int16(d1 | 0x8000)) // sign extend
	} else {
		disp32 = int32(int16(d1)) & 0x00007fff // zero extend
	}
	disp32 = (disp32 << 16) | (int32(d2) & 0x0000ffff)
	if mode == "PC" {
		disp32++ // see p.1-12 of PoP
	}
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... decode31bitDisp got: %d %d, returning: %d\n", d1, d2, disp32)
	}
	return disp32
}

func decodeCarry(cry DgWordT) byte {
	switch cry {
	case 0:
		return ' '
	case 1:
		return 'Z'
	case 2:
		return 'O' // Letter 'O' for One
	case 3:
		return 'C'
	}
	return '*'
}

func decodeIndirect(i bool) byte {
	if i {
		return '@'
	}
	return ' '
}

func decodeIOFlags(fl DgWordT) byte {
	return ioFlags[fl]
}

func decodeIOTest(t DgWordT) string {
	return ioTests[t]
}

func decodeMode(ix DgWordT) string {
	return modes[ix]
}

func decodeNoLoad(n bool) byte {
	if n {
		return '#'
	}
	return ' '
}

func decodeShift(sh DgWordT) byte {
	switch sh {
	case 0:
		return ' '
	case 1:
		return 'L'
	case 2:
		return 'R'
	case 3:
		return 'S'
	}
	return '*'
}

func decodeSkip(skp DgWordT) string {
	return skips[skp]
}

func loHiToByte(loHi bool) byte {
	if loHi {
		return 'H'
	}
	return 'L'
}

func modeToString(mode string) string {
	if mode == "Absolute" {
		return ""
	}
	return "," + mode
}

func skipToString(s string) string {
	if s == "NONE" {
		return ""
	}
	return s
}
