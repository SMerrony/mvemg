// tto - console output
package main

import (
	"net"
)

type Tto struct {
	conn net.Conn
}

func (t Tto) ttoPutChar(c byte) {
	t.conn.Write([]byte{c})
}

func (t Tto) ttoPutString(s string) {
	t.conn.Write([]byte(s))
}

func (t Tto) ttoPutStringNL(s string) {
	t.conn.Write([]byte(s))
	t.conn.Write([]byte{ASCII_NL})
}

func (t Tto) ttoPutNLString(s string) {
	t.conn.Write([]byte{ASCII_NL})
	t.conn.Write([]byte(s))
}

func (t Tto) ttoReset() {

}
