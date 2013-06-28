package main

import (
	"encoding/json"
	"errors"
	"github.com/kpawlik/geojson"
	"net"
)

// Statuses are the intended states of nodes. For example, if a node
// is intended to be always-online and therefore should usually be
// available, its status would be "active."
const (
	StatusActive = uint32(1 << iota) // << 0 active/planned
	_
	_
	_
	_ // << 4
	_
	_
	StatusPhysical //      physical server/virtual server
	StatusInternet // << 8 internet access/no internet
	StatusWireless //      wireless access/no wireless
	StatusWired    //      wired access/no wired
	_
	_ // << 12
	_
	_
	_
	_ // << 16
	_
	_
	_
	_ // << 20
	_
	_
	_
	StatusPingable // << 24 pingable/down
	_
	_
	_
	_ // << 28
	_
	_
	_
)

// Node is the basic type for a computer, radio, transmitter, or any
// other sort of node in a mesh network.
type Node struct {
	/* Required Fields */
	// SourceID is the local ID of the source map of this node. If the
	// SourceID is 0, the node is considered to be local.
	SourceID int `json:"-"`

	// Status is a list of bit flags representing the node's status,
	// such as whether it is active or planned, has wireless access,
	// is a physical server, etc.
	Status uint32

	// Latitude and Longitude represent the physical location of the
	// node on earth.
	Latitude, Longitude float64

	// Addr is the network address (for the meshnet protocol) of the
	// node.
	Addr IP

	// RetrieveTime is only used if the node is cached, and comes from
	// another map. It is the Unix time (in seconds) at which the node
	// was retrieved from its home instance. If it is zero, the node
	// is not cached.
	RetrieveTime int64 `json:",omitempty"`

	/* Optional Fields */

	// OwnerName is the node's owner's real or screen name.
	OwnerName string

	// OwnerEmail is the node's owner's email address.
	OwnerEmail string `json:",omitempty"`
}

// Feature returns the Node as a *geojson.Feature.
func (n *Node) Feature() (f *geojson.Feature) {
	// Set the properties.
	properties := make(map[string]interface{}, 2)
	properties["OwnerName"] = n.OwnerName
	properties["Status"] = n.Status

	// Create and return the feature.
	return geojson.NewFeature(
		geojson.NewPoint(geojson.Coordinate{
			geojson.CoordType(n.Longitude),
			geojson.CoordType(n.Latitude)}),
		properties,
		n.Addr)
}

// IP is a wrapper for net.IP which implements the json.Marshaler and
// json.Unmarshaler.
type IP net.IP

var IncorrectlyFormattedIP = errors.New("incorrectly formatted ip address")

func (ip IP) MarshalJSON() ([]byte, error) {
	return json.Marshal(net.IP(ip).String())
}

func (ip *IP) UnmarshalJSON(b []byte) error {
	if b[0] != '"' {
		// If a quote is not the first character, the next bit will
		// segfault, so we should return an error.
		return IncorrectlyFormattedIP
	}
	tip := net.ParseIP(string(b[1 : len(b)-1]))
	if tip == nil {
		return IncorrectlyFormattedIP
	}
	// Don't think too hard about this part.
	*ip = *(*IP)(&tip)
	return nil
}

func (ip IP) String() string {
	return net.IP(ip).String()
}

// FeatureCollectionNodes returns a *geojson.FeatureCollection type
// from the given nodes, in order.
func FeatureCollectionNodes(nodes []*Node) *geojson.FeatureCollection {
	features := make([]*geojson.Feature, len(nodes))
	for i, n := range nodes {

		//		properties["Address"] = n.Addr
		features[i] = n.Feature()
	}
	return geojson.NewFeatureCollection(features)
}
