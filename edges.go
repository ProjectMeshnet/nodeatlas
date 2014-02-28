package main

import (
	"errors"
)

var NetworkAdminNotConnectedError = errors.New("Network admin interface not connected.")

type Network interface {
	// Connect initializes the object and connects to whatever
	// administration interfaces necessary.
	Connect() error

	// PeersOf retrieves all IP addresses known to be connected to the
	// given IP. It can return a nil slice.
	PeersOf(IP) ([]IP, error)

	// PeersOfAll functions similarly to PeersOf, but gives connected
	// IPs for all given IPs, in the order they are given. Slices can
	// be nil.
	PeersOfAll([]IP) ([][]IP, error)
}
