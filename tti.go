// tti - console input
package main

import (
	"log"
	"net"
	"os"
)

type Tti struct {
	conn net.Conn
}

func (t Tti) ttiGetChar() byte {
	b := make([]byte, 80)
	n, err := t.conn.Read(b)
	if err != nil || n == 0 {
		log.Println("ERROR: could not read from console port: ", err.Error())
		os.Exit(1)
	}
	return b[0]
}

func (t Tti) ttiReset() {

}
