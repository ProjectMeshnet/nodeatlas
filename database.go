package main

// Copyright (C) 2013 Alexander Bauer, Luke Evers, Daniel Supernault,
// Dylan Whichard, and contributors; (GPLv3) see LICENSE or doc.go

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

var (
	Db DB
)

type DB struct {
	*sql.DB
	DriverName string
	ReadOnly   bool
}

// InitializeTables issues the commands to create all tables and
// columns. Will not create the tables if they exist already. If
// it encounters an error, it is returned.
func (db DB) InitializeTables() (err error) {
	// First, create the 'nodes' table.
	_, err = db.Query(`CREATE TABLE IF NOT EXISTS nodes (
address BINARY(16) PRIMARY KEY,
owner VARCHAR(255) NOT NULL,
email VARCHAR(255) NOT NULL,
contact VARCHAR(255),
details VARCHAR(255),
pgp BINARY(8),
lat FLOAT NOT NULL,
lon FLOAT NOT NULL,
status INT NOT NULL,
updated INT NOT NULL);`)
	if err != nil {
		return
	}
	_, err = db.Query(`CREATE TABLE IF NOT EXISTS nodes_cached (
address BINARY(16) PRIMARY KEY,
owner VARCHAR(255) NOT NULL,
details VARCHAR(255),
lat FLOAT NOT NULL,
lon FLOAT NOT NULL,
status INT NOT NULL,
source INT NOT NULL,
retrieved INT NOT NULL);`)
	if err != nil {
		return
	}
	_, err = db.Query(`CREATE TABLE IF NOT EXISTS nodes_verify_queue (
id INT PRIMARY KEY,
address BINARY(16) NOT NULL,
owner VARCHAR(255) NOT NULL,
email VARCHAR(255) NOT NULL,
contact VARCHAR(255),
details VARCHAR(255),
pgp BINARY(8),
lat FLOAT NOT NULL,
lon FLOAT NOT NULL,
status INT NOT NULL,
verifysent BOOL NOT NULL,
expiration INT NOT NULL);`)
	if err != nil {
		return
	}
	// <mysql> SQL? A standard? Hahahaha!
	if db.DriverName == "mysql" {
		_, err = db.Query(`CREATE TABLE IF NOT EXISTS cached_maps (
id INTEGER PRIMARY KEY AUTO_INCREMENT,
hostname VARCHAR(255) NOT NULL,
name VARCHAR(255) NOT NULL);`)
	} else {
		_, err = db.Query(`CREATE TABLE IF NOT EXISTS cached_maps (
id INTEGER PRIMARY KEY AUTOINCREMENT,
hostname VARCHAR(255) NOT NULL,
name VARCHAR(255) NOT NULL);`)
	}
	if err != nil {
		return
	}

	_, err = db.Query(`CREATE TABLE IF NOT EXISTS captcha (
id BINARY(32) NOT NULL,
solution BINARY(6) NOT NULL,
expiration INT NOT NULL);`)
	if err != nil {
		return
	}

	return
}

// LenNodes returns the number of nodes in the database. If there is
// an error, it returns -1 and logs the incident.
func (db DB) LenNodes(useCached bool) (n int) {
	// Count the number of rows in the 'nodes' table.
	var row *sql.Row
	if useCached {
		row = db.QueryRow("SELECT COUNT(*) FROM (SELECT address FROM nodes UNION SELECT address FROM nodes_cached) AS cachedNodes;")
	} else {
		row = db.QueryRow("SELECT COUNT(*) FROM nodes;")
	}
	// Write that number to n, and return if there is no
	// error. Otherwise, log it and return zero.
	if err := row.Scan(&n); err != nil {
		l.Errf("Error counting the number of nodes: %s", err)
		n = -1
	}
	return
}

// DumpNodes returns an array containing all nodes in the database,
// including both local and cached nodes.
func (db DB) DumpNodes() (nodes []*Node, err error) {
	// Begin by getting the required length of the array. If we get
	// -1, then there has been an error.
	if n := db.LenNodes(true); n != -1 {
		// If successful, initialize the array with the length.
		nodes = make([]*Node, n)
	} else {
		// Otherwise, error out.
		l.Errf("Could not count number of nodes in database\n")
		return nil, errors.New("Could not count number of nodes")
	}

	// Perform the query.
	rows, err := db.Query(`
SELECT address,owner,contact,details,pgp,lat,lon,status,0
FROM nodes
UNION SELECT address,owner,"",details,"",lat,lon,status,source
FROM nodes_cached;`)
	if err != nil {
		l.Errf("Error dumping database: %s", err)
		return
	}
	defer rows.Close()

	// Now, loop through, initialize the nodes, and fill them out
	// using only the selected columns.
	for i := 0; rows.Next(); i++ {
		// Initialize the node and put it in the table.
		node := new(Node)
		nodes[i] = node

		// Create temporary values to simplify scanning.
		contact := sql.NullString{}
		details := sql.NullString{}

		// Scan all of the values into it.
		err = rows.Scan(&node.Addr, &node.OwnerName,
			&contact, &details, &node.PGP,
			&node.Latitude, &node.Longitude, &node.Status, &node.SourceID)
		if err != nil {
			l.Errf("Error dumping database: %s", err)
			return
		}

		node.Contact = contact.String
		node.Details = details.String
	}
	return
}

