package main

// Copyright (C) 2013 Alexander Bauer, Luke Evers, Dylan Whichard,
// and contributors; (GPLv3) see LICENSE or doc.go

import (
	"encoding/json"
	"io"
	"os"
)

// Import reads a slice of JSON-encoded Nodes from the given io.Reader
// and adds them to the database. Cache-related fields such as
// RetrieveTime are discarded.
func Import(r io.Reader) (err error) {
	// Unmarshal the Nodes from the given io.Reader.
	nodes := make([]*Node, 0)
	err = json.NewDecoder(r).Decode(&nodes)
	if err != nil {
		return
	}

	// Insert them into the database as new. Timestamps will be the
	// current time.
	err = Db.AddNodes(nodes)
	return
}

// ImportFile opens the given file and imports JSON-encoded Nodes from
// it.
func ImportFile(path string) (err error) {
	// Open the file in readonly mode.
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	// Pass it along to Import.
	return Import(f)
}
