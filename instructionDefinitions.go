// InstructionDefinitions.go

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

// Instruction Types
const (
	NOVA_MEMREF = iota
	NOVA_OP
	NOVA_IO
	NOVA_PC
	ECLIPSE_MEMREF
	ECLIPSE_OP
	ECLIPSE_PC
	ECLIPSE_STACK
	EAGLE_IO
	EAGLE_PC
	EAGLE_OP
	EAGLE_MEMREF
	EAGLE_STACK
)

// Instruction Formats
const (
	DERR_FMT = iota
	IMM_MODE_2_WORD_FMT
	IMM_ONEACC_FMT
	IO_FLAGS_DEV_FMT
	IO_TEST_DEV_FMT
	LNDO_4_WORD_FMT
	NOACC_MODE_3_WORD_FMT
	NOACC_MODE_IMM_IND_3_WORD_FMT
	NOACC_MODE_IND_2_WORD_E_FMT
	NOACC_MODE_IND_2_WORD_X_FMT
	NOACC_MODE_IND_3_WORD_FMT
	NOACC_MODE_IND_3_WORD_XCALL_FMT
	NOACC_MODE_IND_4_WORD_FMT
	NOVA_DATA_IO_FMT
	NOVA_NOACC_EFF_ADDR_FMT
	NOVA_ONEACC_EFF_ADDR_FMT
	NOVA_TWOACC_MULT_OP_FMT
	ONEACC_IMM_2_WORD_FMT
	ONEACC_IMMWD_2_WORD_FMT
	ONEACC_IMM_3_WORD_FMT
	ONEACC_IMMDWD_3_WORD_FMT
	ONEACC_MODE_2_WORD_E_FMT
	ONEACC_MODE_2_WORD_X_B_FMT
	ONEACC_MODE_3_WORD_FMT
	ONEACC_MODE_IND_2_WORD_E_FMT
	ONEACC_MODE_IND_2_WORD_X_FMT
	ONEACC_MODE_IND_3_WORD_FMT
	ONEACC_1_WORD_FMT
	UNIQUE_1_WORD_FMT
	UNIQUE_2_WORD_FMT
	SPLIT_8BIT_DISP_FMT
	THREE_WORD_DO_FMT
	TWOACC_1_WORD_FMT
	TWOACC_IMM_2_WORD_FMT
	WSKB_FMT
)

