package main

import (
	"testing"
	"fmt"
	"os"
)

var err error
var conf *Config

func TestReadConfig(t *testing.T) {
	fmt.Println("Testing Read Config")
	conf, err = ReadConfig("conf.json.example")
	if err != nil {
		t.Errorf("Could not read config: %s", err)
	}
}

func TestWriteConfig(t *testing.T) {
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
