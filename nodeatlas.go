package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/inhies/go-utils/log"
	"net/http"
	"os"
)

var Version = "0"

var (
	Conf *Config
	l    *log.Logger
)

var (
	fConf = flag.String("conf", "conf.json", "path to configuration file")
	fRes  = flag.String("res", "res/", "path to template and image directory")
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
	l.Infof("Starting NodeAtlas v%s\n", Version)

	// Connect to the database with configured parameters.
	db, err := sql.Open(Conf.Database.DriverName,
		Conf.Database.Resource)
	if err != nil {
		l.Fatalf("Could not connect to database: %s", err)
	}
	Db = DB{db} // Wrap the *sql.DB type.
	l.Debug("Connected to database\n")

	// Initialize the database with all of its tables.
	err = Db.InitializeTables()
	if err != nil {
		l.Fatalf("Could not initialize database: %s", err)
	}
	l.Debug("Initialized database\n")

	// Start the HTTP server.
	err = StartServer(Conf.Addr, Conf.Prefix)
	if err != nil {
		l.Fatalf("Server crashed: %s", err)
	}
}

// StartServer is a simple helper function to register any handlers
// (such as the API) and start the HTTP server on the given
// address. If it crashes, it returns the error.
func StartServer(addr, prefix string) (err error) {
	// Register any handlers.
	RegisterAPI(prefix)
	l.Debug("Registered API handler\n")

	err = RegisterTemplates()
	if err != nil {
		return
	}
	http.HandleFunc("/", HandleRoot)
	http.HandleFunc("/res/", HandleRes)

	// Start the HTTP server and return any errors if it crashes.
	l.Infof("Starting HTTP server on %q\n", addr)
	return http.ListenAndServe(addr, nil)
}
