// instructions.go
package main

import (
	"fmt"
	"log"
)

// N.B. If using the instruction map exported from the C emulator, then
//      the order of these const definitions must also be in sync with
//      the corresponding enums in the C version.

// instruction types
const (
	NOVA_MEMREF    = iota
	NOVA_OP        = iota
	NOVA_IO        = iota
	NOVA_PC        = iota
	ECLIPSE_MEMREF = iota
	ECLIPSE_OP     = iota
	ECLIPSE_PC     = iota
	ECLIPSE_STACK  = iota
	EAGLE_IO       = iota
	EAGLE_PC       = iota
	EAGLE_OP       = iota
	EAGLE_MEMREF   = iota
	EAGLE_STACK    = iota
)

// instruction formats
const (
	UNDEFINED_FMT                 = iota
	DERR_FMT                      = iota
	IMM_MODE_2_WORD_FMT           = iota
	IMM_ONEACC_FMT                = iota
	IO_FLAGS_DEV_FMT              = iota
	IO_RESET_FMT                  = iota
	IO_TEST_DEV_FMT               = iota
	LNDO_4_WORD_FMT               = iota
	NOACC_MODE_3_WORD_FMT         = iota
	NOACC_MODE_IMM_IND_3_WORD_FMT = iota
	NOACC_MODE_IND_2_WORD_E_FMT   = iota
	NOACC_MODE_IND_2_WORD_X_FMT   = iota
	NOACC_MODE_IND_3_WORD_FMT     = iota
	NOACC_MODE_IND_4_WORD_FMT     = iota
	NOVA_DATA_IO_FMT              = iota
	NOVA_NOACC_EFF_ADDR_FMT       = iota
	NOVA_ONEACC_EFF_ADDR_FMT      = iota
	NOVA_TWOACC_MULT_OP_FMT       = iota
	ONEACC_2_WORD_FMT             = iota
	ONEACC_IMM_2_WORD_FMT         = iota
	ONEACC_IMM_3_WORD_FMT         = iota
	ONEACC_IMM_IND_3_WORD_FMT     = iota
	ONEACC_MODE_2_WORD_E_FMT      = iota
	ONEACC_MODE_2_WORD_X_B_FMT    = iota
	ONEACC_MODE_3_WORD_FMT        = iota
	ONEACC_MODE_IND_2_WORD_E_FMT  = iota
	ONEACC_MODE_IND_2_WORD_X_FMT  = iota
	ONEACC_MODE_IND_3_WORD_FMT    = iota
	ONEACC_1_WORD_FMT             = iota
	UNIQUE_1_WORD_FMT             = iota
	UNIQUE_2_WORD_FMT             = iota
	SPLIT_8BIT_DISP_FMT           = iota
	THREE_WORD_DO_FMT             = iota
	TWOACC_1_WORD_FMT             = iota
	WSKB_FMT                      = iota
)

// the characteristics of each instruction
type instrChars struct {
	// mnemonic   string  // DG standard assembler mnemonic for opcode
	bits      dg_word // bit-pattern for opcode
	mask      dg_word // mask for unique bits in opcode
	instrFmt  int     // opcode layout
	instrLen  int     // # of words in opcode and any following args
	instrType int     // class of opcode (somewhat arbitrary)
	//xeqCounter uint64  // count of # times instruction hit during this run
}

type InstructionSet map[string]instrChars

var instructionSet = make(InstructionSet)

var ioFlags = [...]byte{' ', 'S', 'C', 'P'}
var ioTests = [...]string{"BN", "BZ", "DN", "DZ"}
var modes = [...]string{"Absolute", "PC", "AC2", "AC3"}
var skips = [...]string{"NONE", "SKP", "SZC", "SNC", "SZR", "SNR", "SEZ", "SBN"}

type DecodedInstr struct {
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
	disp              int32 // we express displacements as signed 32
	loHiBit           bool
	offset, immVal    int32
	imm16b            dg_word
	imm32b            dg_dword
	argCount, bitNum  int
	disassembly       string
}

var decodedInstr DecodedInstr

