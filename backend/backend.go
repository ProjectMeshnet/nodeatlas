package main

import (
	"github.com/inhies/go-utils/log"
	"os"
)

var (
	l *log.Logger
)

func main() {
	// Begin API
	var err error
	l, err = log.NewLevel(log.DEBUG, true, os.Stdout, "", log.Ldate|log.Ltime)
	if err != nil {
		os.Exit(1)
	}

	err = StartAPI()
	if err != nil {
		l.Fatal("API crashed:", err)
	}
}
