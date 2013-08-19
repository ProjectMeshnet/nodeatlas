// NodeAtlas - Federated mapping for mesh networks
//
// Copyright (C) 2013 Alexander Bauer, Luke Evers, Daniel Supernault,
// Dylan Whichard, and contributors
//
// This program is free software: you can redistribute it and/or
// modify it under the terms of the GNU General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see http://www.gnu.org/licenses/
//
package main
// Copyright (C) 2013 Alexander Bauer, Luke Evers, Daniel Supernault,
// Dylan Whichard, and contributors; (GPLv3) see LICENSE or doc.go

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/inhies/go-log"
	"html/template"
	"net"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"
)

var Version = "0.5.11"

var (
	LogLevel = log.LogLevel(log.INFO)
	LogFlags = log.Ldate | log.Ltime // 2006/01/02 15:04:05
	LogFile  = os.Stdout             // *os.File type
)

var (
	Conf      *Config
	StaticDir string // Directory for compiled files.
	Pulse     *time.Ticker

	t *template.Template
	l *log.Logger

	shutdown = sync.NewCond(&sync.Mutex{})
)

var (
	ignoreServerCrash bool
)

var (
	fConf = flag.String("conf", "conf.json", "path to configuration file")
	fRes  = flag.String("res", "res/", "path to resource directory")

	fLog   = flag.String("file", "", "Logfile (defaults to stdout)")
	fDebug = flag.Bool("debug", false, "maximize verbosity")
	fQuiet = flag.Bool("q", false, "only output errors")

	fReadOnly = flag.Bool("readonly", false, "disallow database changes")

	fImport = flag.String("import", "", "import a JSON array of nodes")
	fTestDB = flag.Bool("testdb", false, "test the database")
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

	// Set logging parameters based on flags.
	if *fDebug {
		LogLevel = log.DEBUG
		LogFlags |= log.Lshortfile // Include the filename and line
	} else if *fQuiet {
		LogLevel = log.ERR
	}

	if len(*fLog) > 0 {
		// If a file is specified, open it with the appropriate flags,
		// which will cause it to be created if not existent, and only
		// append data when writing to it. It will inherit all
		// permissions from its parent directory.
		LogFile, err = os.OpenFile(*fLog,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Could not open logfile: %s", err)
			os.Exit(1)
		}
		defer LogFile.Close()
	} // Otherwise, default to os.Stdout.

	l, err = log.NewLevel(LogLevel, true, LogFile, "", LogFlags)
	if err != nil {
		fmt.Printf("Could start logger: %s", err)
		os.Exit(1)
	}

	l.Infof("Starting NodeAtlas %s\n", Version)

	// Compile and template the static directory.
	StaticDir, err = CompileStatic(*fRes, Conf)
	if err != nil {
		if len(StaticDir) > 0 {
			// Try to remove the directory if it was created.
			err := os.RemoveAll(StaticDir)
			if err != nil {
				l.Emergf("Could not remove static directory %q: %s",
					StaticDir, err)
			}
		}
		l.Fatalf("Could not compile static files: %s", err)
	}
	l.Debugf("Compiled static files to %q\n", StaticDir)

	if *fTestDB {
		// Open up a temporary database using sqlite3.
		tempDB := path.Join(os.TempDir(), "nodeatlas-test.db")

		db, err := sql.Open("sqlite3", tempDB)
		if err != nil {
			l.Fatalf("Could not connect to temporary database: %s", err)
		}

		// Perform all of the tests in sequence.
		TestDatabase(DB{db, false})
		err = db.Close()
		if err != nil {
			l.Emergf("Could not close temporary database: %s", err)
		}

		// Finally, remove the database file and exit.
		err = os.Remove(tempDB)
		if err != nil {
			l.Emergf("Could not remove temporary database %q: %s",
				tempDB, err)
		}
		return
	}

	// Connect to the database with configured parameters.
	db, err := sql.Open(Conf.Database.DriverName,
		Conf.Database.Resource)
	if err != nil {
		l.Fatalf("Could not connect to database: %s", err)
	}
	// Wrap the *sql.DB type.
	Db = DB{
		DB:       db,
		ReadOnly: (*fReadOnly || Conf.Database.ReadOnly),
	}
	l.Debug("Connected to database\n")
	if Db.ReadOnly {
		l.Warning("Database is read only\n")
	}

	// Initialize the database with all of its tables.
	err = Db.InitializeTables()
	if err != nil {
		l.Fatalf("Could not initialize database: %s", err)
	}
	l.Debug("Initialized database\n")
	l.Infof("Nodes: %d (%d local)\n", Db.LenNodes(true), Db.LenNodes(false))

	// Check action flags and abandon normal startup if any are set.
	if len(*fImport) != 0 {
		err := ImportFile(*fImport)
		if err != nil {
			l.Fatalf("Import failed: %s", err)
		} else {
			l.Printf("Import successful!")
		}
		return
	}

	// Listen for OS signals.
	go ListenSignal()

	// Set up the initial RSS feed so that it can be served once
	// online. If there is an error, log it, but continue starting up.
	if err = GenerateNodeRSS(); err != nil {
		l.Errf("Error generating Node RSS feed: %s", err)
	}

	// Start the Heartbeat.
	Heartbeat()
	l.Debug("Heartbeat started\n")

	go func() {
		// Start the HTTP server. This will block until the server
		// encounters an error.
		err = StartServer()
		if err != nil && !ignoreServerCrash {
			l.Fatalf("Server crashed: %s", err)
		}
	}()

	// Finally, block until told to exit.
	shutdown.L.Lock()
	shutdown.Wait()
}

