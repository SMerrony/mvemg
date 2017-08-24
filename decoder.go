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
	"mvemg/dg"
	"mvemg/logging"
	"mvemg/memory"
	"mvemg/util"
)

// decodedInstrT defines the MV/Em internal decode of an opcode and any
// parameters.
type decodedInstrT struct {
	mnemonic          string
	instrFmt          int
	instrType         int
	instrLength       int
	disassembly       string
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
	immWord           dg.WordT
	immDword          dg.DwordT
	argCount          int
	bitNum            int
	variant           interface{}
}

// here are the types for the variant portion of the decoded instruction...
type immMode2WordT struct {
	immU16 uint16
	mode   string
	ind    byte
	disp15 int16
}
type immOneAccT struct {
	immU16 uint16
	acd    int
}
type ioFlagsDevT struct {
	f     byte
	ioDev int
}
type ioTestDevT struct {
	t     string
	ioDev int
}
type lndo4WordT struct {
	acd       int
	mode      string
	ind       byte
	disp31    int32
	offsetU16 uint16
}
type noAccMode3WordT struct {
	mode   string
	immU32 uint32
}
type noAccModeInd2WordT struct {
	mode   string
	ind    byte
	disp15 int16
}
type noAccModeInd3WordT struct {
	mode   string
	ind    byte
	disp31 int32
}
type noAccModeInd3WordXcallT struct {
	mode     string
	ind      byte
	disp15   int16
	argCount int
}
type noAccModeImmInd3WordT struct {
	immU16 uint16
	mode   string
	ind    byte
	disp31 int32
}
type noAccModeInd4WordT struct {
	mode     string
	ind      byte
	disp31   int32
	argCount int
}
type novaDataIoT struct {
	acd   int
	f     byte
	ioDev int
}
type novaNoAccEffAddrT struct {
	mode   string
	ind    byte
	disp15 int16
}

const numPosOpcodes = 65536

var opCodeLookup [numPosOpcodes]string

func decoderGenAllPossOpcodes() {
	for opcode := 0; opcode < numPosOpcodes; opcode++ {
		mnem, found := instructionMatch(dg.WordT(opcode), false, false, false)
		if found {
			opCodeLookup[opcode] = mnem
		} else {
			opCodeLookup[opcode] = ""
		}
	}
}

// InstructionFind looks up an opcode in the opcode lookup table and returns
// the corresponding mnemonic
func instructionLookup(opcode dg.WordT, lefMode bool, ioOn bool, atuOn bool) (string, bool) {
	if opCodeLookup[opcode] != "" {
		if opCodeLookup[opcode] == "LEF" && lefMode {
			return "", false
		}
		return opCodeLookup[opcode], true
	}
	return "", false
}

