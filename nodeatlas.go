package main

// Copyright (C) 2013 Alexander Bauer, Luke Evers, Daniel Supernault,
// Dylan Whichard, and contributors; (GPLv3) see LICENSE or doc.go

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/inhies/go-log"
	"html/template"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var Version = "0.5.12"

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

	// defaultResLocation and defaultConfLocation are used as flag
	// defaults.
	defaultResLocation  = "res/"
	defaultConfLocation = "conf.json"
)

var (
	fConf = flag.String("conf", defaultConfLocation,
		"path to configuration file")
	fRes = flag.String("res", defaultResLocation,
		"path to resource directory")

	fLog   = flag.String("file", "", "Logfile (defaults to stdout)")
	fDebug = flag.Bool("debug", false, "maximize verbosity")
	fQuiet = flag.Bool("q", false, "only output errors")

	fReadOnly = flag.Bool("readonly", false, "disallow database changes")

	fImport = flag.String("import", "", "import a JSON array of nodes")
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

	// Connect to the database with configured parameters.
	db, err := sql.Open(Conf.Database.DriverName,
		Conf.Database.Resource)
	if err != nil {
		l.Fatalf("Could not connect to database: %s", err)
	}
	// Wrap the *sql.DB type.
	Db = DB{
		DB:         db,
		DriverName: Conf.Database.DriverName,
		ReadOnly:   (*fReadOnly || Conf.Database.ReadOnly),
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
	// online. If there is an error, it will be logged, but won't
	// prevent startup.
	CleanNodeRSS()

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
	CleanNodeRSS()
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