func instructionsInit() {

	// the following section was generated programmatically from the corresponding structure
	// in the C version of MV/Em.  Ideally, a tool would exist to maintain the structure in a
	// neutral form and both versions would use/import the data.

	instructionSet["ADC"] = instrChars{0x8400, 0x8700, NOVA_TWOACC_MULT_OP_FMT, 1, NOVA_OP}
	instructionSet["ADD"] = instrChars{0x8600, 0x8700, NOVA_TWOACC_MULT_OP_FMT, 1, NOVA_OP}
	instructionSet["ADDI"] = instrChars{0xe7f8, 0xe7ff, ONEACC_2_WORD_FMT, 2, EAGLE_OP}
	instructionSet["ADI"] = instrChars{0x8008, 0x87ff, IMM_ONEACC_FMT, 1, ECLIPSE_OP}
	instructionSet["ANC"] = instrChars{0x8188, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["AND"] = instrChars{0x8700, 0x8700, NOVA_TWOACC_MULT_OP_FMT, 1, NOVA_OP}
	instructionSet["ANDI"] = instrChars{0xc7f8, 0xe7ff, ONEACC_2_WORD_FMT, 2, EAGLE_OP}
	instructionSet["BAM"] = instrChars{0x97c8, 0xffff, UNIQUE_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["BKPT"] = instrChars{0xc789, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_PC}
	instructionSet["BLM"] = instrChars{0xb7c8, 0xffff, UNIQUE_1_WORD_FMT, 1, ECLIPSE_MEMREF}
	instructionSet["BTO"] = instrChars{0x8408, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_OP}
	instructionSet["BTZ"] = instrChars{0x8448, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_OP}
	instructionSet["CIO"] = instrChars{0x85e9, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["CIOI"] = instrChars{0x8509, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["CLM"] = instrChars{0x84f8, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_PC}
	instructionSet["CMP"] = instrChars{0xdfa8, 0xffff, UNIQUE_1_WORD_FMT, 1, ECLIPSE_MEMREF}
	instructionSet["CMT"] = instrChars{0xefa8, 0xffff, UNIQUE_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["CMV"] = instrChars{0xd7a8, 0xffff, UNIQUE_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["COB"] = instrChars{0x8588, 0x87ff, TWOACC_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["COM"] = instrChars{0x8000, 0x8700, NOVA_TWOACC_MULT_OP_FMT, 1, NOVA_OP}
	instructionSet["CRYTC"] = instrChars{0xa7e9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["CRYTO"] = instrChars{0xa7c9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["CRYTZ"] = instrChars{0xa7d9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["CTR"] = instrChars{0xe7a8, 0xffff, UNIQUE_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["CVWN"] = instrChars{0xe669, 0xe7ff, ONEACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["DAD"] = instrChars{0x8088, 0x87ff, TWOACC_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["DEQUE"] = instrChars{0xe7c9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["DERR"] = instrChars{0x8f09, 0x8fcf, DERR_FMT, 1, EAGLE_OP}
	instructionSet["DHXL"] = instrChars{0x8388, 0x87ff, IMM_ONEACC_FMT, 1, NOVA_OP}
	instructionSet["DHXR"] = instrChars{0x83c8, 0x87ff, IMM_ONEACC_FMT, 1, NOVA_OP}
	instructionSet["DIA"] = instrChars{0x6100, 0xe700, NOVA_DATA_IO_FMT, 1, NOVA_IO}
	instructionSet["DIB"] = instrChars{0x6300, 0xe700, NOVA_DATA_IO_FMT, 1, NOVA_IO}
	instructionSet["DIC"] = instrChars{0x6500, 0xe700, NOVA_DATA_IO_FMT, 1, NOVA_IO}
	instructionSet["DIV"] = instrChars{0xd7c8, 0xffff, UNIQUE_1_WORD_FMT, 1, ECLIPSE_OP}
	instructionSet["DIVS"] = instrChars{0xdfc8, 0xffff, UNIQUE_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["DIVX"] = instrChars{0xbfc8, 0xffff, UNIQUE_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["DLSH"] = instrChars{0x82c8, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_OP}
	instructionSet["DOA"] = instrChars{0x6200, 0xe700, NOVA_DATA_IO_FMT, 1, NOVA_IO}
	instructionSet["DOB"] = instrChars{0x6400, 0xe700, NOVA_DATA_IO_FMT, 1, NOVA_IO}
	instructionSet["DOC"] = instrChars{0x6600, 0xe700, NOVA_DATA_IO_FMT, 1, NOVA_IO}
	instructionSet["DSB"] = instrChars{0x80c8, 0x87ff, TWOACC_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["DSPA"] = instrChars{0xc478, 0xe4ff, ONEACC_MODE_IND_2_WORD_E_FMT, 2, ECLIPSE_PC}
	instructionSet["DSZ"] = instrChars{0x1800, 0xf800, NOVA_NOACC_EFF_ADDR_FMT, 1, NOVA_MEMREF}
	instructionSet["DSZTS"] = instrChars{0xc7d9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["ECLID"] = instrChars{0xffc8, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["EDIT"] = instrChars{0xf7a8, 0xffff, UNIQUE_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["EDSZ"] = instrChars{0x9c38, 0xfcff, NOACC_MODE_IND_2_WORD_E_FMT, 2, ECLIPSE_PC}
	instructionSet["EISZ"] = instrChars{0x9438, 0xfcff, NOACC_MODE_IND_2_WORD_E_FMT, 2, ECLIPSE_PC}
	instructionSet["EJMP"] = instrChars{0x8438, 0xfcff, NOACC_MODE_IND_2_WORD_E_FMT, 2, ECLIPSE_PC}
	instructionSet["EJSR"] = instrChars{0x8c38, 0xfcff, NOACC_MODE_IND_2_WORD_E_FMT, 2, ECLIPSE_PC}
	instructionSet["ELDA"] = instrChars{0xa438, 0xe4ff, ONEACC_MODE_IND_2_WORD_E_FMT, 2, ECLIPSE_MEMREF}
	instructionSet["ELDB"] = instrChars{0x8478, 0xe4ff, ONEACC_MODE_IND_2_WORD_E_FMT, 2, ECLIPSE_MEMREF}
	instructionSet["ELEF"] = instrChars{0xe438, 0xe4ff, ONEACC_MODE_IND_2_WORD_E_FMT, 2, ECLIPSE_OP}
	instructionSet["ENQH"] = instrChars{0xc7e9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["ENQT"] = instrChars{0xc7f9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["ESTA"] = instrChars{0xc438, 0xe4ff, ONEACC_MODE_IND_2_WORD_E_FMT, 2, ECLIPSE_OP}
	instructionSet["ESTB"] = instrChars{0xa478, 0xe4ff, ONEACC_MODE_2_WORD_E_FMT, 2, ECLIPSE_OP}
	instructionSet["FNS"] = instrChars{0x86a8, 0xffff, UNIQUE_1_WORD_FMT, 1, NOVA_PC}
	instructionSet["FSA"] = instrChars{0x8ea8, 0xffff, UNIQUE_1_WORD_FMT, 1, NOVA_PC}
	instructionSet["FXTD"] = instrChars{0xa779, 0xffff, UNIQUE_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["FXTE"] = instrChars{0xc749, 0xffff, UNIQUE_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["HALT"] = instrChars{0x647f, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["HLV"] = instrChars{0xc6f8, 0xe7ff, ONEACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["HXL"] = instrChars{0x8308, 0x87ff, IMM_ONEACC_FMT, 1, ECLIPSE_OP}
	instructionSet["HXR"] = instrChars{0x8348, 0x87ff, IMM_ONEACC_FMT, 1, ECLIPSE_OP}
	instructionSet["INC"] = instrChars{0x8300, 0x8700, NOVA_TWOACC_MULT_OP_FMT, 1, NOVA_OP}
	instructionSet["INTA"] = instrChars{0x633f, 0xe7ff, ONEACC_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["INTDS"] = instrChars{0x60bf, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["INTEN"] = instrChars{0x607f, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["IOR"] = instrChars{0x8108, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_OP}
	instructionSet["IORI"] = instrChars{0x87f8, 0xe7ff, ONEACC_2_WORD_FMT, 2, ECLIPSE_OP}
	instructionSet["IORST"] = instrChars{0x653f, 0xe73f, IO_RESET_FMT, 1, NOVA_IO}
	instructionSet["ISZ"] = instrChars{0x1000, 0xf800, NOVA_NOACC_EFF_ADDR_FMT, 1, NOVA_MEMREF}
	instructionSet["ISZTS"] = instrChars{0xc7c9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["JMP"] = instrChars{0x0, 0xf800, NOVA_NOACC_EFF_ADDR_FMT, 1, NOVA_PC}
	instructionSet["JSR"] = instrChars{0x800, 0xf800, NOVA_NOACC_EFF_ADDR_FMT, 1, NOVA_PC}
	instructionSet["LCALL"] = instrChars{0xa6c9, 0xe7ff, NOACC_MODE_IND_4_WORD_FMT, 4, EAGLE_MEMREF}
	instructionSet["LCPID"] = instrChars{0x8759, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["LDA"] = instrChars{0x2000, 0xe000, NOVA_ONEACC_EFF_ADDR_FMT, 1, NOVA_MEMREF}
	instructionSet["LDAFP"] = instrChars{0xc669, 0xe7ff, ONEACC_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["LDASB"] = instrChars{0xc649, 0xe7ff, ONEACC_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["LDASL"] = instrChars{0xa669, 0xe7ff, ONEACC_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["LDASP"] = instrChars{0xa669, 0xe7ff, ONEACC_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["LDATS"] = instrChars{0x8649, 0xe7ff, ONEACC_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["LDB"] = instrChars{0x85c8, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_OP}
	instructionSet["LDSP"] = instrChars{0x8519, 0x87ff, ONEACC_MODE_IND_3_WORD_FMT, 3, EAGLE_PC}
	instructionSet["LEF"] = instrChars{0x6000, 0xe000, NOVA_ONEACC_EFF_ADDR_FMT, 1, NOVA_MEMREF}
	instructionSet["LJMP"] = instrChars{0xa6d9, 0xe7ff, NOACC_MODE_IND_3_WORD_FMT, 3, EAGLE_PC}
	instructionSet["LJSR"] = instrChars{0xa6e9, 0xe7ff, NOACC_MODE_IND_3_WORD_FMT, 3, EAGLE_PC}
	instructionSet["LLDB"] = instrChars{0x84c9, 0x87ff, ONEACC_MODE_3_WORD_FMT, 3, EAGLE_OP}
	instructionSet["LLEF"] = instrChars{0x83e9, 0x87ff, ONEACC_MODE_IND_3_WORD_FMT, 3, EAGLE_OP}
	instructionSet["LLEFB"] = instrChars{0x84e9, 0x87ff, ONEACC_MODE_3_WORD_FMT, 3, EAGLE_OP}
	instructionSet["LMRF"] = instrChars{0x87c9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["LNADD"] = instrChars{0x8218, 0x87ff, ONEACC_MODE_IND_3_WORD_FMT, 3, EAGLE_OP}
	instructionSet["LNADI"] = instrChars{0x8618, 0x87ff, NOACC_MODE_IMM_IND_3_WORD_FMT, 3, EAGLE_OP}
	instructionSet["LNDIV"] = instrChars{0x82d8, 0x87ff, ONEACC_MODE_IND_3_WORD_FMT, 3, EAGLE_OP}
	instructionSet["LNDO"] = instrChars{0x8698, 0x87ff, LNDO_4_WORD_FMT, 4, EAGLE_PC}
	instructionSet["LNDSZ"] = instrChars{0x86d9, 0xe7ff, NOACC_MODE_IND_3_WORD_FMT, 3, EAGLE_PC}
	instructionSet["LNISZ"] = instrChars{0x86c9, 0xe7ff, NOACC_MODE_IND_3_WORD_FMT, 3, EAGLE_PC}
	instructionSet["LNLDA"] = instrChars{0x83c9, 0x87ff, ONEACC_MODE_IND_3_WORD_FMT, 3, EAGLE_MEMREF}
	instructionSet["LNMUL"] = instrChars{0x8298, 0x87ff, ONEACC_MODE_IND_3_WORD_FMT, 3, EAGLE_MEMREF}
	instructionSet["LNSBI"] = instrChars{0x8658, 0x87ff, NOACC_MODE_IMM_IND_3_WORD_FMT, 3, EAGLE_MEMREF}
	instructionSet["LNSTA"] = instrChars{0x83d9, 0x87ff, ONEACC_MODE_IND_3_WORD_FMT, 3, EAGLE_MEMREF}
	instructionSet["LNSUB"] = instrChars{0x8258, 0x87ff, ONEACC_MODE_IND_3_WORD_FMT, 3, EAGLE_MEMREF}
	instructionSet["LOB"] = instrChars{0x8508, 0x87ff, TWOACC_1_WORD_FMT, 1, NOVA_OP}
	instructionSet["LPEF"] = instrChars{0xa6f9, 0xe7ff, NOACC_MODE_IND_3_WORD_FMT, 3, EAGLE_MEMREF}
	instructionSet["LPEFB"] = instrChars{0xc6f9, 0xe7ff, NOACC_MODE_3_WORD_FMT, 3, EAGLE_OP}
	instructionSet["LPHY"] = instrChars{0x87e9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["LSH"] = instrChars{0x8288, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_OP}
	instructionSet["LWDO"] = instrChars{0x8798, 0x87ff, LNDO_4_WORD_FMT, 4, EAGLE_PC}
	instructionSet["LWDSZ"] = instrChars{0x86f9, 0xe7ff, NOACC_MODE_IND_3_WORD_FMT, 3, EAGLE_PC}
	instructionSet["LWISZ"] = instrChars{0x86e9, 0xe7ff, NOACC_MODE_IND_3_WORD_FMT, 3, EAGLE_PC}
	instructionSet["LWLDA"] = instrChars{0x83f9, 0x87ff, ONEACC_MODE_IND_3_WORD_FMT, 3, EAGLE_MEMREF}
	instructionSet["LWSTA"] = instrChars{0x84f9, 0x87ff, ONEACC_MODE_IND_3_WORD_FMT, 3, EAGLE_MEMREF}
	instructionSet["MOV"] = instrChars{0x8200, 0x8700, NOVA_TWOACC_MULT_OP_FMT, 1, NOVA_OP}
	instructionSet["MSKO"] = instrChars{0x643f, 0xe7ff, ONEACC_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["MUL"] = instrChars{0xc7c8, 0xffff, UNIQUE_1_WORD_FMT, 1, ECLIPSE_OP}
	instructionSet["NADD"] = instrChars{0x8049, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["NADDI"] = instrChars{0xc639, 0xe7ff, ONEACC_2_WORD_FMT, 2, EAGLE_OP}
	instructionSet["NADI"] = instrChars{0x8599, 0x87ff, IMM_ONEACC_FMT, 1, EAGLE_OP}
	instructionSet["NCLID"] = instrChars{0x683f, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["NEG"] = instrChars{0x8100, 0x8700, NOVA_TWOACC_MULT_OP_FMT, 1, NOVA_OP}
	instructionSet["NIO"] = instrChars{0x6000, 0xff00, IO_FLAGS_DEV_FMT, 1, EAGLE_IO}
	instructionSet["NLDAI"] = instrChars{0xc629, 0xe7ff, ONEACC_2_WORD_FMT, 2, EAGLE_OP}
	instructionSet["NMUL"] = instrChars{0x8069, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["NSBI"] = instrChars{0x85a9, 0x87ff, IMM_ONEACC_FMT, 1, EAGLE_OP}
	instructionSet["NSUB"] = instrChars{0x8059, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["PIO"] = instrChars{0x85d9, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["POP"] = instrChars{0x8688, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_STACK}
	instructionSet["POPJ"] = instrChars{0x9fc8, 0xffff, UNIQUE_1_WORD_FMT, 1, ECLIPSE_STACK}
	instructionSet["PRTRST"] = instrChars{0x85d9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["PRTSEL"] = instrChars{0x783f, 0xffff, UNIQUE_1_WORD_FMT, 1, NOVA_IO}
	instructionSet["PSH"] = instrChars{0x8648, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_STACK}
	instructionSet["PSHJ"] = instrChars{0x84b8, 0xfcff, NOACC_MODE_IND_2_WORD_E_FMT, 2, ECLIPSE_STACK}
	instructionSet["READS"] = instrChars{0x613f, 0xe7ff, ONEACC_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["RTN"] = instrChars{0xafc8, 0xffff, UNIQUE_1_WORD_FMT, 1, ECLIPSE_STACK}
	instructionSet["SAVE"] = instrChars{0xe7c8, 0xffff, UNIQUE_2_WORD_FMT, 2, ECLIPSE_STACK}
	instructionSet["SBI"] = instrChars{0x8048, 0x87ff, IMM_ONEACC_FMT, 1, ECLIPSE_OP}
	instructionSet["SEX"] = instrChars{0x8349, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["SGT"] = instrChars{0x8208, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_PC}
	instructionSet["SKP"] = instrChars{0x6700, 0xff00, IO_TEST_DEV_FMT, 1, NOVA_IO}
	instructionSet["SNB"] = instrChars{0x85f8, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_PC}
	instructionSet["SPSR"] = instrChars{0xa7a9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["SPTE"] = instrChars{0xe729, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["SSPT"] = instrChars{0xe7d9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["STA"] = instrChars{0x4000, 0xe000, NOVA_ONEACC_EFF_ADDR_FMT, 1, NOVA_MEMREF}
	instructionSet["STAFP"] = instrChars{0xc679, 0xe7ff, ONEACC_1_WORD_FMT, 1, EAGLE_STACK}
	instructionSet["STASB"] = instrChars{0xc659, 0xe7ff, ONEACC_1_WORD_FMT, 1, EAGLE_STACK}
	instructionSet["STASL"] = instrChars{0xa679, 0xe7ff, ONEACC_1_WORD_FMT, 1, EAGLE_STACK}
	instructionSet["STASP"] = instrChars{0xa659, 0xe7ff, ONEACC_1_WORD_FMT, 1, EAGLE_STACK}
	instructionSet["STATS"] = instrChars{0x8659, 0xe7ff, ONEACC_1_WORD_FMT, 1, EAGLE_STACK}
	instructionSet["STB"] = instrChars{0x8608, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_OP}
	instructionSet["SUB"] = instrChars{0x8500, 0x8700, NOVA_TWOACC_MULT_OP_FMT, 1, NOVA_OP}
	instructionSet["SZB"] = instrChars{0x8488, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_OP}
	instructionSet["SZBO"] = instrChars{0x84c8, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_OP}
	instructionSet["WADC"] = instrChars{0x8249, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WADD"] = instrChars{0x8149, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WADDI"] = instrChars{0x8689, 0xe7ff, ONEACC_IMM_3_WORD_FMT, 3, EAGLE_OP}
	instructionSet["WADI"] = instrChars{0x84b9, 0x87ff, IMM_ONEACC_FMT, 1, EAGLE_OP}
	instructionSet["WANC"] = instrChars{0x8549, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WAND"] = instrChars{0x8449, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WANDI"] = instrChars{0x8699, 0xe7ff, ONEACC_IMM_3_WORD_FMT, 3, EAGLE_OP}
	instructionSet["WASH"] = instrChars{0x8279, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WASHI"] = instrChars{0xc6a9, 0xe7ff, ONEACC_IMM_2_WORD_FMT, 2, EAGLE_OP}
	instructionSet["WBLM"] = instrChars{0xe749, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_MEMREF}
	instructionSet["WBR"] = instrChars{0x8038, 0x843f, SPLIT_8BIT_DISP_FMT, 1, EAGLE_PC}
	instructionSet["WBTO"] = instrChars{0x8299, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WBTZ"] = instrChars{0x82a9, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WCLM"] = instrChars{0x8569, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WCOM"] = instrChars{0x8459, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WINC"] = instrChars{0x8259, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WIORI"] = instrChars{0x86a9, 0xe7ff, ONEACC_IMM_3_WORD_FMT, 3, EAGLE_OP}
	instructionSet["WLDAI"] = instrChars{0xc689, 0xe7ff, ONEACC_IMM_3_WORD_FMT, 3, EAGLE_OP}
	instructionSet["WLMP"] = instrChars{0xa7f9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_IO}
	instructionSet["WLSH"] = instrChars{0x8559, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WLSHI"] = instrChars{0xe6d9, 0xe7ff, ONEACC_2_WORD_FMT, 2, EAGLE_OP}
	instructionSet["WLSI"] = instrChars{0x85b9, 0x87ff, IMM_ONEACC_FMT, 1, EAGLE_OP}
	instructionSet["WMOV"] = instrChars{0x8379, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WMOVR"] = instrChars{0xe699, 0xe7ff, ONEACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WNEG"] = instrChars{0x8269, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["WRTN"] = instrChars{0x87a9, 0xffff, UNIQUE_1_WORD_FMT, 1, EAGLE_PC}
	instructionSet["WSBI"] = instrChars{0x8589, 0x87ff, IMM_ONEACC_FMT, 1, EAGLE_OP}
	instructionSet["WSEQ"] = instrChars{0x80b9, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_PC}
	instructionSet["WSEQI"] = instrChars{0xe6c9, 0xe7ff, ONEACC_IMM_2_WORD_FMT, 2, EAGLE_PC}
	instructionSet["WSGE"] = instrChars{0x8199, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_PC}
	instructionSet["WSGT"] = instrChars{0x81b9, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_PC}
	instructionSet["WSGTI"] = instrChars{0xe689, 0xe7ff, ONEACC_IMM_2_WORD_FMT, 2, EAGLE_PC}
	instructionSet["WSKBO"] = instrChars{0x8f49, 0x8fcf, WSKB_FMT, 1, EAGLE_PC}
	instructionSet["WSKBZ"] = instrChars{0x8f89, 0x8fcf, WSKB_FMT, 1, EAGLE_PC}
	instructionSet["WSLE"] = instrChars{0x81a9, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_PC}
	instructionSet["WSLEI"] = instrChars{0xe6a9, 0xe7ff, ONEACC_IMM_2_WORD_FMT, 2, EAGLE_PC}
	instructionSet["WSLT"] = instrChars{0x8289, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_PC}
	instructionSet["WSNE"] = instrChars{0x8189, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_PC}
	instructionSet["WSNEI"] = instrChars{0xe6e9, 0xe7ff, ONEACC_IMM_2_WORD_FMT, 2, EAGLE_PC}
	instructionSet["WSTB"] = instrChars{0x8539, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_MEMREF}
	instructionSet["WSUB"] = instrChars{0x8159, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}
	instructionSet["XCALL"] = instrChars{0x8609, 0xe7ff, NOACC_MODE_IND_3_WORD_FMT, 3, EAGLE_PC}
	instructionSet["XCH"] = instrChars{0x81c8, 0x87ff, TWOACC_1_WORD_FMT, 1, ECLIPSE_OP}
	instructionSet["XJMP"] = instrChars{0xc609, 0xe7ff, NOACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_PC}
	instructionSet["XJSR"] = instrChars{0xc619, 0xe7ff, NOACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_PC}
	instructionSet["XLDB"] = instrChars{0x8419, 0x87ff, ONEACC_MODE_2_WORD_X_B_FMT, 2, EAGLE_MEMREF}
	instructionSet["XLEF"] = instrChars{0x8409, 0x87ff, ONEACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_MEMREF}
	instructionSet["XLEFB"] = instrChars{0x8439, 0x87ff, ONEACC_MODE_2_WORD_X_B_FMT, 2, EAGLE_MEMREF}
	instructionSet["XNADD"] = instrChars{0x8018, 0x87ff, ONEACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_MEMREF}
	instructionSet["XNADI"] = instrChars{0x8418, 0x87ff, IMM_MODE_2_WORD_FMT, 2, EAGLE_MEMREF}
	instructionSet["XNDO"] = instrChars{0x8498, 0x87ff, THREE_WORD_DO_FMT, 3, EAGLE_MEMREF}
	instructionSet["XNDSZ"] = instrChars{0xa609, 0xe7ff, NOACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_PC}
	instructionSet["XNISZ"] = instrChars{0x8639, 0xe7ff, NOACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_PC}
	instructionSet["XNLDA"] = instrChars{0x8329, 0x87ff, ONEACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_MEMREF}
	instructionSet["XNSBI"] = instrChars{0x8458, 0x87ff, IMM_MODE_2_WORD_FMT, 2, EAGLE_MEMREF}
	instructionSet["XNSTA"] = instrChars{0x8339, 0x87ff, ONEACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_MEMREF}
	instructionSet["XNSUB"] = instrChars{0x8058, 0x87ff, IMM_MODE_2_WORD_FMT, 2, EAGLE_MEMREF}
	instructionSet["XPEF"] = instrChars{0x86e9, 0xe7ff, NOACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_MEMREF}
	instructionSet["XSTB"] = instrChars{0x8429, 0x87ff, ONEACC_MODE_2_WORD_X_B_FMT, 2, EAGLE_MEMREF}
	instructionSet["XWADD"] = instrChars{0x8118, 0x87ff, ONEACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_MEMREF}
	instructionSet["XWADI"] = instrChars{0x8518, 0x87ff, IMM_MODE_2_WORD_FMT, 2, EAGLE_MEMREF}
	instructionSet["XWLDA"] = instrChars{0x8309, 0x87ff, ONEACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_MEMREF}
	instructionSet["XWSBI"] = instrChars{0x8558, 0x87ff, IMM_MODE_2_WORD_FMT, 2, EAGLE_OP}
	instructionSet["XWSTA"] = instrChars{0x8319, 0x87ff, ONEACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_MEMREF}
	instructionSet["XWSUB"] = instrChars{0x8158, 0x87ff, ONEACC_MODE_IND_2_WORD_X_FMT, 2, EAGLE_OP}
	instructionSet["ZEX"] = instrChars{0x8359, 0x87ff, TWOACC_1_WORD_FMT, 1, EAGLE_OP}

	log.Printf("INFO: %d Instruction Set Opcodes loaded\n", len(instructionSet))
}

func instructionFind(opcode dg_word, lefMode bool, ioOn bool, atuOn bool) string {
	var tail dg_word
	for mn, insChar := range instructionSet {
		if (opcode & insChar.mask) == insChar.bits {
			// there are some exceptions to the normal decoding...
			switch mn {
			case "LEF":
				if lefMode {
					return mn
				}
			case "ADC", "ADD", "AND", "COM", "INC", "MOV", "NEG", "SUB":
				tail = opcode & 0x000f
				if tail < 8 || tail > 9 {
					return mn
				}
			default:
				return mn

			}
		}
	}
	return ""
}

/* Decode an opcode
N.B. For the moment this function both decodes and disassembles the given instruction,
for performance in the future these two tasks should probably either be separated or
controlled by flags passed into the function.
*/
func instructionDecode(opcode dg_word, pc dg_phys_addr, lefMode bool, ioOn bool, autOn bool) (*DecodedInstr, bool) {
	var secondWord, thirdWord, fourthWord dg_word
	var tmp8bit dg_byte

	decodedInstr = DecodedInstr{}
	decodedInstr.disassembly = "; Unknown instruction"
	mnem := instructionFind(opcode, lefMode, ioOn, autOn)
	if mnem == "" {
		log.Printf("INFO: instructionFind failed to return anything to instructionDecode for location %d\n", pc)
		return &decodedInstr, false
	}
	decodedInstr.mnemonic = mnem
	decodedInstr.disassembly = mnem
	decodedInstr.instrFmt = instructionSet[mnem].instrFmt
	decodedInstr.instrType = instructionSet[mnem].instrType
	decodedInstr.instrLength = instructionSet[mnem].instrLen

	switch decodedInstr.instrFmt {

	case NOVA_NOACC_EFF_ADDR_FMT:
		decodedInstr.ind = decodeIndirect(testWbit(opcode, 5))
		decodedInstr.mode = decodeMode(getWbits(opcode, 6, 2))
		decodedInstr.disp = decode8bitDisp(dg_byte(opcode&0x00ff), decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %c%d.%s",
			decodedInstr.ind, decodedInstr.disp, modeToString(decodedInstr.mode))

	case NOVA_ONEACC_EFF_ADDR_FMT:
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		decodedInstr.ind = decodeIndirect(testWbit(opcode, 5))
		decodedInstr.mode = decodeMode(getWbits(opcode, 6, 2))
		decodedInstr.disp = decode8bitDisp(dg_byte(opcode&0x00ff), decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%c%d.%s",
			decodedInstr.acd, decodedInstr.ind, decodedInstr.disp, modeToString(decodedInstr.mode))

	case NOVA_TWOACC_MULT_OP_FMT:
		decodedInstr.acs = decodeAcc(getWbits(opcode, 1, 2))
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		decodedInstr.sh = decodeShift(getWbits(opcode, 8, 2))
		decodedInstr.c = decodeCarry(getWbits(opcode, 10, 2))
		decodedInstr.nl = decodeNoLoad(testWbit(opcode, 12))
		decodedInstr.skip = decodeSkip(getWbits(opcode, 13, 3))
		decodedInstr.disassembly += fmt.Sprintf("%c%c%c %d,%d %s",
			decodedInstr.c, decodedInstr.sh, decodedInstr.nl, decodedInstr.acs, decodedInstr.acd, skipToString(decodedInstr.skip))

	case NOVA_DATA_IO_FMT:
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		decodedInstr.f = decodeIOFlags(getWbits(opcode, 8, 2))
		decodedInstr.ioDev = int(getWbits(opcode, 10, 6))
		decodedInstr.disassembly += fmt.Sprintf("%c %d,%s",
			decodedInstr.f, decodedInstr.acd, deviceToString(decodedInstr.ioDev))

	case IO_FLAGS_DEV_FMT:
		decodedInstr.f = decodeIOFlags(getWbits(opcode, 8, 2))
		decodedInstr.ioDev = int(getWbits(opcode, 10, 6))
		decodedInstr.disassembly += fmt.Sprintf("%c %s",
			decodedInstr.f, deviceToString(decodedInstr.ioDev))

	case IO_RESET_FMT:
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2)) // TODO is this needed/used?

	case IO_TEST_DEV_FMT:
		decodedInstr.t = decodeIOTest(getWbits(opcode, 8, 2))
		decodedInstr.ioDev = int(getWbits(opcode, 10, 6))
		decodedInstr.disassembly += fmt.Sprintf("%s %s", decodedInstr.t, deviceToString(decodedInstr.ioDev))

	case UNIQUE_1_WORD_FMT:
		// nothing to do in this case

	case UNIQUE_2_WORD_FMT:
		decodedInstr.imm16b = memReadWord(pc + 1)
		decodedInstr.disassembly += fmt.Sprintf(" %d. [2-Word OpCode]", int16(decodedInstr.imm16b))

	case ONEACC_1_WORD_FMT:
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		decodedInstr.disassembly += fmt.Sprintf(" %d", decodedInstr.acd)

	case TWOACC_1_WORD_FMT: // eg. WSUB
		decodedInstr.acs = decodeAcc(getWbits(opcode, 1, 2))
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d", decodedInstr.acs, decodedInstr.acd)

	case ONEACC_2_WORD_FMT: // eg. ADDI, IORI, NADDI, NLDAI, WLSHI
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		decodedInstr.imm16b = memReadWord(pc + 1)
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%d [2-Word OpCode]", int16(decodedInstr.imm16b), decodedInstr.acd)

	case NOACC_MODE_3_WORD_FMT: // eg. LPEFB
		decodedInstr.mode = decodeMode(getWbits(opcode, 3, 2))
		decodedInstr.imm32b = memReadDWord(pc + 1)
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%s [3-Word OpCode]", decodedInstr.imm32b, modeToString(decodedInstr.mode))

	case NOACC_MODE_IND_2_WORD_E_FMT, NOACC_MODE_IND_2_WORD_X_FMT:
		log.Printf("X_FMT: Mnemonic is <%s>\n", decodedInstr.mnemonic)
		switch decodedInstr.mnemonic {
		case "XJMP", "XJSR", "XNDSZ", "XNISZ", "XPEF":
			decodedInstr.mode = decodeMode(getWbits(opcode, 3, 2))
		case "EDSZ", "EISZ", "EJMP", "EJSR", "PSHJ":
			decodedInstr.mode = decodeMode(getWbits(opcode, 6, 2))
		}
		secondWord = memReadWord(pc + 1)
		decodedInstr.ind = decodeIndirect(testWbit(secondWord, 0))
		switch decodedInstr.mnemonic {
		case "EJSR", "EJMP": // FIXME - maybe more exceptions needed here
			decodedInstr.disp = decode15bitEclipseDisp(secondWord, decodedInstr.mode)
		default:
			decodedInstr.disp = decode15bitDisp(secondWord, decodedInstr.mode)
		}
		decodedInstr.disassembly += fmt.Sprintf(" %c%d.%s [2-Word OpCode]",
			decodedInstr.ind, decodedInstr.disp, modeToString(decodedInstr.mode))

	case NOACC_MODE_IND_3_WORD_FMT: // eg. LJMP/LJSR, LNDSZ, LWDS
		decodedInstr.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		thirdWord = memReadWord(pc + 2)
		decodedInstr.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstr.disp = decode31bitDisp(secondWord, thirdWord, decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %c%d.%s [3-Word OpCode]",
			decodedInstr.ind, decodedInstr.disp, modeToString(decodedInstr.mode))

	case LNDO_4_WORD_FMT:
		decodedInstr.acd = decodeAcc(getWbits(opcode, 1, 2))
		decodedInstr.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		thirdWord = memReadWord(pc + 2)
		fourthWord = memReadWord(pc + 3)
		decodedInstr.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstr.disp = decode31bitDisp(secondWord, thirdWord, decodedInstr.mode)
		decodedInstr.offset = int32(fourthWord)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d.,%c%d%s [4-Word OpCode]",
			decodedInstr.acd, decodedInstr.offset, decodedInstr.ind, decodedInstr.disp, modeToString(decodedInstr.mode))

	case NOACC_MODE_IMM_IND_3_WORD_FMT: // eg. LNADI, LNSBI
		decodedInstr.immVal = decode2bitImm(getWbits(opcode, 1, 2))
		decodedInstr.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		thirdWord = memReadWord(pc + 2)
		decodedInstr.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstr.disp = decode31bitDisp(secondWord, thirdWord, decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%c%d.%s [3-Word OpCode]",
			decodedInstr.immVal, decodedInstr.ind, decodedInstr.disp, modeToString(decodedInstr.mode))

	case NOACC_MODE_IND_4_WORD_FMT: // eg. LCALL
		decodedInstr.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		thirdWord = memReadWord(pc + 2)
		fourthWord = memReadWord(pc + 3)
		decodedInstr.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstr.disp = decode31bitDisp(secondWord, thirdWord, decodedInstr.mode)
		decodedInstr.argCount = int(fourthWord)
		decodedInstr.disassembly += fmt.Sprintf(" %c%d.%s,%d. [4-Word OpCode]",
			decodedInstr.ind, decodedInstr.disp, modeToString(decodedInstr.mode), decodedInstr.argCount)

	case ONEACC_MODE_2_WORD_X_B_FMT: // eg. XLDB, XLEFB, XSTB
		decodedInstr.mode = decodeMode(getWbits(opcode, 1, 2))
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstr.disp, decodedInstr.loHiBit = decode16bitByteDisp(secondWord)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d.+%c%s [2-Word OpCode]",
			decodedInstr.acd, decodedInstr.disp*2, loHiToByte(decodedInstr.loHiBit), modeToString(decodedInstr.mode))

	case ONEACC_MODE_2_WORD_E_FMT: // eg. ELDB, ESTB
		decodedInstr.mode = decodeMode(getWbits(opcode, 6, 2))
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstr.disp, decodedInstr.loHiBit = decode16bitByteDisp(secondWord)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d.+%c%s [2-Word OpCode]",
			decodedInstr.acd, decodedInstr.disp*2, loHiToByte(decodedInstr.loHiBit), modeToString(decodedInstr.mode))

	case ONEACC_MODE_IND_2_WORD_E_FMT: // eg. ELDA
		decodedInstr.mode = decodeMode(getWbits(opcode, 6, 2))
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstr.ind = decodeIndirect(testWbit(secondWord, 0))
		switch decodedInstr.mnemonic {
		case "ELEF": // FIXME - more exceptions needed here...
			decodedInstr.disp = decode15bitEclipseDisp(secondWord, decodedInstr.mode)
		default:
			decodedInstr.disp = decode15bitDisp(secondWord, decodedInstr.mode)
		}
		decodedInstr.disassembly += fmt.Sprintf(" %d,%c%d.%s [2-Word OpCode]",
			decodedInstr.acd, decodedInstr.ind, decodedInstr.disp, modeToString(decodedInstr.mode))

	case ONEACC_MODE_IND_2_WORD_X_FMT: // eg. XNLDA/XWSTA
		decodedInstr.mode = decodeMode(getWbits(opcode, 1, 2))
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstr.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstr.disp = decode15bitDisp(secondWord, decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%c%d.%s [2-Word OpCode]",
			decodedInstr.acd, decodedInstr.ind, decodedInstr.disp, modeToString(decodedInstr.mode))

	case ONEACC_MODE_3_WORD_FMT: // eg. LLDB, LLEFB
		decodedInstr.mode = decodeMode(getWbits(opcode, 1, 2))
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		thirdWord = memReadWord(pc + 2)
		decodedInstr.disp = decode31bitDisp(secondWord, thirdWord, decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d.%s [3-Word OpCode]",
			decodedInstr.acd, decodedInstr.disp, modeToString(decodedInstr.mode))

	case ONEACC_MODE_IND_3_WORD_FMT: // eg. LWLDA/LWSTA,LNLDA
		decodedInstr.mode = decodeMode(getWbits(opcode, 1, 2))
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstr.ind = decodeIndirect(testWbit(secondWord, 0))
		thirdWord = memReadWord(pc + 2)
		decodedInstr.disp = decode31bitDisp(secondWord, thirdWord, decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %d,%c%d.%s [3-Word OpCode]",
			decodedInstr.acd, decodedInstr.ind, decodedInstr.disp, modeToString(decodedInstr.mode))

	case ONEACC_IMM_IND_3_WORD_FMT:
		decodedInstr.immVal = decode2bitImm(getWbits(opcode, 1, 2))
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstr.ind = decodeIndirect(testWbit(secondWord, 0))
		thirdWord = memReadWord(pc + 2)
		decodedInstr.disp = decode31bitDisp(secondWord, thirdWord, decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%d.%c%d. [3-Word OpCode]",
			decodedInstr.immVal, decodedInstr.acd, decodedInstr.ind, decodedInstr.disp)

	case IMM_ONEACC_FMT: // eg. HXL, NADI, WLSI
		decodedInstr.immVal = decode2bitImm(getWbits(opcode, 1, 2))
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%d", decodedInstr.immVal, decodedInstr.acd)

	case IMM_MODE_2_WORD_FMT: // eg. XNADI
		decodedInstr.immVal = decode2bitImm(getWbits(opcode, 1, 2))
		decodedInstr.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstr.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstr.disp = decode15bitDisp(secondWord, decodedInstr.mode)
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%d.%s [2-Word OpCode]",
			decodedInstr.immVal, decodedInstr.disp, modeToString(decodedInstr.mode))

	case THREE_WORD_DO_FMT: // eg. XNDO
		decodedInstr.acd = decodeAcc(getWbits(opcode, 1, 2))
		decodedInstr.mode = decodeMode(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstr.ind = decodeIndirect(testWbit(secondWord, 0))
		decodedInstr.disp = decode15bitDisp(secondWord, decodedInstr.mode)
		thirdWord = memReadWord(pc + 2)
		decodedInstr.offset = int32(uint16(thirdWord))
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d. %c%d.%s [3-Word OpCode]",
			decodedInstr.acd, decodedInstr.offset, decodedInstr.ind, decodedInstr.disp, modeToString(decodedInstr.mode))

	case ONEACC_IMM_2_WORD_FMT:
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		secondWord = memReadWord(pc + 1)
		decodedInstr.imm16b = secondWord
		decodedInstr.disassembly += fmt.Sprintf(" %d,%d. [2-Word OpCode]", decodedInstr.acd, decodedInstr.imm16b)

	case ONEACC_IMM_3_WORD_FMT:
		decodedInstr.acd = decodeAcc(getWbits(opcode, 3, 2))
		decodedInstr.imm32b = memReadDWord(pc + 1)
		decodedInstr.disassembly += fmt.Sprintf(" %d.,%d [3-Word OpCode]", decodedInstr.imm32b, decodedInstr.acd)

	case SPLIT_8BIT_DISP_FMT:
		tmp8bit = dg_byte(getWbits(opcode, 1, 4) & 0xff)
		tmp8bit = tmp8bit << 4
		tmp8bit |= dg_byte(getWbits(opcode, 6, 4) & 0xff)
		decodedInstr.disp = decode8bitDisp(tmp8bit, "PC")
		decodedInstr.disassembly += fmt.Sprintf(" %d.", int32(decodedInstr.disp))

	case WSKB_FMT:
		tmp8bit = dg_byte(getWbits(opcode, 1, 3) & 0xff)
		tmp8bit = tmp8bit << 2
		tmp8bit |= dg_byte(getWbits(opcode, 10, 2) & 0xff)
		decodedInstr.bitNum = int(uint8(tmp8bit))
		decodedInstr.disassembly += fmt.Sprintf(" %d.", decodedInstr.bitNum)

	default:
		log.Printf("ERROR: Invalid instruction format (%d) for instruction %s", decodedInstr.instrFmt, decodedInstr.mnemonic)
		return nil, false
	}
	return &decodedInstr, true
}

/* decoders for (parts of) operands below here... */

func decode2bitImm(i dg_word) int32 {
	// to expand range (by 1!) 1 is subtracted from operand
	return int32(i + 1)
}

// TODO Do we need this?
func decode8bitDisp(d8 dg_byte, mode string) int32 {
	var disp int32
	if mode == "Absolute" {
		disp = int32(d8) & 0x00ff // unsigned offset
	} else {
		// signed offset...
		disp = int32(int8(d8)) // this should sign-extend
	}
	return disp
}

func decode15bitDisp(d15 dg_word, mode string) int32 {
	var disp int32
	if mode == "Absolute" {
		disp = int32(d15 & 0x7fff) // zero extend
	} else {
		if testWbit(d15, 1) {
			disp = int32(int16(d15 | 0x8000)) // sign extend
		} else {
			disp = int32(d15 & 0x7fff) // zero extend
		}
		if mode == "PC" {
			disp++ // see p.1-12 of PoP
		}
	}
	log.Printf("... decode15bitDisp got: %d, returning: %d\n", d15, disp)
	return disp
}

func decode15bitEclipseDisp(d15 dg_word, mode string) int32 {
	var disp int32
	if mode == "Absolute" {
		disp = int32(d15 & 0x7fff) // zero extend
	} else {
		if testWbit(d15, 1) {
			disp = int32(int16(d15 | 0xc000)) // sign extend
		} else {
			disp = int32(d15 & 0x3fff) // zero extend
		}
		if mode == "PC" {
			disp++ // see p.1-12 of PoP
		}
	}
	log.Printf("... decode15bitEclispeDisp got: %d, returning: %d\n", d15, disp)
	return disp
}

func decode16bitByteDisp(d16 dg_word) (int32, bool) {
	var disp int32
	loHi := testWbit(d16, 15)
	disp = int32(d16 >> 1)
	log.Printf("... decode16bitByteDisp got: %d, returning %d\n", d16, disp)
	return disp, loHi
}

func decode31bitDisp(d1, d2 dg_word, mode string) int32 {
	// FIXME Test this!
	var disp int32
	if testWbit(d1, 1) {
		disp = int32(int16(d1 | 0x8000)) // sign extend
	} else {
		disp = int32(int16(d1)) & 0x00007fff // zero extend
	}
	disp = (disp << 16) | (int32(d2) & 0x0000ffff)
	if mode == "PC" {
		disp++ // see p.1-12 of PoP
	}
	log.Printf("... decode31bitDisp got: %d %d, returning: %d\n", d1, d2, disp)
	return disp
}

func decodeAcc(ac dg_word) int {
	return int(ac)
}

func decodeCarry(cry dg_word) byte {
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

func decodeIOFlags(fl dg_word) byte {
	return ioFlags[fl]
}

func decodeIOTest(t dg_word) string {
	return ioTests[t]
}

func decodeMode(ix dg_word) string {
	return modes[ix]
}

func decodeNoLoad(n bool) byte {
	if n {
		return '#'
	}
	return ' '
}

func decodeShift(sh dg_word) byte {
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

func decodeSkip(skp dg_word) string {
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
