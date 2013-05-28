package main

import (
	"encoding/json"
	"errors"
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
	ip = IP(net.ParseIP(string(b)))
	if ip == nil {
		return IncorrectlyFormattedIP
	}
	return nil
}
