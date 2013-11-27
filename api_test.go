package main

import (
	"testing"
	"fmt"
)

type PGP struct {
	Key   string
	Valid bool
}

func TestPGP(t *testing.T) {
	
	testPGP := []PGP{
		{"CAFEBABE", true},
		{"caFeD00d", true},
		{"12345678", true},
		{"ORANGEEE", false},
		{"12O21333", false},
		{"CAFEBABECAFEBABE", true},
		{"C2AccAb3CDDDad2F", true},
		{"1234567890ABCDEF", true},
		{"3ABC12AAAAC2134D", true},
		{"ABC12CCAADD333X9", false},
	}
	
	for i, p := range testPGP {
		fmt.Println(i)
		if PGPRegexp.Match([]byte(p.Key)) && !p.Valid {
			t.Errorf("%s matched when it should not have!", p.Key)
		}
	}
}

func TestEmail(t *testing.T) {

}

