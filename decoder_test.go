// decoder_test.go
package main

import (
	"testing"
)

func TestDecodeMode(t *testing.T) {
	var modeIxdg.WordT
	var decMode string

	modeIx = 0
	decMode = decodeMode(modeIx)
	if decMode != "Absolute" {
		t.Error("Expected <Absolute>, got ", decMode)
	}

	modeIx = 1
	decMode = decodeMode(modeIx)
	if decMode != "PC" {
		t.Error("Expected <PC>, got ", decMode)
	}
}

func Test2bitImm(t *testing.T) {
	ttable := []struct {
		idg.WordT
		o uint16
	}{
		{0, 1},
		{1, 2},
		{3, 4},
	}
	for _, tt := range ttable {
		res := decode2bitImm(tt.i)
		if res != tt.o {
			t.Errorf("Expected %d, got %d", res, tt.o)
		}
	}
}

func TestDecode8bitDisp(t *testing.T) {
	var dbdg.ByteT
	var md string
	var res int16

	db = 7
	md = "Absolute"
	res = decode8bitDisp(db, md)
	if res != 7 {
		t.Error("Expected 7, got ", res)
	}

	db = 7
	md = "PC"
	res = decode8bitDisp(db, md)
	if res != 7 {
		t.Error("Expected 7, got ", res)
	}

	db = 0xff
	md = "PC"
	res = decode8bitDisp(db, md)
	if res != -1 {
		t.Error("Expected -1, got ", res)
	}
}

func TestDecode15bitDisp(t *testing.T) {
	var disp15dg.WordT
	var m string
	var res int16

	disp15 = 300
	m = "Absolute"
	res = decode15bitDisp(disp15, m)
	if res != 300 {
		t.Error("Absolute Mode: Expected 300, got ", res)
	}

	disp15 = 300
	m = "AC2"
	res = decode15bitDisp(disp15, m)
	if res != 300 {
		t.Error("AC2 Mode: Expected 300, got ", res)
	}

	disp15 = 0x7ed4
	m = "AC2"
	res = decode15bitDisp(disp15, m)
	if res != -300 {
		t.Error("AC2 Mode: Expected -300, got ", res)
	}

	disp15 = 0x7ed4
	m = "PC"
	res = decode15bitDisp(disp15, m)
	if res != -299 {
		t.Error("PC Mode: Expected -299, got ", res)
	}
}