// instructionMatch looks for a match for the opcode in the instruction set and returns
// the corresponding mnemonic
func instructionMatch(opcode dg.WordT, lefMode bool, ioOn bool, atuOn bool) (string, bool) {
	var tail dg.WordT
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
func instructionDecode(opcode dg.WordT, pc dg.PhysAddrT, lefMode bool, ioOn bool, autOn bool) (*decodedInstrT, bool) {
	var decodedInstr decodedInstrT
	var secondWord, thirdWord, fourthWord dg.WordT

	decodedInstr.disassembly = "; Unknown instruction"

	mnem, found := instructionLookup(opcode, lefMode, ioOn, autOn)
	if !found {
		logging.DebugPrint(logging.DebugLog, "INFO: instructionFind failed to return anything to instructionDecode for location %d.\n", pc)
		return &decodedInstr, false
	}
	decodedInstr.mnemonic = mnem
	decodedInstr.disassembly = mnem
	decodedInstr.instrFmt = instructionSet[mnem].instrFmt
	decodedInstr.instrType = instructionSet[mnem].instrType
	decodedInstr.instrLength = instructionSet[mnem].instrLen

	switch decodedInstr.instrFmt {

	case IMM_MODE_2_WORD_FMT: // eg. XNADI, XNSBI, XNSUB, XWADI, XWSBI
		var immMode2Word immMode2WordT
		immMode2Word.immU16 = decode2bitImm(util.GetWbits(opcode, 1, 2))
		immMode2Word.mode = decodeMode(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		immMode2Word.ind = decodeIndirect(util.TestWbit(secondWord, 0))
		immMode2Word.disp15 = decode15bitDisp(secondWord, immMode2Word.mode)
		decodedInstr.variant = immMode2Word
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%d.%s [2-Word OpCode]",
			immMode2Word.immU16, immMode2Word.disp15, modeToString(immMode2Word.mode))

	case IMM_ONEACC_FMT: // eg. ADI, HXL, NADI, SBI, WADI, WLSI, WSBI
		// N.B. Immediate value is encoded by assembler to be one less than required
		//      This is handled by decode2bitImm()
		var immOneAcc immOneAccT
		immOneAcc.immU16 = decode2bitImm(util.GetWbits(opcode, 1, 2))
		immOneAcc.acd = int(util.GetWbits(opcode, 3, 2))
		decodedInstr.variant = immOneAcc
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%d", immOneAcc.immU16, immOneAcc.acd)

	case IO_FLAGS_DEV_FMT:
		var ioFlagsDev ioFlagsDevT
		ioFlagsDev.f = decodeIOFlags(util.GetWbits(opcode, 8, 2))
		ioFlagsDev.ioDev = int(util.GetWbits(opcode, 10, 6))
		decodedInstr.variant = ioFlagsDev
		decodedInstr.disassembly += fmt.Sprintf("%c %s",
			ioFlagsDev.f, deviceToString(ioFlagsDev.ioDev))

	case IO_TEST_DEV_FMT:
		var ioTestDev ioTestDevT
		ioTestDev.t = decodeIOTest(util.GetWbits(opcode, 8, 2))
		ioTestDev.ioDev = int(util.GetWbits(opcode, 10, 6))
		decodedInstr.variant = ioTestDev
		decodedInstr.disassembly += fmt.Sprintf("%s %s", ioTestDev.t, deviceToString(ioTestDev.ioDev))

	case LNDO_4_WORD_FMT:
		var lndo4Word lndo4WordT
		lndo4Word.acd = int(util.GetWbits(opcode, 1, 2))
		lndo4Word.mode = decodeMode(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		thirdWord = memory.ReadWord(pc + 2)
		fourthWord = memory.ReadWord(pc + 3)
		lndo4Word.ind = decodeIndirect(util.TestWbit(secondWord, 0))
		lndo4Word.disp31 = decode31bitDisp(secondWord, thirdWord, lndo4Word.mode)
		lndo4Word.offsetU16 = uint16(fourthWord)
		decodedInstr.variant = lndo4Word
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d.,%c%d%s [4-Word OpCode]",
			lndo4Word.acd, lndo4Word.offsetU16, lndo4Word.ind, lndo4Word.disp31,
			modeToString(lndo4Word.mode))

	case NOACC_MODE_3_WORD_FMT: // eg. LPEFB,
		var noAccMode3Word noAccMode3WordT
		noAccMode3Word.mode = decodeMode(util.GetWbits(opcode, 3, 2))
		noAccMode3Word.immU32 = uint32(memory.ReadDWord(pc + 1))
		decodedInstr.variant = noAccMode3Word
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%s [3-Word OpCode]",
			noAccMode3Word.immU32, modeToString(noAccMode3Word.mode))

	case NOACC_MODE_IND_2_WORD_E_FMT, NOACC_MODE_IND_2_WORD_X_FMT:
		var noAccModeInd2Word noAccModeInd2WordT
		logging.DebugPrint(logging.DebugLog, "X_FMT: Mnemonic is <%s>\n", decodedInstr.mnemonic)
		switch decodedInstr.mnemonic {
		case "XJMP", "XJSR", "XNDSZ", "XNISZ", "XPEF", "XWDSZ":
			noAccModeInd2Word.mode = decodeMode(util.GetWbits(opcode, 3, 2))
		case "EDSZ", "EISZ", "EJMP", "EJSR", "PSHJ":
			noAccModeInd2Word.mode = decodeMode(util.GetWbits(opcode, 6, 2))
		}
		secondWord = memory.ReadWord(pc + 1)
		noAccModeInd2Word.ind = decodeIndirect(util.TestWbit(secondWord, 0))
		noAccModeInd2Word.disp15 = decode15bitDisp(secondWord, noAccModeInd2Word.mode)
		decodedInstr.variant = noAccModeInd2Word
		decodedInstr.disassembly += fmt.Sprintf(" %c%d.%s [2-Word OpCode]",
			noAccModeInd2Word.ind, noAccModeInd2Word.disp15, modeToString(noAccModeInd2Word.mode))

	case NOACC_MODE_IND_3_WORD_FMT: // eg. LJMP/LJSR, LNISZ, LNDSZ, LWDS
		var noAccModeInd3Word noAccModeInd3WordT
		noAccModeInd3Word.mode = decodeMode(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		thirdWord = memory.ReadWord(pc + 2)
		noAccModeInd3Word.ind = decodeIndirect(util.TestWbit(secondWord, 0))
		noAccModeInd3Word.disp31 = decode31bitDisp(secondWord, thirdWord, noAccModeInd3Word.mode)
		decodedInstr.variant = noAccModeInd3Word
		decodedInstr.disassembly += fmt.Sprintf(" %c%d.%s [3-Word OpCode]",
			noAccModeInd3Word.ind, noAccModeInd3Word.disp31, modeToString(noAccModeInd3Word.mode))

	case NOACC_MODE_IND_3_WORD_XCALL_FMT: // XCALL
		var noAccModeInd3WordXcall noAccModeInd3WordXcallT
		noAccModeInd3WordXcall.mode = decodeMode(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		thirdWord = memory.ReadWord(pc + 2)
		noAccModeInd3WordXcall.ind = decodeIndirect(util.TestWbit(secondWord, 0))
		noAccModeInd3WordXcall.disp15 = decode15bitDisp(secondWord, noAccModeInd3WordXcall.mode)
		noAccModeInd3WordXcall.argCount = int(thirdWord)
		decodedInstr.variant = noAccModeInd3WordXcall
		decodedInstr.disassembly += fmt.Sprintf(" %c%d.%s, %d [3-Word OpCode]",
			noAccModeInd3WordXcall.ind, noAccModeInd3WordXcall.disp15,
			modeToString(noAccModeInd3WordXcall.mode), noAccModeInd3WordXcall.argCount)

	case NOACC_MODE_IMM_IND_3_WORD_FMT: // eg. LNADI, LNSBI
		var noAccModeImmInd3Word noAccModeImmInd3WordT
		noAccModeImmInd3Word.immU16 = decode2bitImm(util.GetWbits(opcode, 1, 2))
		noAccModeImmInd3Word.mode = decodeMode(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		thirdWord = memory.ReadWord(pc + 2)
		noAccModeImmInd3Word.ind = decodeIndirect(util.TestWbit(secondWord, 0))
		noAccModeImmInd3Word.disp31 = decode31bitDisp(secondWord, thirdWord, decodedInstr.mode)
		decodedInstr.variant = noAccModeImmInd3Word
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%c%d.%s [3-Word OpCode]",
			noAccModeImmInd3Word.immU16, noAccModeImmInd3Word.ind, noAccModeImmInd3Word.disp31,
			modeToString(noAccModeImmInd3Word.mode))

	case NOACC_MODE_IND_4_WORD_FMT: // eg. LCALL
		var noAccModeInd4Word noAccModeInd4WordT
		noAccModeInd4Word.mode = decodeMode(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		thirdWord = memory.ReadWord(pc + 2)
		fourthWord = memory.ReadWord(pc + 3)
		noAccModeInd4Word.ind = decodeIndirect(util.TestWbit(secondWord, 0))
		noAccModeInd4Word.disp31 = decode31bitDisp(secondWord, thirdWord, noAccModeInd4Word.mode)
		noAccModeInd4Word.argCount = int(fourthWord)
		decodedInstr.variant = noAccModeInd4Word
		decodedInstr.disassembly += fmt.Sprintf(" %c%d.%s,%d. [4-Word OpCode]",
			noAccModeInd4Word.ind, noAccModeInd4Word.disp31, modeToString(noAccModeInd4Word.mode),
			noAccModeInd4Word.argCount)

	case NOVA_DATA_IO_FMT: // eg. DOA/B/C, DIA/B/C
		var novaDataIo novaDataIoT
		novaDataIo.acd = int(util.GetWbits(opcode, 3, 2))
		novaDataIo.f = decodeIOFlags(util.GetWbits(opcode, 8, 2))
		novaDataIo.ioDev = int(util.GetWbits(opcode, 10, 6))
		decodedInstr.variant = novaDataIo
		decodedInstr.disassembly += fmt.Sprintf("%c %d,%s",
			novaDataIo.f, novaDataIo.acd, deviceToString(novaDataIo.ioDev))

	case NOVA_NOACC_EFF_ADDR_FMT: // eg. JMP, JSR
		var novaNoAccEffAddr novaNoAccEffAddrT
		novaNoAccEffAddr.ind = decodeIndirect(util.TestWbit(opcode, 5))
		novaNoAccEffAddr.mode = decodeMode(util.GetWbits(opcode, 6, 2))
		novaNoAccEffAddr.disp15 = decode8bitDisp(dg.ByteT(opcode&0x00ff), novaNoAccEffAddr.mode) // NB
		decodedInstr.variant = novaNoAccEffAddr
		decodedInstr.disassembly += fmt.Sprintf(" %c%d.%s",
			novaNoAccEffAddr.ind, novaNoAccEffAddr.disp15, modeToString(novaNoAccEffAddr.mode))

	case NOVA_ONEACC_EFF_ADDR_FMT:
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		decodedInstr.ind = decodeIndirect(util.TestWbit(opcode, 5))
		decodedInstr.mode = decodeMode(util.GetWbits(opcode, 6, 2))
		decodedInstr.disp15 = decode8bitDisp(dg.ByteT(opcode&0x00ff), decodedInstr.mode) // NB
		decodedInstr.disassembly += fmt.Sprintf(" %d,%c%d.%s",
			decodedInstr.acd, decodedInstr.ind, decodedInstr.disp15, modeToString(decodedInstr.mode))

	case NOVA_TWOACC_MULT_OP_FMT:
		decodedInstr.acs = int(util.GetWbits(opcode, 1, 2))
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		decodedInstr.sh = decodeShift(util.GetWbits(opcode, 8, 2))
		decodedInstr.c = decodeCarry(util.GetWbits(opcode, 10, 2))
		decodedInstr.nl = decodeNoLoad(util.TestWbit(opcode, 12))
		decodedInstr.skip = decodeSkip(util.GetWbits(opcode, 13, 3))
		decodedInstr.disassembly += fmt.Sprintf("%c%c%c %d,%d %s",
			decodedInstr.c, decodedInstr.sh, decodedInstr.nl, decodedInstr.acs, decodedInstr.acd, skipToString(decodedInstr.skip))

	case ONEACC_1_WORD_FMT:
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		decodedInstr.disassembly += fmt.Sprintf(" %d", decodedInstr.acd)

	case ONEACC_IMM_2_WORD_FMT: // eg. ADDI, NADDI, NLDAI, , WSEQI, WLSHI, WNADI
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		decodedInstr.immS16 = int16(secondWord)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d. [2-Word OpCode]", decodedInstr.immS16, decodedInstr.acd)

	case ONEACC_IMMWD_2_WORD_FMT: // eg. ANDI, IORI
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		decodedInstr.immWord = secondWord
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%d [2-Word OpCode]", decodedInstr.immWord, decodedInstr.acd)

	case ONEACC_IMM_3_WORD_FMT: // eg. WADDI
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		decodedInstr.immS32 = int32(memory.ReadDWord(pc + 1))
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%d [3-Word OpCode]", decodedInstr.immS32, decodedInstr.acd)

	case ONEACC_IMMDWD_3_WORD_FMT: // eg. WANDI, WIORI, WLDAI
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		decodedInstr.immDword = memory.ReadDWord(pc + 1)
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%d [3-Word OpCode]", decodedInstr.immDword, decodedInstr.acd)

	case ONEACC_MODE_2_WORD_X_B_FMT: // eg. XLDB, XLEFB, XSTB
		decodedInstr.mode = decodeMode(util.GetWbits(opcode, 1, 2))
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		decodedInstr.disp16, decodedInstr.bitLow = decode16bitByteDisp(secondWord)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d.+%c%s [2-Word OpCode]",
			decodedInstr.acd, decodedInstr.disp16*2, loHiToByte(decodedInstr.bitLow), modeToString(decodedInstr.mode))

	case ONEACC_MODE_2_WORD_E_FMT: // eg. ELDB, ESTB
		decodedInstr.mode = decodeMode(util.GetWbits(opcode, 6, 2))
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		decodedInstr.disp16, decodedInstr.bitLow = decode16bitByteDisp(secondWord)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d.+%c%s [2-Word OpCode]",
			decodedInstr.acd, decodedInstr.disp16*2, loHiToByte(decodedInstr.bitLow), modeToString(decodedInstr.mode))

	case ONEACC_MODE_3_WORD_FMT: // eg. LLDB, LLEFB
		decodedInstr.mode = decodeMode(util.GetWbits(opcode, 1, 2))
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		thirdWord = memory.ReadWord(pc + 2)
		decodedInstr.disp31 = decode31bitDisp(secondWord, thirdWord, decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d.%s [3-Word OpCode]",
			decodedInstr.acd, decodedInstr.disp31, modeToString(decodedInstr.mode))

	case ONEACC_MODE_IND_2_WORD_E_FMT: // eg. ELDA
		decodedInstr.mode = decodeMode(util.GetWbits(opcode, 6, 2))
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		decodedInstr.ind = decodeIndirect(util.TestWbit(secondWord, 0))
		decodedInstr.disp15 = decode15bitDisp(secondWord, decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%c%d.%s [2-Word OpCode]",
			decodedInstr.acd, decodedInstr.ind, decodedInstr.disp15, modeToString(decodedInstr.mode))

	case ONEACC_MODE_IND_2_WORD_X_FMT: // eg. XNLDA/XWSTA, XLEF
		decodedInstr.mode = decodeMode(util.GetWbits(opcode, 1, 2))
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		decodedInstr.ind = decodeIndirect(util.TestWbit(secondWord, 0))
		decodedInstr.disp15 = decode15bitDisp(secondWord, decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%c%d.%s [2-Word OpCode]",
			decodedInstr.acd, decodedInstr.ind, decodedInstr.disp15, modeToString(decodedInstr.mode))

	case ONEACC_MODE_IND_3_WORD_FMT: // eg. LWLDA/LWSTA,LNLDA
		decodedInstr.mode = decodeMode(util.GetWbits(opcode, 1, 2))
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		decodedInstr.ind = decodeIndirect(util.TestWbit(secondWord, 0))
		thirdWord = memory.ReadWord(pc + 2)
		decodedInstr.disp31 = decode31bitDisp(secondWord, thirdWord, decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%c%d.%s [3-Word OpCode]",
			decodedInstr.acd, decodedInstr.ind, decodedInstr.disp31, modeToString(decodedInstr.mode))

	case TWOACC_1_WORD_FMT: // eg. WSUB
		decodedInstr.acs = int(util.GetWbits(opcode, 1, 2))
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d", decodedInstr.acs, decodedInstr.acd)

	case SPLIT_8BIT_DISP_FMT: // eg. WBR, always a signed disp
		tmp8bit := dg.ByteT(util.GetWbits(opcode, 1, 4) & 0xff)
		tmp8bit = tmp8bit << 4
		tmp8bit |= dg.ByteT(util.GetWbits(opcode, 6, 4) & 0xff)
		decodedInstr.disp8 = int8(decode8bitDisp(tmp8bit, "PC"))
		decodedInstr.disassembly += fmt.Sprintf(" %d.", int32(decodedInstr.disp8))

	case THREE_WORD_DO_FMT: // eg. XNDO
		decodedInstr.acd = int(util.GetWbits(opcode, 1, 2))
		decodedInstr.mode = decodeMode(util.GetWbits(opcode, 3, 2))
		secondWord = memory.ReadWord(pc + 1)
		decodedInstr.ind = decodeIndirect(util.TestWbit(secondWord, 0))
		decodedInstr.disp15 = decode15bitDisp(secondWord, decodedInstr.mode)
		thirdWord = memory.ReadWord(pc + 2)
		decodedInstr.offsetU16 = uint16(thirdWord)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d. %c%d.%s [3-Word OpCode]",
			decodedInstr.acd, decodedInstr.offsetU16, decodedInstr.ind, decodedInstr.disp15, modeToString(decodedInstr.mode))

	case TWOACC_IMM_2_WORD_FMT: // eg. CIOI
		decodedInstr.acs = int(util.GetWbits(opcode, 1, 2))
		decodedInstr.acd = int(util.GetWbits(opcode, 3, 2))
		decodedInstr.immWord = memory.ReadWord(pc + 1)
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%d,%d", decodedInstr.immWord, decodedInstr.acs,
			decodedInstr.acd)

	case UNIQUE_1_WORD_FMT:
		// nothing to do in this case

	case UNIQUE_2_WORD_FMT: // eg.SAVE, WSAVR, WSAVS
		decodedInstr.immU16 = uint16(memory.ReadWord(pc + 1))
		decodedInstr.disassembly += fmt.Sprintf(" %d. [2-Word OpCode]", decodedInstr.immU16)

	case WSKB_FMT:
		tmp8bit := dg.ByteT(util.GetWbits(opcode, 1, 3) & 0xff)
		tmp8bit = tmp8bit << 2
		tmp8bit |= dg.ByteT(util.GetWbits(opcode, 10, 2) & 0xff)
		decodedInstr.bitNum = int(uint8(tmp8bit))
		decodedInstr.disassembly += fmt.Sprintf(" %d.", decodedInstr.bitNum)

	default:
		logging.DebugPrint(logging.DebugLog, "ERROR: Invalid instruction format (%d) for instruction %s",
			decodedInstr.instrFmt, decodedInstr.mnemonic)
		return nil, false
	}

	return &decodedInstr, true
}

