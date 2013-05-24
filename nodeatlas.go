package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/inhies/go-utils/log"
	"net"
	"net/http"
	"os"
)

var Version = "0.1"

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
	l.Infof("Starting NodeAtlas %s\n", Version)

	// Connect to the database with configured parameters.
	db, err := sql.Open(Conf.Database.DriverName,
		Conf.Database.Resource)
	if err != nil {
		l.Fatalf("Could not connect to database: %s", err)
	}
	Db = DB{db} // Wrap the *sql.DB type.
	l.Debug("Connected to database\n")

	TestDatabase(&Db)

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

func TestDatabase(db *DB) {
	err := db.InitializeTables()
	if err != nil {
		l.Fatalf("Could not initialize tables: %s", err)
	}
	l.Debug("Successfully initialized tables")

	nLocal := db.LenNodes(false)
	nTotal := db.LenNodes(true)
	nCached := nTotal - nTotal
	l.Debugf("Nodes: %d (%d local, %d cached)", nTotal, nLocal, nCached)

	node := Node{
		Addr:       net.ParseIP("ff00::1"),
		OwnerName:  "nodeatlas",
		OwnerEmail: "admin@example.org",
		Latitude:   80.01010,
		Longitude:  -80.10101,
		Status:     StatusActive,
	}
	err = db.AddNode(&node)

	if err != nil {
		l.Errf("Error adding node: %s", err)
	} else {
		l.Debug("Successfully added node")
	}

	l.Debugf("Nodes: %d", db.LenNodes(false))

	node.OwnerName = "DuoNoxSol"
	err = db.UpdateNode(&node)
	if err != nil {
		l.Errf("Error updating node: %s", err)
	} else {
		l.Debug("Successfully updated node")
	}

	ip := net.ParseIP("ff00::1")
	_, err = db.GetNode(ip)
	if err != nil {
		l.Errf("Error retrieving node: %s", err)
	} else {
		l.Debug("Successfully got node")
	}

	err = db.DeleteNode(node.Addr)
	if err != nil {
		l.Errf("Error deleting node: %s", err)
	} else {
		l.Debug("Successfully deleted node")
	}
}
