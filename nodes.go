package main

// Copyright (C) 2013 Alexander Bauer, Luke Evers, Daniel Supernault,
// Dylan Whichard, and contributors; (GPLv3) see LICENSE or doc.go

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/baliw/moverss"
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

	// Contact is public contact information, such as a nickname,
	// email, or xmpp username.
	Contact string `json:",omitempty"`

	// Details is extra information about node
	Details string `json:",omitempty"`

	// PGP is the key ID of the owner's public key.
	PGP PGPID `json:",omitempty"`
}

// Feature returns the Node as a *geojson.Feature.
func (n *Node) Feature() (f *geojson.Feature) {
	// Set the properties.
	properties := make(map[string]interface{}, 6)
	properties["OwnerName"] = n.OwnerName
	properties["Status"] = n.Status

	if len(n.Contact) != 0 {
		properties["Contact"] = n.Contact
	}
	if len(n.PGP) != 0 {
		properties["PGP"] = n.PGP.String()
	}
	if len(n.Details) != 0 {
		properties["Details"] = n.Details
	}
	if n.SourceID != 0 {
		properties["SourceID"] = n.SourceID
	}

	// Create and return the feature.
	return geojson.NewFeature(
		geojson.NewPoint(geojson.Coordinate{
			geojson.CoordType(n.Longitude),
			geojson.CoordType(n.Latitude)}),
		properties,
		n.Addr)
}

// Item returns the Node as a *moverss.Item. It does not set the
// timestamp.
func (n *Node) Item() (i *moverss.Item) {
	return &moverss.Item{
		Link:    Conf.Web.Hostname + "/node/" + n.Addr.String(),
		Title:   n.OwnerName,
		XMLName: NodeXMLName,
	}
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
		features[i] = n.Feature()
	}
	return geojson.NewFeatureCollection(features)
}

type PGPID []byte

var IncorrectlyFormattedPGPID = errors.New("incorrectly formatted PGP ID")

func (pgpid PGPID) MarshalJSON() ([]byte, error) {
	b := make([]byte, len(pgpid)*2)
	hex.Encode(b, pgpid)
	return json.Marshal(string(b))
}

func (pgpid *PGPID) UnmarshalJSON(b []byte) error {
	if b[0] != '"' {
		// If a quote is not the first character, then the next part
		// will segfault.
		return IncorrectlyFormattedPGPID
	}
	b = b[1 : len(b)-1]
	if len(b) != 0 && len(b) != 8 && len(b) != 16 {
		return IncorrectlyFormattedPGPID
	} else if len(b) == 0 {
		// If the length is zero, make the result nil.
		*pgpid = nil
	}
	*pgpid = make(PGPID, len(b)/2)
	_, err := hex.Decode(*pgpid, b)
	return err
}

func (pgpid PGPID) String() string {
	if len(pgpid) == 0 {
		return ""
	}

	b := make([]byte, len(pgpid)*2)
	_ = hex.Encode(b, pgpid)
	return string(b)
}

func DecodePGPID(b []byte) (pgpid PGPID, err error) {
	if len(b) != 0 && len(b) != 8 && len(b) != 16 {
		return nil, IncorrectlyFormattedPGPID
	}
	pgpid = make(PGPID, len(b)/2)
	_, err = hex.Decode(pgpid, b)
	return
}
