package main

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"database/sql"
)

var (
	Db DB
)

type DB struct {
	*sql.DB
}

// InitializeTables issues the commands to create all tables and
// columns. Will not create the tables if they exist already. If
// it encounters an error, it is returned.
func (db DB) InitializeTables() (err error) {
	// First, create the 'nodes' table.
	_, err = db.Query(`CREATE TABLE IF NOT EXISTS nodes (
address BINARY(16) PRIMARY KEY,
owner VARCHAR(255) NOT NULL,
lat FLOAT NOT NULL,
lon FLOAT NOT NULL,
status INT NOT NULL,
updated DATETIME DEFAULT CURRENT_TIMESTAMP);`)
	if err != nil {
		return
	}
	return
}

// LenNodes returns the number of nodes in the database. If there is
// an error, it returns zero and logs the incident.
func (db DB) LenNodes() (n int) {
	// Count the number of rows in the 'nodes' table.
	row := db.QueryRow("SELECT COUNT(*) FROM nodes;")
	// Write that number to n, and return if there is no
	// error. Otherwise, log it and return zero.
	if err := row.Scan(&n); err != nil {
		l.Errf("Error counting the number of nodes: %s", err)
		n = 0
	}
	return
}

