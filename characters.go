// ASCII and DASHER special characters
package main

const (
	ASCII_BEL = 0x07
	ASCII_BS  = 0x08
	ASCII_TAB = 0x09
	ASCII_NL  = 0x0A
	ASCII_FF  = 0x0C
	ASCII_CR  = 0x0D
	ASCII_ESC = 0x1B

	DASHER_ERASE_EOL         = 013
	DASHER_ERASE_PAGE        = 014
	DASHER_CURSOR_LEFT       = 0x19
	DASHER_WRITE_WINDOW_ADDR = 020 //followed by col then row
	DASHER_DIM_ON            = 034
	DASHER_DIM_OFF           = 035
	DASHER_UNDERLINE         = 024
	DASHER_NORMAL            = 025
	DASHER_DELETE            = 0x7F
)
