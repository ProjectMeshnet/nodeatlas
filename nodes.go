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
	StatusActive = iota
	StatusInactive
	StatusPlanned
	StatusPossible
	StatusUnknown // All numbers greater than this are unknown.
)

// Node is the basic type for a computer, radio, transmitter, or any
// other sort of node in a mesh network.
type Node struct {
	/* Required Fields */

	// Status represents the indended availability for the node. For
	// example, StatusActive, StatusPlanned, etc.
	Status int

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

// IP is a wrapper for net.IP which implements the json.Marshaler and
// json.Unmarshaler.
type IP net.IP

var IncorrectlyFormattedIP = errors.New("incorrectly formatted ip address")

func (ip IP) MarshalJSON() ([]byte, error) {
	return json.Marshal(net.IP(ip).String())
}

func (ip IP) UnmarshalJSON(b []byte) error {
	// TODO(DuoNoxSol): This function is untested and may not work. If
	// b retains its quotes, then net.ParseIP() may fail.
	ip = IP(net.ParseIP(string(b)))
	if ip == nil {
		return IncorrectlyFormattedIP
	}
	return nil
}

func (ip IP) String() string {
	return net.IP(ip).String()
}

// MultiPointNodes returns a geojson.MultiPoint type from the
// coordinates of the given nodes, in order.
func MultiPointNodes(nodes []*Node) *geojson.FeatureCollection {
	//
	features := make([]*geojson.Feature, len(nodes))
	for i, n := range nodes {
		properties := make(map[string]interface{}, 1)
		properties["popupContent"] = n.Addr
		features[i] = geojson.NewFeature(
			geojson.NewPoint(geojson.Coordinate{
				geojson.CoordType(n.Longitude),
				geojson.CoordType(n.Latitude)}),
			properties,
			n.Addr)
	}
	return geojson.NewFeatureCollection(features)
}
