package main

import (
	"database/sql"
	"time"
)

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

// GetMapSourceToID returns a mapping of child map hostnames to their
// local IDs. It also includes a mapping of "local" to id 0.
func (db DB) GetMapSourceToID() (sourceToID map[string]int, err error) {
	// Initialize the map and insert the "local" id.
	sourceToID = make(map[string]int, 1)
	sourceToID["local"] = 0

	// Retrieve every pair of hostnames and IDs.
	rows, err := db.Query(`SELECT hostname,id
FROM cached_maps;`)
	if err == sql.ErrNoRows {
		return sourceToID, nil
	} else if err != nil {
		return
	}

	// Put in the rest of the mappings.
	for rows.Next() {
		var hostname string
		var id int
		if err = rows.Scan(&hostname, &id); err != nil {
			return
		}
		sourceToID[hostname] = id
	}

	return
}

// GetMapIDToSource returns a mapping of local IDs to public
// hostnames. ID 0 is "local".
func (db DB) GetMapIDToSource() (IDToSource map[int]string, err error) {
	// Initialize the slice with "local".
	IDToSource = make(map[int]string, 1)
	IDToSource[0] = "local"

	// Retrieve every pair of IDs and hostnames.
	rows, err := db.Query(`SELECT id,hostname
FROM cached_maps;`)
	if err == sql.ErrNoRows {
		return IDToSource, nil
	} else if err != nil {
		return
	}

	// Put in the rest of the IDs.
	for rows.Next() {
		var id int
		var hostname string
		if err = rows.Scan(&id, &hostname); err != nil {
			return
		}
		IDToSource[id] = hostname
	}
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
	// First, get a mapping of IDs to sources for quick access.
	idSources, err := db.GetMapIDToSource()
	if err != nil {
		return
	}

	// Now, prepare the data to be returned. Nodes will be added one
	// at a time to the key arrays.
	sourceMaps = make(map[string][]*Node)
	for _, node := range nodes {
		hostname := idSources[node.SourceID]
		sourcemapNodes := sourceMaps[hostname]
		if sourcemapNodes == nil {
			sourcemapNodes = make([]*Node, 0, 5)
		}

		sourceMaps[hostname] = append(sourcemapNodes, node)
	}
	return
}