// DumpLocal returns a slice containing all of the local nodes in the
// database.
func (db DB) DumpLocal() (nodes []*Node, err error) {
	// Begin by getting the required length of the array. If we get
	// -1, then there has been an error.
	if n := db.LenNodes(true); n != -1 {
		// If successful, initialize the array with the length.
		nodes = make([]*Node, n)
	} else {
		// Otherwise, error out.
		l.Errf("Could not count number of nodes in database\n")
		return nil, errors.New("Could not count number of nodes")
	}

	// Perform the query.
	rows, err := db.Query(`
SELECT address,owner,contact,details,pgp,lat,lon,status
FROM nodes;`)
	if err != nil {
		l.Errf("Error dumping database: %s", err)
		return
	}
	defer rows.Close()

	// Now, loop through, initialize the nodes, and fill them out
	// using only the selected columns.
	for i := 0; rows.Next(); i++ {
		// Initialize the node and put it in the table.
		node := new(Node)
		nodes[i] = node

		// Create temporary values to simplify scanning.
		contact := sql.NullString{}
		details := sql.NullString{}

		// Scan all of the values into it.
		err = rows.Scan(&node.Addr, &node.OwnerName,
			&contact, &details, &node.PGP,
			&node.Latitude, &node.Longitude, &node.Status)
		if err != nil {
			l.Errf("Error dumping database: %s", err)
			return
		}

		node.Contact = contact.String
		node.Details = details.String
	}
	return
}

// DumpChanges returns all nodes, both local and cached, which have
// been updated or retrieved more recently than the given time.
func (db DB) DumpChanges(time time.Time) (nodes []*Node, err error) {
	rows, err := db.Query(`
SELECT address,owner,contact,details,pgp,lat,lon,status
FROM nodes WHERE updated >= ?
UNION
SELECT address,owner,"",details,"",lat,lon,status
FROM nodes_cached WHERE retrieved >= ?;`, time, time)
	if err != nil {
		return
	}

	// Append each node to the array in sequence.
	for rows.Next() {
		node := new(Node)

		contact := sql.NullString{}
		details := sql.NullString{}

		err = rows.Scan(&node.Addr, &node.OwnerName,
			&contact, &details, &node.PGP,
			&node.Latitude, &node.Longitude, &node.Status)
		if err != nil {
			return
		}

		node.Contact = contact.String
		node.Details = details.String

		nodes = append(nodes, node)
	}
	return
}

// AddNode inserts a node into the 'nodes' table with the current
// timestamp.
func (db DB) AddNode(node *Node) (err error) {
	// Inserts a new node into the database
	stmt, err := db.Prepare(`INSERT INTO nodes
(address, owner, email, contact, details, pgp, lat, lon, status, updated)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return
	}
	_, err = stmt.Exec([]byte(node.Addr), node.OwnerName, node.OwnerEmail,
		node.Contact, node.Details, []byte(node.PGP),
		node.Latitude, node.Longitude, node.Status,
		time.Now())
	stmt.Close()
	return
}

func (db DB) AddNodes(nodes []*Node) (err error) {
	stmt, err := db.Prepare(`INSERT INTO nodes
(address, owner, email, contact, details, pgp, lat, lon, status, updated)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`)
	if err != nil {
		return
	}

	for _, node := range nodes {
		_, err = stmt.Exec([]byte(node.Addr),
			node.OwnerName, node.OwnerEmail,
			node.Contact, node.Details, []byte(node.PGP),
			node.Latitude, node.Longitude, node.Status,
			time.Now())
		if err != nil {
			return
		}
	}
	stmt.Close()
	return
}

// UpdateNode replaces the node in the database with the IP matching
// the given node.
func (db DB) UpdateNode(node *Node) (err error) {
	// Updates an existing node in the database
	stmt, err := db.Prepare(`UPDATE nodes SET
owner = ?, contact = ?, details = ?, pgp = ?, lat = ?, lon = ?, status = ?
WHERE address = ?`)
	if err != nil {
		return
	}
	_, err = stmt.Exec(node.OwnerName, node.Contact,
		node.Details, []byte(node.PGP),
		node.Latitude, node.Longitude, node.Status, []byte(node.Addr))
	stmt.Close()
	return
}

// DeleteNode removes the node with the matching IP from the 'nodes'
// table in the database.
func (db DB) DeleteNode(addr IP) (err error) {
	// Deletes the given node from the database
	stmt, err := db.Prepare("DELETE FROM nodes WHERE address = ?")
	if err != nil {
		return
	}
	_, err = stmt.Exec([]byte(addr))

	stmt.Close()
	return
}

// GetNode retrieves a single node from the database using the given
// address. If there is a database error, it will be returned. If no
// node matches, however, both return values will be nil.
func (db DB) GetNode(addr IP) (node *Node, err error) {
	// Retrieves the node with the given address from the database
	stmt, err := db.Prepare(`
SELECT owner, email, contact, details, pgp, lat, lon, status
FROM nodes
WHERE address = ?
UNION
SELECT owner, "", "", details, "", lat, lon, status
FROM nodes_cached
WHERE address = ?
LIMIT 1`)
	if err != nil {
		return
	}

	// Initialize the node and temporary variable.
	node = &Node{Addr: addr}
	baddr := []byte(addr)
	contact := sql.NullString{}
	details := sql.NullString{}

	// Perform the actual query.
	row := stmt.QueryRow(baddr, baddr)
	err = row.Scan(&node.OwnerName, &node.OwnerEmail,
		&contact, &details, &node.PGP,
		&node.Latitude, &node.Longitude, &node.Status)
	stmt.Close()

	node.Contact = contact.String
	node.Details = details.String

	// If the error is of the particular type sql.ErrNoRows, it simply
	// means that the node does not exist. In that case, return (nil,
	// nil).
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return
}
