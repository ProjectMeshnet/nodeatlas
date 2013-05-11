package main

import (
	"flag"
	"fmt"
	"github.com/inhies/go-utils/log"
	"os"
)

var (
	Conf *Config

	l *log.Logger
)

var (
	fConf = flag.String("conf", "conf.json", "path to configuration file")
)

func main() {
	flag.Parse()

	// Load the configuration.
	var err error
	Conf, err = ReadConfig(*fConf)
	if err != nil {
		fmt.Printf("Could not read conf: %s", err)
		os.Exit(1)
	}
	// Touch up any fields if they're in common but unusable forms.
	if len(Conf.Prefix) == 0 {
		Conf.Prefix = "/"
	}

	l, err = log.NewLevel(log.DEBUG, true, os.Stdout, "",
		log.Ldate|log.Ltime)
	if err != nil {
		fmt.Printf("Could start logger: %s", err)
		os.Exit(1)
	}

	// Begin the API.
	err = StartAPI(Conf.APIAddr, Conf.Prefix)
	if err != nil {
		l.Fatal("API crashed:", err)
	}
	
}
