package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

// UpdateMapCache updates the node cache intelligently using
// Conf.ChildMaps. Any unknown map addresses are added to the database
// automatically, and errors are logged.
func UpdateMapCache() {
	// If there are no addresses to retrieve from, do nothing.
	if len(Conf.ChildMaps) == 0 {
		return
	}

	// Because we are refreshing the entire cache, delete all cached
	// nodes.
	err := Db.ClearCache()
	if err != nil {
		l.Errf("Error clearing cache: %s", err)
		return
	}

	// Get a full database dump from all child maps and cache it.
	err = GetAllFromChildMaps(Conf.ChildMaps)
	if err != nil {
		l.Errf("Error updating map cache: %s", err)
	}
}

func (db DB) CacheNode(node *Node, expiry int) (err error) {
	stmt, err := db.Prepare(`INSERT INTO nodes_cached
(address, owner, details, lat, lon, status, expiration)
VALUES(?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return
	}
	_, err = stmt.Exec(node.Addr, node.OwnerName, node.Details,
		node.Latitude, node.Longitude, node.SourceID, node.Status)
	stmt.Close()
	return
}

func (db DB) CacheNodes(nodes []*Node) (err error) {
	stmt, err := db.Prepare(`INSERT INTO nodes_cached
(address, owner, details, lat, lon, status, source, retrieved)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return
	}

	for _, node := range nodes {
		retrieved := node.RetrieveTime
		if retrieved == 0 {
			retrieved = time.Now().Unix()
		}
		_, err = stmt.Exec([]byte(node.Addr), node.OwnerName,
			node.Details,
			node.Latitude, node.Longitude,
			node.Status, node.SourceID, retrieved)
		if err != nil {
			return
		}
	}
	stmt.Close()
	return
}

func (db DB) ClearCache() (err error) {
	_, err = db.Exec(`DELETE FROM nodes_cached;`)
	return err
}

// AddNewMapSource inserts a new map address into the cached_maps
// table.
func (db DB) AddNewMapSource(address, name string) (err error) {
	_, err = db.Exec(`INSERT INTO cached_maps
(hostname,name) VALUES(?, ?)`, address, name)
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

// nodeDumpWrapper is a structure which wraps a response from /api/all
// in which the Data field is a map[string][]*Node.
type nodeDumpWrapper struct {
	Data  map[string][]*Node `json:"data"`
	Error interface{}        `json:"error"`
}

// GetAllFromChildMaps accepts a list of child map addresses to
// retrieve nodes from. It does this concurrently, and puts any nodes
// and newly discovered addresses in the local ID table.
func GetAllFromChildMaps(addresses []string) (err error) {
	// First off, initialize the slice into which we'll be appending
	// all the nodes, and the souceToID map and mutex.
	nodes := make([]*Node, 0)

	sourceToID, err := Db.GetMapSourceToID()
	if err != nil {
		return
	}
	sourceMutex := new(sync.RWMutex)

	// Next, we'll need a WaitGroup so we can block until all requests
	// complete and a mutex to control appending to nodes.
	waiter := new(sync.WaitGroup)
	nodesMutex := new(sync.Mutex)

	// We'll need to wait for len(addresses) goroutines to finish, so
	// put that number in the WaitGroup.
	waiter.Add(len(addresses))

	// Now, start a separate goroutine for every address to
	// concurrently retrieve nodes and append them (thread-safely) to
	// nodes. Whenever appendNodesFromChildMap() finishes, it calls
	// waiter.Done().
	for _, address := range addresses {
		go appendNodesFromChildMap(&nodes, address,
			&sourceToID, sourceMutex, nodesMutex, waiter)
	}

	// Block until all goroutines are finished. This is simple to do
	// with the WaitGroup, which keeps track of the number we're
	// waiting for.
	waiter.Wait()

	return Db.CacheNodes(nodes)
}

// appendNodesFromChildMap is a helper function used by
// GetAllFromChildMaps() which calls GetAllFromChildMap() and
// thread-safely appends the result to the given slice. At the end of
// the function, it calls wg.Done().
func appendNodesFromChildMap(dst *[]*Node, address string,
	sourceToID *map[string]int, sourceMutex *sync.RWMutex,
	dstMutex *sync.Mutex, wg *sync.WaitGroup) {

	// First, retrieve the nodes if possible. If there was an error,
	// it will be logged, and if there were no nodes, we can stop
	// here.
	nodes := GetAllFromChildMap(address, sourceToID, sourceMutex)
	if nodes == nil {
		wg.Done()
		return
	}

	// Now that we have the nodes, we need to lock the destination
	// slice while we append to it.
	dstMutex.Lock()
	*dst = append(*dst, nodes...)
	dstMutex.Unlock()
	wg.Done()
}

// GetAllFromChildMap retrieves a list of nodes from a single remote
// address, and localizes them. If it encounters a remote address that
// is not already known, it safely adds it to the sourceToID map. It
// is safe for concurrent use. If it encounters an error, it will log
// it and return nil.
func GetAllFromChildMap(address string, sourceToID *map[string]int,
	sourceMutex *sync.RWMutex) (nodes []*Node) {
	// Try to get all nodes via the API.
	resp, err := http.Get(strings.TrimRight(address, "/") + "/api/all")
	if err != nil {
		l.Errf("Caching %q produced: %s", address, err)
		return nil
	}

	// Read the data into a the nodeDumpWrapper type, so that it
	// decodes properly.
	var jresp nodeDumpWrapper
	err = json.NewDecoder(resp.Body).Decode(&jresp)
	if err != nil {
		l.Errf("Caching %q produced: %s", address, err)
		return nil
	} else if jresp.Error != nil {
		l.Errf("Caching %q produced remote error: %s",
			address, jresp.Error)
		return nil
	}

	// Prepare an initial slice so that it can be appended to, then
	// loop through and convert sources to IDs.
	//
	// Additionally, use a boolean to keep track of whether we've
	// replaced "local" with the actual address already, to save some
	// needless compares.
	nodes = make([]*Node, 0)
	var replacedLocal bool
	for source, remoteNodes := range jresp.Data {
		// If we come across "local", then replace it with the address
		// we're retrieving from.
		if !replacedLocal && source == "local" {
			source = address
		}

		// First, check if the source is known. If not, then we need
		// to add it and refresh our map. Make sure all reads and
		// writes to sourceToID are threadsafe.
		sourceMutex.RLock()
		id, ok := (*sourceToID)[source]
		sourceMutex.RUnlock()
		if !ok {
			// Add the new source to the database, and put it in the
			// map under the ID len(sourceToID), because that should
			// be unique.
			sourceMutex.Lock()
			err := Db.AddNewMapSource(source, "")
			if err != nil {
				// Uh oh.
				sourceMutex.Unlock()
				l.Errf("Error while caching %q: %s", address, err)
				return
			}

			id = len(*sourceToID)
			(*sourceToID)[source] = id
			sourceMutex.Unlock()

			l.Debugf("Discovered new source map %q, ID %d\n",
				source, id)
		}

		// Once the ID is set, proceed on to add it in all the
		// remoteNodes.
		for _, n := range remoteNodes {
			n.SourceID = id
		}

		// Finally, append remoteNodes to the slice we're returning.
		nodes = append(nodes, remoteNodes...)
	}
	return
}
