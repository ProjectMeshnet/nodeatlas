package main

import (
	"testing"
	"fmt"
	"os"
)

var err error
var conf *Config

func TestReadConfig(t *testing.T) {
	// func ReadConfig(path string) (conf *Config, err error)
	fmt.Println("Testing Read Config")
	conf, err = ReadConfig("conf.json.example")
	if err != nil {
		t.Errorf("Could not read config: %s", err)
	}
}

func TestWriteConfig(t *testing.T) {
	// func WriteConfig(conf *Config, path string) (err error)
	fmt.Println("Testing Write Config")
	err = WriteConfig(conf, "conf.write.test")
	if err != nil {
		t.Errorf("Could not write config: %s", err)
	}
	err = os.Remove("conf.write.test")
	if err != nil {
		t.Errorf("Could not delete test config: %s", err)
	}
}

func TestMarshalJSON(t *testing.T) {
	// func (d Duration) MarshalJSON() ([]byte, error)
}

func TestUnmarshalJSON(t *testing.T) {
	// There are two funcs for UnmarshallJSON, so this test func
	// will test them both rather than having two more test funcs
	//
	// func (d *Duration) UnmarshalJSON(b []byte) error
	// func (n *IPNet) UnmarshalJSON(b []byte) error
}
