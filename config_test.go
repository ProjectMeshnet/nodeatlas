package main

import (
	"testing"
	"fmt"
)

var err error

func TestReadConfig(t *testing.T) {
	fmt.Println("Testing Read Config")
	_, err = ReadConfig("conf.json.example")
	if err != nil {
		t.Errorf("Could not read config: %s", err)
	}
}