// InstructionsInit initialises the instruction characterstics for each instruction(
func instructionsInit() {
	instructionSet["ADC"] = instrChars{0x8400, 0x8700, 1, NOVA_TWOACC_MULT_OP_FMT, NOVA_OP}
	instructionSet["ADD"] = instrChars{0x8600, 0x8700, 1, NOVA_TWOACC_MULT_OP_FMT, NOVA_OP}
	instructionSet["ADDI"] = instrChars{0xe7f8, 0xe7ff, 2, ONEACC_IMM_2_WORD_FMT, EAGLE_OP}
	instructionSet["ADI"] = instrChars{0x8008, 0x87ff, 1, IMM_ONEACC_FMT, ECLIPSE_OP}
	instructionSet["ANC"] = instrChars{0x8188, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["AND"] = instrChars{0x8700, 0x8700, 1, NOVA_TWOACC_MULT_OP_FMT, NOVA_OP}
	instructionSet["ANDI"] = instrChars{0xc7f8, 0xe7ff, 2, ONEACC_IMMWD_2_WORD_FMT, EAGLE_OP}
	instructionSet["BAM"] = instrChars{0x97c8, 0xffff, 1, UNIQUE_1_WORD_FMT, NOVA_OP}
	instructionSet["BKPT"] = instrChars{0xc789, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_PC}
	instructionSet["BLM"] = instrChars{0xb7c8, 0xffff, 1, UNIQUE_1_WORD_FMT, ECLIPSE_MEMREF}
	instructionSet["BTO"] = instrChars{0x8408, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_OP}
	instructionSet["BTZ"] = instrChars{0x8448, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_OP}
	instructionSet["CIO"] = instrChars{0x85e9, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_IO}
	instructionSet["CIOI"] = instrChars{0x85f9, 0x87ff, 2, TWOACC_IMM_2_WORD_FMT, EAGLE_IO}
	instructionSet["CLM"] = instrChars{0x84f8, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_PC}
	instructionSet["CMP"] = instrChars{0xdfa8, 0xffff, 1, UNIQUE_1_WORD_FMT, ECLIPSE_MEMREF}
	instructionSet["CMT"] = instrChars{0xefa8, 0xffff, 1, UNIQUE_1_WORD_FMT, NOVA_OP}
	instructionSet["CMV"] = instrChars{0xd7a8, 0xffff, 1, UNIQUE_1_WORD_FMT, NOVA_OP}
	instructionSet["COB"] = instrChars{0x8588, 0x87ff, 1, TWOACC_1_WORD_FMT, NOVA_OP}
	instructionSet["COM"] = instrChars{0x8000, 0x8700, 1, NOVA_TWOACC_MULT_OP_FMT, NOVA_OP}
	instructionSet["CRYTC"] = instrChars{0xa7e9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["CRYTO"] = instrChars{0xa7c9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["CRYTZ"] = instrChars{0xa7d9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["CTR"] = instrChars{0xe7a8, 0xffff, 1, UNIQUE_1_WORD_FMT, NOVA_OP}
	instructionSet["CVWN"] = instrChars{0xe669, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["DAD"] = instrChars{0x8088, 0x87ff, 1, TWOACC_1_WORD_FMT, NOVA_OP}
	instructionSet["DEQUE"] = instrChars{0xe7c9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["DERR"] = instrChars{0x8f09, 0x8fcf, 1, DERR_FMT, EAGLE_OP}
	instructionSet["DHXL"] = instrChars{0x8388, 0x87ff, 1, IMM_ONEACC_FMT, NOVA_OP}
	instructionSet["DHXR"] = instrChars{0x83c8, 0x87ff, 1, IMM_ONEACC_FMT, NOVA_OP}
	instructionSet["DIA"] = instrChars{0x6100, 0xe700, 1, NOVA_DATA_IO_FMT, NOVA_IO}
	instructionSet["DIB"] = instrChars{0x6300, 0xe700, 1, NOVA_DATA_IO_FMT, NOVA_IO}
	instructionSet["DIC"] = instrChars{0x6500, 0xe700, 1, NOVA_DATA_IO_FMT, NOVA_IO}
	instructionSet["DIV"] = instrChars{0xd7c8, 0xffff, 1, UNIQUE_1_WORD_FMT, ECLIPSE_OP}
	instructionSet["DIVS"] = instrChars{0xdfc8, 0xffff, 1, UNIQUE_1_WORD_FMT, NOVA_OP}
	instructionSet["DIVX"] = instrChars{0xbfc8, 0xffff, 1, UNIQUE_1_WORD_FMT, NOVA_OP}
	instructionSet["DLSH"] = instrChars{0x82c8, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_OP}
	instructionSet["DOA"] = instrChars{0x6200, 0xe700, 1, NOVA_DATA_IO_FMT, NOVA_IO}
	instructionSet["DOB"] = instrChars{0x6400, 0xe700, 1, NOVA_DATA_IO_FMT, NOVA_IO}
	instructionSet["DOC"] = instrChars{0x6600, 0xe700, 1, NOVA_DATA_IO_FMT, NOVA_IO}
	instructionSet["DSB"] = instrChars{0x80c8, 0x87ff, 1, TWOACC_1_WORD_FMT, NOVA_OP}
	instructionSet["DSPA"] = instrChars{0xc478, 0xe4ff, 2, ONEACC_MODE_IND_2_WORD_E_FMT, ECLIPSE_PC}
	instructionSet["DSZ"] = instrChars{0x1800, 0xf800, 1, NOVA_NOACC_EFF_ADDR_FMT, NOVA_MEMREF}
	instructionSet["DSZTS"] = instrChars{0xc7d9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["ECLID"] = instrChars{0xffc8, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["EDIT"] = instrChars{0xf7a8, 0xffff, 1, UNIQUE_1_WORD_FMT, NOVA_OP}
	instructionSet["EDSZ"] = instrChars{0x9c38, 0xfcff, 2, NOACC_MODE_IND_2_WORD_E_FMT, ECLIPSE_PC}
	instructionSet["EISZ"] = instrChars{0x9438, 0xfcff, 2, NOACC_MODE_IND_2_WORD_E_FMT, ECLIPSE_PC}
	instructionSet["EJMP"] = instrChars{0x8438, 0xfcff, 2, NOACC_MODE_IND_2_WORD_E_FMT, ECLIPSE_PC}
	instructionSet["EJSR"] = instrChars{0x8c38, 0xfcff, 2, NOACC_MODE_IND_2_WORD_E_FMT, ECLIPSE_PC}
	instructionSet["ELDA"] = instrChars{0xa438, 0xe4ff, 2, ONEACC_MODE_IND_2_WORD_E_FMT, ECLIPSE_MEMREF}
	instructionSet["ELDB"] = instrChars{0x8478, 0xe4ff, 2, ONEACC_MODE_IND_2_WORD_E_FMT, ECLIPSE_MEMREF}
	instructionSet["ELEF"] = instrChars{0xe438, 0xe4ff, 2, ONEACC_MODE_IND_2_WORD_E_FMT, ECLIPSE_OP}
	instructionSet["ENQH"] = instrChars{0xc7e9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["ENQT"] = instrChars{0xc7f9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["ESTA"] = instrChars{0xc438, 0xe4ff, 2, ONEACC_MODE_IND_2_WORD_E_FMT, ECLIPSE_OP}
	instructionSet["ESTB"] = instrChars{0xa478, 0xe4ff, 2, ONEACC_MODE_2_WORD_E_FMT, ECLIPSE_OP}
	instructionSet["FNS"] = instrChars{0x86a8, 0xffff, 1, UNIQUE_1_WORD_FMT, NOVA_PC}
	instructionSet["FSA"] = instrChars{0x8ea8, 0xffff, 1, UNIQUE_1_WORD_FMT, NOVA_PC}
	instructionSet["FXTD"] = instrChars{0xa779, 0xffff, 1, UNIQUE_1_WORD_FMT, NOVA_OP}
	instructionSet["FXTE"] = instrChars{0xc749, 0xffff, 1, UNIQUE_1_WORD_FMT, NOVA_OP}
	instructionSet["HALT"] = instrChars{0x647f, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_IO}
	instructionSet["HLV"] = instrChars{0xc6f8, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["HXL"] = instrChars{0x8308, 0x87ff, 1, IMM_ONEACC_FMT, ECLIPSE_OP}
	instructionSet["HXR"] = instrChars{0x8348, 0x87ff, 1, IMM_ONEACC_FMT, ECLIPSE_OP}
	instructionSet["INC"] = instrChars{0x8300, 0x8700, 1, NOVA_TWOACC_MULT_OP_FMT, NOVA_OP}
	instructionSet["INTA"] = instrChars{0x633f, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_IO}
	instructionSet["INTDS"] = instrChars{0x60bf, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_IO}
	instructionSet["INTEN"] = instrChars{0x607f, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_IO}
	instructionSet["IOR"] = instrChars{0x8108, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_OP}
	instructionSet["IORI"] = instrChars{0x87f8, 0xe7ff, 2, ONEACC_IMMWD_2_WORD_FMT, ECLIPSE_OP}
	instructionSet["IORST"] = instrChars{0x653f, 0xe73f, 1, ONEACC_1_WORD_FMT, NOVA_IO}
	instructionSet["ISZ"] = instrChars{0x1000, 0xf800, 1, NOVA_NOACC_EFF_ADDR_FMT, NOVA_MEMREF}
	instructionSet["ISZTS"] = instrChars{0xc7c9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["JMP"] = instrChars{0x0, 0xf800, 1, NOVA_NOACC_EFF_ADDR_FMT, NOVA_PC}
	instructionSet["JSR"] = instrChars{0x800, 0xf800, 1, NOVA_NOACC_EFF_ADDR_FMT, NOVA_PC}
	instructionSet["LCALL"] = instrChars{0xa6c9, 0xe7ff, 4, NOACC_MODE_IND_4_WORD_FMT, EAGLE_PC}
	instructionSet["LCPID"] = instrChars{0x8759, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_IO}
	instructionSet["LDA"] = instrChars{0x2000, 0xe000, 1, NOVA_ONEACC_EFF_ADDR_FMT, NOVA_MEMREF}
	instructionSet["LDAFP"] = instrChars{0xc669, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_STACK}
	instructionSet["LDASB"] = instrChars{0xc649, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_STACK}
	instructionSet["LDASL"] = instrChars{0xa669, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_STACK}
	instructionSet["LDASP"] = instrChars{0xa649, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_STACK}
	instructionSet["LDATS"] = instrChars{0x8649, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["LDB"] = instrChars{0x85c8, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_OP}
	instructionSet["LDSP"] = instrChars{0x8519, 0x87ff, 3, ONEACC_MODE_IND_3_WORD_FMT, EAGLE_PC}
	instructionSet["LEF"] = instrChars{0x6000, 0xe000, 1, NOVA_ONEACC_EFF_ADDR_FMT, NOVA_MEMREF}
	instructionSet["LJMP"] = instrChars{0xa6d9, 0xe7ff, 3, NOACC_MODE_IND_3_WORD_FMT, EAGLE_PC}
	instructionSet["LJSR"] = instrChars{0xa6e9, 0xe7ff, 3, NOACC_MODE_IND_3_WORD_FMT, EAGLE_PC}
	instructionSet["LLDB"] = instrChars{0x84c9, 0x87ff, 3, ONEACC_MODE_3_WORD_FMT, EAGLE_OP}
	instructionSet["LLEF"] = instrChars{0x83e9, 0x87ff, 3, ONEACC_MODE_IND_3_WORD_FMT, EAGLE_OP}
	instructionSet["LLEFB"] = instrChars{0x84e9, 0x87ff, 3, ONEACC_MODE_3_WORD_FMT, EAGLE_OP}
	instructionSet["LMRF"] = instrChars{0x87c9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["LNADD"] = instrChars{0x8218, 0x87ff, 3, ONEACC_MODE_IND_3_WORD_FMT, EAGLE_OP}
	instructionSet["LNADI"] = instrChars{0x8618, 0x87ff, 3, NOACC_MODE_IMM_IND_3_WORD_FMT, EAGLE_OP}
	instructionSet["LNDIV"] = instrChars{0x82d8, 0x87ff, 3, ONEACC_MODE_IND_3_WORD_FMT, EAGLE_OP}
	instructionSet["LNDO"] = instrChars{0x8698, 0x87ff, 4, LNDO_4_WORD_FMT, EAGLE_PC}
	instructionSet["LNDSZ"] = instrChars{0x86d9, 0xe7ff, 3, NOACC_MODE_IND_3_WORD_FMT, EAGLE_PC}
	instructionSet["LNISZ"] = instrChars{0x86c9, 0xe7ff, 3, NOACC_MODE_IND_3_WORD_FMT, EAGLE_PC}
	instructionSet["LNLDA"] = instrChars{0x83c9, 0x87ff, 3, ONEACC_MODE_IND_3_WORD_FMT, EAGLE_MEMREF}
	instructionSet["LNMUL"] = instrChars{0x8298, 0x87ff, 3, ONEACC_MODE_IND_3_WORD_FMT, EAGLE_MEMREF}
	instructionSet["LNSBI"] = instrChars{0x8658, 0x87ff, 3, NOACC_MODE_IMM_IND_3_WORD_FMT, EAGLE_MEMREF}
	instructionSet["LNSTA"] = instrChars{0x83d9, 0x87ff, 3, ONEACC_MODE_IND_3_WORD_FMT, EAGLE_MEMREF}
	instructionSet["LNSUB"] = instrChars{0x8258, 0x87ff, 3, ONEACC_MODE_IND_3_WORD_FMT, EAGLE_MEMREF}
	instructionSet["LOB"] = instrChars{0x8508, 0x87ff, 1, TWOACC_1_WORD_FMT, NOVA_OP}
	instructionSet["LPEF"] = instrChars{0xa6f9, 0xe7ff, 3, NOACC_MODE_IND_3_WORD_FMT, EAGLE_STACK}
	instructionSet["LPEFB"] = instrChars{0xc6f9, 0xe7ff, 3, NOACC_MODE_3_WORD_FMT, EAGLE_STACK}
	instructionSet["LPHY"] = instrChars{0x87e9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["LPSHJ"] = instrChars{0xC6C9, 0xE7FF, 3, NOACC_MODE_IND_3_WORD_FMT, EAGLE_PC}
	instructionSet["LSH"] = instrChars{0x8288, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_OP}
	instructionSet["LWDO"] = instrChars{0x8798, 0x87ff, 4, LNDO_4_WORD_FMT, EAGLE_PC}
	instructionSet["LWDSZ"] = instrChars{0x86f9, 0xe7ff, 3, NOACC_MODE_IND_3_WORD_FMT, EAGLE_PC}
	instructionSet["LWISZ"] = instrChars{0x86e9, 0xe7ff, 3, NOACC_MODE_IND_3_WORD_FMT, EAGLE_PC}
	instructionSet["LWLDA"] = instrChars{0x83f9, 0x87ff, 3, ONEACC_MODE_IND_3_WORD_FMT, EAGLE_MEMREF}
	instructionSet["LWSTA"] = instrChars{0x84f9, 0x87ff, 3, ONEACC_MODE_IND_3_WORD_FMT, EAGLE_MEMREF}
	instructionSet["MOV"] = instrChars{0x8200, 0x8700, 1, NOVA_TWOACC_MULT_OP_FMT, NOVA_OP}
	instructionSet["MSKO"] = instrChars{0x643f, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_IO}
	instructionSet["MUL"] = instrChars{0xc7c8, 0xffff, 1, UNIQUE_1_WORD_FMT, ECLIPSE_OP}
	instructionSet["NADD"] = instrChars{0x8049, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["NADDI"] = instrChars{0xc639, 0xe7ff, 2, ONEACC_IMM_2_WORD_FMT, EAGLE_OP}
	instructionSet["NADI"] = instrChars{0x8599, 0x87ff, 1, IMM_ONEACC_FMT, EAGLE_OP}
	instructionSet["NCLID"] = instrChars{0x683f, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_IO}
	instructionSet["NEG"] = instrChars{0x8100, 0x8700, 1, NOVA_TWOACC_MULT_OP_FMT, NOVA_OP}
	instructionSet["NIO"] = instrChars{0x6000, 0xff00, 1, IO_FLAGS_DEV_FMT, NOVA_IO}
	instructionSet["NLDAI"] = instrChars{0xc629, 0xe7ff, 2, ONEACC_IMM_2_WORD_FMT, EAGLE_OP}
	instructionSet["NMUL"] = instrChars{0x8069, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["NSBI"] = instrChars{0x85a9, 0x87ff, 1, IMM_ONEACC_FMT, EAGLE_OP}
	instructionSet["NSUB"] = instrChars{0x8059, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["PIO"] = instrChars{0x85d9, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_IO}
	instructionSet["POP"] = instrChars{0x8688, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_STACK}
	instructionSet["POPJ"] = instrChars{0x9fc8, 0xffff, 1, UNIQUE_1_WORD_FMT, ECLIPSE_STACK}
	instructionSet["PRTRST"] = instrChars{0x85d9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_IO}
	instructionSet["PRTSEL"] = instrChars{0x783f, 0xffff, 1, UNIQUE_1_WORD_FMT, NOVA_IO}
	instructionSet["PSH"] = instrChars{0x8648, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_STACK}
	instructionSet["PSHJ"] = instrChars{0x84b8, 0xfcff, 2, NOACC_MODE_IND_2_WORD_E_FMT, ECLIPSE_STACK}
	instructionSet["READS"] = instrChars{0x613f, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_IO}
	instructionSet["RTN"] = instrChars{0xafc8, 0xffff, 1, UNIQUE_1_WORD_FMT, ECLIPSE_STACK}
	instructionSet["SAVE"] = instrChars{0xe7c8, 0xffff, 2, UNIQUE_2_WORD_FMT, ECLIPSE_STACK}
	instructionSet["SBI"] = instrChars{0x8048, 0x87ff, 1, IMM_ONEACC_FMT, ECLIPSE_OP}
	instructionSet["SEX"] = instrChars{0x8349, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["SGT"] = instrChars{0x8208, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_PC}
	instructionSet["SKP"] = instrChars{0x6700, 0xff00, 1, IO_TEST_DEV_FMT, NOVA_IO}
	instructionSet["SNB"] = instrChars{0x85f8, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_PC}
	instructionSet["SPSR"] = instrChars{0xa7a9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["SPTE"] = instrChars{0xe729, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["SSPT"] = instrChars{0xe7d9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_OP}
	instructionSet["STA"] = instrChars{0x4000, 0xe000, 1, NOVA_ONEACC_EFF_ADDR_FMT, NOVA_MEMREF}
	instructionSet["STAFP"] = instrChars{0xc679, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_STACK}
	instructionSet["STASB"] = instrChars{0xc659, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_STACK}
	instructionSet["STASL"] = instrChars{0xa679, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_STACK}
	instructionSet["STASP"] = instrChars{0xa659, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_STACK}
	instructionSet["STATS"] = instrChars{0x8659, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_STACK}
	instructionSet["STB"] = instrChars{0x8608, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_OP}
	instructionSet["SUB"] = instrChars{0x8500, 0x8700, 1, NOVA_TWOACC_MULT_OP_FMT, NOVA_OP}
	instructionSet["SZB"] = instrChars{0x8488, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_PC}
	instructionSet["SZBO"] = instrChars{0x84c8, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_OP}
	instructionSet["WADC"] = instrChars{0x8249, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WADD"] = instrChars{0x8149, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WADDI"] = instrChars{0x8689, 0xe7ff, 3, ONEACC_IMM_3_WORD_FMT, EAGLE_OP}
	instructionSet["WADI"] = instrChars{0x84b9, 0x87ff, 1, IMM_ONEACC_FMT, EAGLE_OP}
	instructionSet["WANC"] = instrChars{0x8549, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WAND"] = instrChars{0x8449, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WANDI"] = instrChars{0x8699, 0xe7ff, 3, ONEACC_IMMDWD_3_WORD_FMT, EAGLE_OP}
	instructionSet["WASH"] = instrChars{0x8279, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WASHI"] = instrChars{0xc6a9, 0xe7ff, 2, ONEACC_IMM_2_WORD_FMT, EAGLE_OP}
	instructionSet["WBLM"] = instrChars{0xe749, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_MEMREF}
	instructionSet["WBR"] = instrChars{0x8038, 0x843f, 1, SPLIT_8BIT_DISP_FMT, EAGLE_PC}
	instructionSet["WBTO"] = instrChars{0x8299, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WBTZ"] = instrChars{0x82a9, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WCLM"] = instrChars{0x8569, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WCMV"] = instrChars{0x8779, 0xFFFF, 1, UNIQUE_1_WORD_FMT, EAGLE_MEMREF}
	instructionSet["WCOM"] = instrChars{0x8459, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WINC"] = instrChars{0x8259, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WIOR"] = instrChars{0x8469, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WIORI"] = instrChars{0x86a9, 0xe7ff, 3, ONEACC_IMMDWD_3_WORD_FMT, EAGLE_OP}
	instructionSet["WLDAI"] = instrChars{0xc689, 0xe7ff, 3, ONEACC_IMMDWD_3_WORD_FMT, EAGLE_OP}
	instructionSet["WLMP"] = instrChars{0xa7f9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_IO}
	instructionSet["WLSH"] = instrChars{0x8559, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WLSHI"] = instrChars{0xe6d9, 0xe7ff, 2, ONEACC_IMM_2_WORD_FMT, EAGLE_OP}
	instructionSet["WLSI"] = instrChars{0x85b9, 0x87ff, 1, IMM_ONEACC_FMT, EAGLE_OP}
	instructionSet["WMOV"] = instrChars{0x8379, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WMOVR"] = instrChars{0xe699, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WMSP"] = instrChars{0xe649, 0xe7ff, 1, ONEACC_1_WORD_FMT, EAGLE_STACK}
	instructionSet["WNADI"] = instrChars{0xE6F9, 0xE7FF, 2, ONEACC_IMM_2_WORD_FMT, EAGLE_OP}
	instructionSet["WNEG"] = instrChars{0x8269, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["WPOP"] = instrChars{0x8089, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_STACK}
	instructionSet["WPOPJ"] = instrChars{0x8789, 0xFFFF, 1, UNIQUE_1_WORD_FMT, EAGLE_PC}
	instructionSet["WPSH"] = instrChars{0x8579, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_STACK}
	instructionSet["WRTN"] = instrChars{0x87a9, 0xffff, 1, UNIQUE_1_WORD_FMT, EAGLE_PC}
	instructionSet["WSAVR"] = instrChars{0xA729, 0xFFFF, 2, UNIQUE_2_WORD_FMT, EAGLE_STACK}
	instructionSet["WSAVS"] = instrChars{0xA739, 0xFFFF, 2, UNIQUE_2_WORD_FMT, EAGLE_STACK}
	instructionSet["WSBI"] = instrChars{0x8589, 0x87ff, 1, IMM_ONEACC_FMT, EAGLE_OP}
	instructionSet["WSEQ"] = instrChars{0x80b9, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_PC}
	instructionSet["WSEQI"] = instrChars{0xe6c9, 0xe7ff, 2, ONEACC_IMM_2_WORD_FMT, EAGLE_PC}
	instructionSet["WSGE"] = instrChars{0x8199, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_PC}
	instructionSet["WSGT"] = instrChars{0x81b9, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_PC}
	instructionSet["WSGTI"] = instrChars{0xe689, 0xe7ff, 2, ONEACC_IMM_2_WORD_FMT, EAGLE_PC}
	instructionSet["WSKBO"] = instrChars{0x8f49, 0x8fcf, 1, WSKB_FMT, EAGLE_PC}
	instructionSet["WSKBZ"] = instrChars{0x8f89, 0x8fcf, 1, WSKB_FMT, EAGLE_PC}
	instructionSet["WSLE"] = instrChars{0x81a9, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_PC}
	instructionSet["WSLEI"] = instrChars{0xe6a9, 0xe7ff, 2, ONEACC_IMM_2_WORD_FMT, EAGLE_PC}
	instructionSet["WSLT"] = instrChars{0x8289, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_PC}
	instructionSet["WSNE"] = instrChars{0x8189, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_PC}
	instructionSet["WSNEI"] = instrChars{0xe6e9, 0xe7ff, 2, ONEACC_IMM_2_WORD_FMT, EAGLE_PC}
	instructionSet["WSTB"] = instrChars{0x8539, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_MEMREF}
	instructionSet["WSUB"] = instrChars{0x8159, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
	instructionSet["XCALL"] = instrChars{0x8609, 0xe7ff, 3, NOACC_MODE_IND_3_WORD_XCALL_FMT, EAGLE_PC}
	instructionSet["XCH"] = instrChars{0x81c8, 0x87ff, 1, TWOACC_1_WORD_FMT, ECLIPSE_OP}
	instructionSet["XJMP"] = instrChars{0xc609, 0xe7ff, 2, NOACC_MODE_IND_2_WORD_X_FMT, EAGLE_PC}
	instructionSet["XJSR"] = instrChars{0xc619, 0xe7ff, 2, NOACC_MODE_IND_2_WORD_X_FMT, EAGLE_PC}
	instructionSet["XLDB"] = instrChars{0x8419, 0x87ff, 2, ONEACC_MODE_2_WORD_X_B_FMT, EAGLE_MEMREF}
	instructionSet["XLEF"] = instrChars{0x8409, 0x87ff, 2, ONEACC_MODE_IND_2_WORD_X_FMT, EAGLE_MEMREF}
	instructionSet["XLEFB"] = instrChars{0x8439, 0x87ff, 2, ONEACC_MODE_2_WORD_X_B_FMT, EAGLE_MEMREF}
	instructionSet["XNADD"] = instrChars{0x8018, 0x87ff, 2, ONEACC_MODE_IND_2_WORD_X_FMT, EAGLE_MEMREF}
	instructionSet["XNADI"] = instrChars{0x8418, 0x87ff, 2, IMM_MODE_2_WORD_FMT, EAGLE_MEMREF}
	instructionSet["XNDO"] = instrChars{0x8498, 0x87ff, 3, THREE_WORD_DO_FMT, EAGLE_MEMREF}
	instructionSet["XNDSZ"] = instrChars{0xa609, 0xe7ff, 2, NOACC_MODE_IND_2_WORD_X_FMT, EAGLE_PC}
	instructionSet["XNISZ"] = instrChars{0x8639, 0xe7ff, 2, NOACC_MODE_IND_2_WORD_X_FMT, EAGLE_PC}
	instructionSet["XNLDA"] = instrChars{0x8329, 0x87ff, 2, ONEACC_MODE_IND_2_WORD_X_FMT, EAGLE_MEMREF}
	instructionSet["XNSBI"] = instrChars{0x8458, 0x87ff, 2, IMM_MODE_2_WORD_FMT, EAGLE_MEMREF}
	instructionSet["XNSTA"] = instrChars{0x8339, 0x87ff, 2, ONEACC_MODE_IND_2_WORD_X_FMT, EAGLE_MEMREF}
	instructionSet["XNSUB"] = instrChars{0x8058, 0x87ff, 2, IMM_MODE_2_WORD_FMT, EAGLE_MEMREF}
	instructionSet["XPEF"] = instrChars{0x8629, 0xe7ff, 2, NOACC_MODE_IND_2_WORD_X_FMT, EAGLE_STACK}
	instructionSet["XSTB"] = instrChars{0x8429, 0x87ff, 2, ONEACC_MODE_2_WORD_X_B_FMT, EAGLE_MEMREF}
	instructionSet["XWADD"] = instrChars{0x8118, 0x87ff, 2, ONEACC_MODE_IND_2_WORD_X_FMT, EAGLE_MEMREF}
	instructionSet["XWADI"] = instrChars{0x8518, 0x87ff, 2, IMM_MODE_2_WORD_FMT, EAGLE_MEMREF}
	instructionSet["XWDSZ"] = instrChars{0xA639, 0xE7FF, 2, NOACC_MODE_IND_2_WORD_X_FMT, EAGLE_PC}
	instructionSet["XWLDA"] = instrChars{0x8309, 0x87ff, 2, ONEACC_MODE_IND_2_WORD_X_FMT, EAGLE_MEMREF}
	instructionSet["XWSBI"] = instrChars{0x8558, 0x87ff, 2, IMM_MODE_2_WORD_FMT, EAGLE_OP}
	instructionSet["XWSTA"] = instrChars{0x8319, 0x87ff, 2, ONEACC_MODE_IND_2_WORD_X_FMT, EAGLE_MEMREF}
	instructionSet["XWSUB"] = instrChars{0x8158, 0x87ff, 2, ONEACC_MODE_IND_2_WORD_X_FMT, EAGLE_OP}
	instructionSet["ZEX"] = instrChars{0x8359, 0x87ff, 1, TWOACC_1_WORD_FMT, EAGLE_OP}
}
