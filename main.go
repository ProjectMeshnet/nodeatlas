package main

import (
	"flag"
	"fmt"
	"github.com/inhies/go-utils/log"
	"net/http"
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

	// Start the HTTP server.
	err = StartServer(Conf.Addr, Conf.Prefix)
	if err != nil {
		l.Fatal("Server crashed:", err)
	}

}

// StartServer is a simple helper function to register any handlers
// (such as the API) and start the HTTP server on the given
// address. If it crashes, it returns the error.
func StartServer(addr, prefix string) error {
	// Register any handlers.
	RegisterAPI(prefix)

	// Start the HTTP server and return any errors if it crashes.
	return http.ListenAndServe(addr, nil)
}
