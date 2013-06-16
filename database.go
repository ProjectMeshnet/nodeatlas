package main

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
	ReadOnly bool
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
email VARCHAR(255) NOT NULL,
lat FLOAT NOT NULL,
lon FLOAT NOT NULL,
status INT NOT NULL,
source INT NOT NULL,
retrieved DATETIME DEFAULT CURRENT_TIMESTAMP,
expiration DATETIME);`)
	if err != nil {
		return
	}
	_, err = db.Query(`CREATE TABLE IF NOT EXISTS nodes_verify_queue (
id INT PRIMARY KEY,
address BINARY(16) NOT NULL,
owner VARCHAR(255) NOT NULL,
email VARCHAR(255) NOT NULL,
lat FLOAT NOT NULL,
lon FLOAT NOT NULL,
status INT NOT NULL,
expiration DATETIME);`)
	if err != nil {
		return
	}

	_, err = db.Query(`CREATE TABLE IF NOT EXISTS cached_maps (
id INTEGER PRIMARY KEY AUTOINCREMENT,
hostname VARCHAR(255) NOT NULL,
name VARCHAR(255) NOT NULL);`)
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
		return nil, errors.New("Could not count number of nodes")
	}

	// Perform the query.
	rows, err := db.Query(`
SELECT address,owner,lat,lon,status,0
FROM nodes
UNION SELECT address,owner,lat,lon,status,retrieved
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

		// Scan all of the values into it.
		err = rows.Scan(&node.Addr, &node.OwnerName, &node.Latitude,
			&node.Longitude, &node.Status, &node.RetrieveTime)
		if err != nil {
			return
		}
	}
	return
}

// AddNode inserts a node into the 'nodes' table with the current
// timestamp.
func (db DB) AddNode(node *Node) (err error) {
	// Inserts a new node into the database
	stmt, err := db.Prepare(`INSERT INTO nodes
(address, owner, email, lat, lon, status)
VALUES(?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return
	}
	_, err = stmt.Exec([]byte(node.Addr), node.OwnerName, node.OwnerEmail, node.Latitude, node.Longitude, node.Status)
	stmt.Close()
	return
}

// UpdateNode replaces the node in the database with the IP matching
// the given node.
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

// DeleteNode removes the node with the matching IP from the database.
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
	stmt, err := db.Prepare(`SELECT owner, email, lat, lon, status
FROM nodes
WHERE address = ?
UNION
SELECT owner, email, lat, lon, status
FROM nodes_cached
WHERE address = ?
LIMIT 1`)
	if err != nil {
		return
	}

	// Initialize the node and temporary variable.
	node = &Node{Addr: addr}
	baddr := []byte(addr)

	// Perform the actual query.
	row := stmt.QueryRow(baddr, baddr)
	err = row.Scan(&node.OwnerName, &node.OwnerEmail, &node.Latitude, &node.Longitude, &node.Status)
	stmt.Close()

	// If the error is of the particular type sql.ErrNoRows, it simply
	// means that the node does not exist. In that case, return (nil,
	// nil).
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return
}

func (db DB) CacheNode(node *Node, source int, expiry int) (err error) {
	stmt, err := db.Prepare(`INSERT INTO nodes_cached
(address, owner, email, lat, lon, status, source, expiration)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return
	}
	_, err = stmt.Exec(node.Addr, node.OwnerName, node.OwnerEmail, node.Latitude, node.Longitude, node.Status)
	stmt.Close()
	return
}

func (db DB) CacheNodes(nodes []*Node, source string) (err error) {
	stmt, err := db.Prepare(`INSERT INTO nodes_cached
(address, owner, email, lat, lon, status, source, retrieved)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return
	}

	for _, node := range nodes {
		retrieved := node.RetrieveTime
		if retrieved == 0 {
			retrieved = time.Now().Unix()
		}
		_, err = stmt.Exec([]byte(node.Addr), node.OwnerName, node.OwnerEmail, node.Latitude, node.Longitude, node.Status, source, retrieved)
		if err != nil {
			return
		}
	}
	stmt.Close()
	return
}

func (db DB) FindSourceMap(id int) (source string, err error) {
	if id == 0 {
		return "local", nil
	}
	row := db.QueryRow(`SELECT hostname
FROM cached_maps
WHERE id=?`, id)

	err = row.Scan(&source)
	return
}

func (db DB) CacheFormatNodes(nodes []*Node) (sourceMaps map[string][]*Node, err error) {
	sourceMaps = make(map[string][]*Node)
	for _, node := range nodes {
		hostname, err := db.FindSourceMap(node.SourceID)
		if err != nil {
			return nil, err
		}
		sourcemapNodes := sourceMaps[hostname]
		if sourcemapNodes == nil {
			sourcemapNodes = make([]*Node, 0, 5)
		}

		sourceMaps[hostname] = append(sourcemapNodes, node)
	}
	return
}

func (db DB) QueueNode(id int64, expire Duration, node *Node) (err error) {
	// Insert the node into the table with the expiration time set to
	// the current time plus the grace period.
	_, err = db.Exec(`INSERT INTO nodes_verify_queue
(id, address, owner, email, lat, lon, status, expiration)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		id, []byte(node.Addr), node.OwnerName, node.OwnerEmail,
		node.Latitude, node.Longitude, node.Status,
		time.Now().Add(time.Duration(expire)))
	return
}

func (db DB) DeleteQueuedNode(id int64) (err error) {
	_, err = db.Exec(`DELETE FROM nodes_verify_queue WHERE id = ?`, id)
	return
}

// VerifyQueuedNode removes a node (as identified by the id) from the
// queue and inserts it into the nodes table.
func (db DB) VerifyQueuedNode(id int64) (addr IP, err error) {
	// Get the node via the id.
	var node = new(Node)
	err = db.QueryRow(`SELECT address,owner,email,lat,lon,status
FROM nodes_verify_queue WHERE id = ?`, id).Scan(
		&node.Addr, &node.OwnerName, &node.OwnerEmail,
		&node.Latitude, &node.Longitude, &node.Status)
	if err != nil {
		return
	}

	// Add the node to the regular database.
	return node.Addr, db.AddNode(node)
}
