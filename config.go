package main

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	// Name is the string by which this instance of NodeAtlas will be
	// referred to. It usually describes the entire project name or
	// the region about which it focuses.
	Name string

	// Hostname is the address which NodeAtlas should identify itself
	// as. For example, in a verification email, NodeAtlas would give
	// the verification link as
	// http://<hostname><prefix>/<long-random-id>
	Hostname string

	// Prefix is the URL prefix which is required to access the front
	// end. For example, with a prefix of "/nodeatlas", NodeAtlas
	// would be able to respond to http://example.com/nodeatlas.
	Prefix string

	// Addr is the network interface and port to which NodeAtlas
	// should bind. This of the form "0.0.0.0:8077" (for global
	// binding on port 8077) or "127.0.0.1:8077" (for only localhost
	// binding on port 8077) or ":8077" (which is equivalent to
	// "0.0.0.0:8077".)
	Addr string

	// ChildMaps is a list of addresses from which to pull lists of
	// nodes every heartbeat. Please note that these maps are trusted
	// fully, and they could easily introduce false nodes to the
	// database temporarily (until cleared by the CacheExpiration.
	ChildMaps []string

	// Database is the structure which contains the database driver
	// name, such as "sqlite3" or "mysql", and the database resource,
	// such as a path to .db file, or username, password, and name.
	Database struct {
		DriverName string
		Resource   string
		ReadOnly   bool
	}

	// HeartbeatRate is the amount of time to wait between performing
	// regular tasks, such as clearing expired nodes from the queue
	// and cache.
	HeartbeatRate Duration

	// CacheExpiration is the amount of time for which to store cached
	// nodes before considering them outdated, and removing them.
	CacheExpiration Duration

	// VerificationExpiration is the amount of time to allow users to
	// verify nodes by email after initially placing them. See the
	// documentation for time.ParseDuration for format information.
	VerificationExpiration Duration

	// SMTP contains the information necessary to connect to a mail
	// relay, so as to send verification email to registered nodes. If
	// it is nil, then node registration will be disabled.
	SMTP *struct {
		// EmailAddress will be given as the "From" address when
		// sending email.
		EmailAddress string

		// NoAuthenticate determines whether NodeAtlas should attempt to
		// authenticate with the SMTP relay or not. Unless the relay
		// is local, leave this false.
		NoAuthenticate bool

		// Username and Password are the credentials required by the
		// server to log in.
		Username, Password string

		// ServerAddress is the address of the SMTP relay, including
		// the port.
		ServerAddress string
	}

	// Map contains the information used by NodeAtlas to power the
	// Leaflet.js map.
	Map struct {
		// Tileserver is the URL used for loading tiles. It is of the
		// form "http://{s}.tile.osm.org/{z}/{x}/{y}.png", so that
		// Leaflet.js can use it.
		Tileserver string

		// Center contains the coordinates on which to center the map.
		Center struct {
			Latitude, Longitude float64
		}

		// Zoom is the Leaflet.js zoom level to start the map at.
		Zoom int
	}
}

// ReadConfig uses os and encoding/json to read a configuration from
// the filesystem. It returns any errors it encounters.
func ReadConfig(path string) (conf *Config, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	conf = &Config{}
	err = json.NewDecoder(f).Decode(conf)
	return
}

// WriteConfig uses os and encoding/json to write a configuration to
// the filesystem. It creates the file if it doesn't exist and returns
// any errors it encounters.
func WriteConfig(conf *Config, path string) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(conf)
	return
}

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	if b[0] != '"' {
		// If the duration is not a string, then consider it to be the
		// zero duration, so we do not have to set it.
		return nil
	}
	dur, err := time.ParseDuration(string(b[1 : len(b)-1]))
	if err != nil {
		return err
	}
	*d = *(*Duration)(&dur)
	return nil
}
