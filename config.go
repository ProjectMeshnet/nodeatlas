package main

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"time"
)

type Config struct {
	// Name is the string by which this instance of NodeAtlas will be
	// referred to. It usually describes the entire project name or
	// the region about which it focuses.
	Name string

	// AdminContact is the information shown in the about menu or when
	// displaying errors. It should be at least a name and email
	// address, such as "Your Name <email@example.com>" with an
	// optional PGP key ID, such as "Your Name <email@example.com>
	// 0123ABCD"
	AdminContact string

	// Web is the structure which contains information relating to the
	// backend of the HTTP webserver.
	Web struct {
		// Hostname is the address which NodeAtlas should identify
		// itself as. For example, in a verification email, NodeAtlas
		// would give the verification link as
		// http://<hostname><prefix>/verify/<long-random-id>
		Hostname string

		// Prefix is the URL prefix which is required to access the
		// front end. For example, with a prefix of "/nodeatlas",
		// NodeAtlas would be able to respond to
		// http://example.com/nodeatlas.
		Prefix string

		// Addr is the network protocol, interface, and port to which
		// NodeAtlas should bind. For example, "tcp://0.0.0.0:8077"
		// will bind globally to the 8077 TCP port, and
		// "unix://nodeatlas.sock" will create a UNIX socket at
		// nodeatlas.sock.
		Addr string

		// DeproxyHeaderFields is a list of HTTP header fields that
		// should be used instead of the connecting IP when verifying
		// nodes and logging major errors. They must be in
		// canonicalized form, such as "X-Forwarded-For" or
		// "X-Real-IP".
		DeproxyHeaderFields []string

		// HeaderSnippet is a snippet of code which is inserted into
		// the <head> of each page. For example, one could include a
		// script tieing into Pikwik.
		HeaderSnippet string
	}

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
	// relay, so as to send verification email to registered nodes.
	SMTP *struct {
		// VerifyDisabled controls whether email verification is used
		// for newly registered nodes. If it is false or omitted, an
		// email will be sent using the SMTP settings defined in this
		// struct.
		VerifyDisabled bool

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
		// Favicon is the icon to be displayed in the browser when
		// viewing the map. It is a filename to be loaded from
		// `<*fRes>/icon/`.
		Favicon string

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

		// ClusterRadius is the range (in pixels) at which markers on
		// the map will cluster together.
		ClusterRadius int

		// Attribution is the "map data" copyright notice placed at
		// the bottom right of the map, meant to credit the
		// maintainers of the tileserver.
		Attribution string
	}

	// Verify contains the list of steps used to ensure that new nodes
	// are valid when registered. They can be enabled or disabled
	// according to one's needs.
	Verify struct {
		// Netmask, if not nil, is a CIDR-form network mask which
		// requires that nodes registered have an Addr which matches
		// it. For example, "fc00::/8" would only allow IPv6 addresses
		// in which the first two digits are "fc", and
		// "192.168.0.0/16" would only allow IPv4 addresses in which
		// the first two bytes are "192.168".
		Netmask *IPNet

		// FromNode requires the verification request (GET
		// /api/verify?id=<long_random_id>) to originate from the
		// address of the node that is being verified.
		FromNode bool
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

// IPNet is a wrapper for net.IPNet which implements json.Unmarshaler.

type IPNet net.IPNet

var InvalidIPNetError = errors.New("network mask is invalid")

func (n *IPNet) UnmarshalJSON(b []byte) error {
	if b[0] != '"' {
		// If the IPNet is not given as a string, then it is invalid
		// and should return an error.
		return InvalidIPNetError
	}
	_, ipnet, err := net.ParseCIDR(string(b[1 : len(b)-1]))
	if err != nil {
		return err
	}
	*n = *(*IPNet)(ipnet)
	return nil
}
