// decoder_test.go
package main

import (
	"testing"

	"github.com/SMerrony/dgemug/dg"
)

func Test2bitImm(t *testing.T) {
	ttable := []struct {
		i dg.WordT
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
	var db dg.ByteT
	var md int
	var res int16

	db = 7
	md = absoluteMode
	res = decode8bitDisp(db, md)
	if res != 7 {
		t.Error("Expected 7, got ", res)
	}

	db = 7
	md = pcMode
	res = decode8bitDisp(db, md)
	if res != 7 {
		t.Error("Expected 7, got ", res)
	}

	db = 0xff
	md = pcMode
	res = decode8bitDisp(db, md)
	if res != -1 {
		t.Error("Expected -1, got ", res)
	}
}

func TestDecode15bitDisp(t *testing.T) {
	var disp15 dg.WordT
	var m int
	var res int16

	disp15 = 300
	m = absoluteMode
	res = decode15bitDisp(disp15, m)
	if res != 300 {
		t.Error("Absolute Mode: Expected 300, got ", res)
	}

	disp15 = 300
	m = ac2Mode
	res = decode15bitDisp(disp15, m)
	if res != 300 {
		t.Error("AC2 Mode: Expected 300, got ", res)
	}

	disp15 = 0x7ed4
	m = ac2Mode
	res = decode15bitDisp(disp15, m)
	if res != -300 {
		t.Error("AC2 Mode: Expected -300, got ", res)
	}

	disp15 = 0x7ed4
	m = pcMode
	res = decode15bitDisp(disp15, m)
	if res != -299 {
		t.Error("PC Mode: Expected -299, got ", res)
	}
}