/* decoders for (parts of) operands below here... */

var disp16 int16

func decode2bitImm(i dg.WordT) uint16 {
	// to expand range (by 1!) 1 is subtracted from operand
	return uint16(i + 1)
}

// Decode8BitDisp must return signed 16-bit as the result could be
// either 8-bit signed or 8-bit unsigned
func decode8bitDisp(d8 dg.ByteT, mode string) int16 {
	if mode == "Absolute" {
		disp16 = int16(d8) & 0x00ff // unsigned offset
	} else {
		// signed offset...
		disp16 = int16(int8(d8)) // this should sign-extend
	}
	return disp16
}

func decode15bitDisp(d15 dg.WordT, mode string) int16 {
	if mode == "Absolute" {
		disp16 = int16(d15 & 0x7fff) // zero extend
	} else {
		if util.TestWbit(d15, 1) {
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

// func decode15bitEclipseDisp(d15 dg.WordT, mode string) int16 {
// 	if mode == "Absolute" {
// 		disp16 = int16(d15 & 0x7fff) // zero extend
// 	} else {
// 		if util.TestWbit(d15, 1) {
// 			disp16 = int16(d15 | 0xc000) // sign extend
// 		} else {
// 			disp16 = int16(d15 & 0x3fff) // zero extend
// 		}
// 		if mode == "PC" {
// 			disp16++ // see p.1-12 of PoP
// 		}
// 	}
// 	if debugLogging {
// 		logging.DebugPrint(logging.DebugLog, "... decode15bitEclispeDisp got: %d, returning: %d\n", d15, disp16)
// 	}
// 	return disp16
// }

func decode16bitByteDisp(d16 dg.WordT) (int16, bool) {
	loHi := util.TestWbit(d16, 15)
	disp16 = int16(d16 >> 1)
	if debugLogging {
		logging.DebugPrint(logging.DebugLog, "... decode16bitByteDisp got: %d, returning %d\n", d16, disp16)
	}
	return disp16, loHi
}

func decode31bitDisp(d1, d2 dg.WordT, mode string) int32 {
	// FIXME Test this!
	var disp32 int32
	if util.TestWbit(d1, 1) {
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

func decodeCarry(cry dg.WordT) byte {
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

func decodeIOFlags(fl dg.WordT) byte {
	return ioFlags[fl]
}

func decodeIOTest(t dg.WordT) string {
	return ioTests[t]
}

func decodeMode(ix dg.WordT) string {
	return modes[ix]
}

func decodeNoLoad(n bool) byte {
	if n {
		return '#'
	}
	return ' '
}

func decodeShift(sh dg.WordT) byte {
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

func decodeSkip(skp dg.WordT) string {
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