// Heartbeat starts a time.Ticker to perform tasks on a regular
// schedule, as set by Conf.HeartbeatRate, which are documented
// below. The global variable Pulse is its ticker. To restart the
// timer, invoke Heartbeat() again.
//
// Tasks:
// - Db.DeleteExpiredFromQueue()
// - UpdateMapCache()
func Heartbeat() {
	// If the timer was not nil, then the timer must restart.
	if Pulse != nil {
		// If we are resetting, stop the existing timer before
		// replacing it.
		Pulse.Stop()
	}

	// Otherwise, create the Ticker and spawn a goroutine to check it.
	Pulse = time.NewTicker(time.Duration(Conf.HeartbeatRate))
	go func() {
		for {
			if _, ok := <-Pulse.C; !ok {
				// If the channel closes, warn that the heartbeat has
				// stopped.
				l.Warning("Heartbeat stopped\n")
				return
			}
			// Otherwise, perform scheduled tasks.
			doHeartbeatTasks()
		}
	}()
}

// doHeartbeatTasks is the underlying function which is executed by
// Heartbeat() at regular intervals. It can be called directly to
// perform the tasks that are usually performed regularly.
func doHeartbeatTasks() {
	l.Debug("Heartbeat\n")
	Db.DeleteExpiredFromQueue()
	UpdateMapCache()
	ClearExpiredCAPTCHA()
	ResendVerificationEmails()
}

// ListenSignal uses os/signal to wait for OS signals, such as SIGHUP
// and SIGINT, and perform the appropriate actions as listed below.
//     SIGHUP: reload configuration file
//     SIGINT, SIGKILL, SIGTERM: gracefully shut down
func ListenSignal() {
	// Create the channel and use signal.Notify to listen for any
	// specified signals.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1, syscall.SIGUSR2,
		os.Interrupt, os.Kill, syscall.SIGTERM)
	for sig := range c {
		switch sig {
		case syscall.SIGUSR1:
			l.Info("Forced heartbeat\n")
			doHeartbeatTasks()
		case syscall.SIGUSR2:
			l.Info("Reloading config\n")

			// Reload the configuration, but keep the old one if
			// there's an error.
			conf, err := ReadConfig(*fConf)
			if err != nil {
				l.Errf("Could not read conf; using old one: %s", err)
				continue
			}
			Conf = conf

			// Recompile the static directory, but be able to restore
			// the previous one if there's an error.
			oldStaticDir := StaticDir

			StaticDir, err = CompileStatic(*fRes, Conf)
			if err != nil {
				l.Errf("Error recompiling static directory: %s", err)
				StaticDir = oldStaticDir
				continue
			}

			// Remove the old one, and report if there's an error, but
			// continue even if there's an error.
			err = os.RemoveAll(oldStaticDir)
			if err != nil {
				l.Errf("Error removing old static directory: %s", err)
			}

			// Restart the heartbeat ticker.
			Heartbeat()
		case os.Interrupt, os.Kill, syscall.SIGTERM:
			l.Infof("Caught %s; NodeAtlas over and out\n", sig)
			var err error

			// Close the HTTP Listener. If a UNIX socket is in use, it
			// will automatically be removed. We need to tell the
			// server to ignore errors, because closing the listener
			// will cause http.Server.Serve() to return one.
			ignoreServerCrash = true
			listener.Close()

			// Close the database connection.
			err = Db.Close()
			if err != nil {
				// If closing the database gave an error, report it
				// and set the exit code.
				l.Errf("Database could not be closed: %s", err)
			}

			// Delete the directory of static files.
			err = os.RemoveAll(StaticDir)
			if err != nil {
				// If the static directory coldn't be removed, report
				// it, give the location of the directory, and set the
				// exit code.
				l.Errf("Static directory %q could not be removed: %s",
					StaticDir, err)
			}

			// Finally, tell the main routine to stop waiting and
			// exit.
			shutdown.Broadcast()
		}
	}
}

func TestDatabase(db DB) {
	err := db.InitializeTables()
	if err != nil {
		l.Fatalf("Could not initialize tables: %s", err)
	}
	l.Debug("Successfully initialized tables\n")

	node := &Node{
		Addr:       IP(net.ParseIP("ff00::1")),
		OwnerName:  "nodeatlas",
		OwnerEmail: "admin@example.org",
		Latitude:   80.01010,
		Longitude:  -80.10101,
		Status:     uint32(0),
	}

	nodeCached := &Node{
		Addr:         IP(net.ParseIP("ff00::2")),
		OwnerName:    "test",
		OwnerEmail:   "nothing@example.com",
		Latitude:     34.14523,
		Longitude:    5.3635,
		Status:       uint32(0),
		RetrieveTime: time.Now().Unix(),
	}

	err = db.AddNode(node)

	if err != nil {
		l.Errf("Error adding node: %s", err)
	} else {
		l.Debug("Successfully added node\n")
	}

	l.Debugf("Nodes: %d", db.LenNodes(false))

	node.Status = StatusActive
	err = db.UpdateNode(node)
	if err != nil {
		l.Errf("Error updating node: %s", err)
	} else {
		l.Debug("Successfully updated node")
	}

	ip := IP(net.ParseIP("ff00::1"))
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

	nodes := []*Node{node, nodeCached}

	err = db.CacheNodes(nodes)
	if err != nil {
		l.Errf("Error caching nodes: %s", err)
	} else {
		l.Debug("Successfully cached nodes")
	}
}
