package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	// Name is the string by which this instance of NodeAtlas will be
	// referred to. It usually describes the entire project name or
	// the region about which it focuses.
	Name string

	// Prefix is the URL prefix which is required to access the front
	// end. For example, with a prefix of "/nodeatlas", NodeAtlas
	// would be able to respond to http://example.com/nodeatlas.
	Prefix string

	// Addr is the network interface and port to which the NodeAtlas
	// front end should bind. This of the form "0.0.0.0:8080" (for
	// global binding on port 8080) or "127.0.0.1:8080" (for only
	// localhost binding on port 8080) or ":8080" (which is equivalent
	// to "0.0.0.0:8080".)
	Addr string

	// APIAddr is the network interface and port to which the
	// NodeAtlas API should bind. It conforms to the same syntax as
	// Addr.
	APIAddr string
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
