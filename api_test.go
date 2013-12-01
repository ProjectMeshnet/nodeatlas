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
	fmt.Println("Testing PGP regex")
	
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
		fmt.Print(i)
		fmt.Print(" - " + p.Key + " is valid? ")
		fmt.Print(PGPRegexp.Match([]byte(p.Key)))
		fmt.Print(" (should be ")
		fmt.Print(p.Valid)
		fmt.Println(")")
		if PGPRegexp.Match([]byte(p.Key)) != p.Valid {
			t.Errorf("%s returned %v when it should have returned %v!", p.Key, !p.Valid, p.Valid)
		}
	}
}

func TestEmail(t *testing.T) {

}
