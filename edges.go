package main

import (
	"errors"
	"github.com/inhies/go-cjdns/admin"
	"net"
	"strings"
)

var NetworkAdminNotConnectedError = errors.New("Network admin interface not connected")
var NetworkAdminCredentialsMissingError = errors.New("Network admin credentials missing")
var NetworkAdminCredentialsInvalidError = errors.New("Network admin credentials invalid")

var KnownPeers []Pair

type Pair struct {
	A IP
	B IP
}

type Peers struct {
	Source       IP
	Destinations []IP
}

type Network interface {
	// Connect initializes the object and connects to whatever
	// administration interfaces necessary.
	Connect(*Config) error

	// Close closes any open connections and removes any stored
	// passwords.
	Close() error

	// PeersOf retrieves all IP addresses known to be connected to the
	// given IP. It can return nil.
	PeersOf(IP) (*Peers, error)

	// PeersOfAll functions similarly to PeersOf, but gives connected
	// IPs for all given IPs, in the order they are given. Slices can
	// be nil.
	PeersOfAll([]IP) ([]*Peers, error)
}

// PopulateRoutes finds the peers of every known node in the
// database. It is blocking, and may wait on network IO.
func PopulatePeers(db DB) {
	if Conf.NetworkAdmin == nil {
		l.Infoln("Network admin interface not specified; skipping")
		return
	}

	// Choose which kind of network to connect to.
	var network Network
	switch strings.ToLower(Conf.NetworkAdmin.Type) {
	case "cjdns":
		network = &CJDNSNetwork{}
	}

	// Dump all the nodes in the database.
	nodes, err := db.DumpNodes()
	if err != nil {
		l.Errf("Error listing peers: %s", err)
		return
	}

	// Reduce the list of nodes into just a list of IPs.
	ips := make([]IP, len(nodes))
	for i, node := range nodes {
		ips[i] = node.Addr
	}

	// Connect to the network and find the peers for the whole list of
	// IPs.
	err = network.Connect(Conf)
	if err != nil {
		l.Errf("Error listing peers: %s", err)
		return
	}
	peers, err := network.PeersOfAll(ips)
	if err != nil {
		l.Errf("Error listing peers: %s", err)
		return
	}

	// Allocate a set of pairs with enough capacity to hold the entire
	// list of peers, assuming each one has two connections.
	pairs := make([]Pair, 0, len(peers))

	// Flatten the peer network. Remove duplicates by only adding a
	// connection between nodes if they are already sorted such that
	// the node with the lesser IP is first.
	// TODO(DuoNoxSol): Only discard them in this way if both nodes
	// are in the map
	for _, peer := range peers {
		for _, destinationIP := range peer.Destinations {
			if peer.Source.LessThan(destinationIP) {
				pairs = append(pairs, Pair{
					A: peer.Source,
					B: destinationIP,
				})
			}
		}
	}

	l.Infof("Peering data refreshed")
	KnownPeers = pairs
}

type CJDNSNetwork struct {
	// connected reports whether the Network is currently connected to
	// the admin interface.
	connected bool

	// conn is the connection opened to the CJDNS admin interface. Its
	// methods can be used to access the interface.
	conn *admin.Conn

	// Routes is the slice of all known routes in the currently
	// connected network. It is used in calculating the peers of any
	// given node.
	Routes admin.Routes
}

func (n *CJDNSNetwork) Connect(conf *Config) (err error) {
	// Check to make sure that the credentials can be retrieved. If
	// not, error and exit.
	if conf.NetworkAdmin == nil ||
		conf.NetworkAdmin.Credentials == nil {
		return NetworkAdminCredentialsMissingError
	}

	// Try to cast the credentials to the appropriate type. If this
	// fails, report them invalid.
	credentials, err := makeCJDNSAdminConfig(
		conf.NetworkAdmin.Credentials)
	if err != nil {
		return
	}

	n.conn, err = admin.Connect(credentials)
	if err == nil {
		n.connected = true
	}
	return
}

func (n *CJDNSNetwork) Close() error {
	n.connected = false
	return n.conn.Conn.Close()
}

func (n *CJDNSNetwork) PeersOf(ip IP) (peers *Peers, err error) {
	// First, ensure that the Network is connected. If not, return the
	// appropriate error.
	if !n.connected {
		return nil, NetworkAdminNotConnectedError
	}

	if len(n.Routes) == 0 {
		n.Routes, err = n.conn.NodeStore_dumpTable()
		if err != nil {
			return
		}
	}

	// Find all of the routes from the given IP to its peers. Strip
	// these of extra data, and return just the slice of IPs.
	peerRoutes := n.Routes.Peers(net.IP(ip))
	if err != nil {
		return
	}

	peers = &Peers{
		Source:       ip,
		Destinations: make([]IP, len(peerRoutes)),
	}

	for i, route := range peerRoutes {
		peers.Destinations[i] = IP(*route.IP)
	}
	return
}

func (n *CJDNSNetwork) PeersOfAll(ips []IP) (peers []*Peers, err error) {
	peers = make([]*Peers, len(ips))
	for i, ip := range ips {
		peers[i], err = n.PeersOf(ip)
		if err != nil {
			return nil, err
		}
	}
	return
}

func makeCJDNSAdminConfig(raw map[string]interface{}) (netconf *admin.CjdnsAdminConfig, err error) {
	// Set a default error, to save some keystrokes.
	err = NetworkAdminCredentialsInvalidError
	var ok bool

	netconf = &admin.CjdnsAdminConfig{}
	netconf.Addr, ok = raw["addr"].(string)
	if !ok {
		return
	}
	portf, ok := raw["port"].(float64)
	if !ok {
		return
	}
	netconf.Port = int(portf)
	netconf.Password, ok = raw["password"].(string)
	if !ok {
		return
	}

	// This last one is optional, so don't return an error if it
	// fails.
	netconf.Config, ok = raw["config"].(string)

	return netconf, nil
}
