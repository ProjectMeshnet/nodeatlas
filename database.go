package main

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"database/sql"
	"net"
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
	_, err = db.Query(`CREATE TABLE IF NOT EXISTS nodes_cached (
address BINARY(16) PRIMARY KEY,
owner VARCHAR(255) NOT NULL,
lat FLOAT NOT NULL,
lon FLOAT NOT NULL,
status INT NOT NULL,
source VARCHAR(255) NOT NULL,
retrieved DATETIME DEFAULT CURRENT_TIMESTAMP,
expiration DATETIME);`)
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
		row = db.QueryRow("SELECT COUNT(*) FROM (SELECT address FROM nodes UNION SELECT address FROM nodes_cached);")
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

// AddNode
func (db DB) AddNode(node *Node) (err error) {
	// Inserts a new node into the database
	stmt, err := db.Prepare(`INSERT INTO nodes
(address, owner, lat, lon, status)
VALUES(?, ?, ?, ?, ?)`)
	if err != nil {
		return
	}
	_, err = stmt.Exec([]byte(node.Addr), node.OwnerName, node.Latitude, node.Longitude, node.Status)
	stmt.Close()
	return
}

func (db DB) UpdateNode(node *Node) (err error) {
	// Updates an existing node in the database
	stmt, err := db.Prepare(`UPDATE nodes SET 
owner = ?, lat = ?, lon = ?, status = ?
WHERE address = ?`)
	if err != nil {
		return
	}
	_, err = stmt.Exec(node.OwnerName, node.Latitude, node.Longitude, node.Status, []byte(node.Addr))
	stmt.Close()
	return
}

func (db DB) DeleteNode(node *Node) (err error) {
	// Deletes the given node from the database
	stmt, err := db.Prepare("DELETE FROM nodes WHERE address = ?")
	if err != nil {
		return
	}
	_, err = stmt.Exec([]byte(node.Addr))

	stmt.Close()
	return
}

func (db DB) GetNode(addr *net.IP) (node *Node, err error) {
	// Retrieves the node with the given address from the database
	stmt, err := db.Prepare(`SELECT address, owner, lat, lon, status
FROM nodes
WHERE address = ?
UNION
SELECT address, owner, lat, lon, status
FROM nodes_cached
WHERE address = ?
LIMIT 1`)
	if err != nil {
		node = nil
		return
	}
	row := stmt.QueryRow(addr, addr)
	err = row.Scan(&node.Addr, &node.OwnerName, &node.Latitude, &node.Longitude, &node.Status)
	stmt.Close()
	return
}

func (db DB) CacheNode(node *Node, source string, expiry int) (err error) {
	stmt, err := db.Prepare(`INSERT INTO nodes_cached
(address, owner, lat, lon, status, source, expiration)
VALUES(?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return
	}
	_, err = stmt.Exec(node.Addr, node.OwnerName, node.Latitude, node.Longitude, node.Status, source, expiry)
	stmt.Close()
	return
}
