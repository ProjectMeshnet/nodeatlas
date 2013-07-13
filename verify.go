package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// QueueNode inserts the given node into the verify queue with its
// expiration time set to the current time plus the grace period, its
// emailsent field set by the matching argument, and identified by the
// given ID.
func (db DB) QueueNode(id int64, emailsent bool, grace Duration, node *Node) (err error) {
	_, err = db.Exec(`INSERT INTO nodes_verify_queue
(id, address, owner, email, contact, details, pgp,
lat, lon, status, verifysent, expiration)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, []byte(node.Addr), node.OwnerName, node.OwnerEmail,
		node.Contact, node.Details, []byte(node.PGP),
		node.Latitude, node.Longitude, node.Status,
		emailsent, time.Now().Add(time.Duration(grace)))
	return
}

// DeleteExpiredFromQueue removes expired nodes from the verify queue
// by checking if their expiration stamp is past the current time.
func (db DB) DeleteExpiredFromQueue() (err error) {
	_, err = db.Exec(`DELETE FROM nodes_verify_queue
WHERE expiration <= ?;`, time.Now())
	return
}

// VerifyQueuedNode removes a node (as identified by the id) from the
// queue, performs VerifyRequest checks, and inserts it into the nodes
// table. If it encounters an error, the node remains in the verify
// queue.
func (db DB) VerifyQueuedNode(id int64, r *http.Request) (addr IP, verifyerr error, err error) {
	// Get the node via the id.
	var node = new(Node)
	contact := sql.NullString{}
	details := sql.NullString{}

	err = db.QueryRow(`
SELECT address,owner,email,contact,details,pgp,lat,lon,status
FROM nodes_verify_queue WHERE id = ?;`, id).Scan(
		&node.Addr, &node.OwnerName, &node.OwnerEmail,
		&contact, &details, &node.PGP,
		&node.Latitude, &node.Longitude, &node.Status)
	if err != nil {
		return
	}
	node.Contact = contact.String
	node.Details = details.String

	// Perform VerifyRequest checks.
	verifyerr = VerifyRequest(node, r)
	if verifyerr != nil {
		return
	}

	err = db.AddNode(node)
	if err != nil {
		return
	}

	_, err = db.Exec(`DELETE FROM nodes_verify_queue
WHERE id = ?;`, id)
	if err != nil {
		l.Errf("Could not clear verified node %d: %s", id, err)
	}

	// Add the node to the regular database.
	return node.Addr, nil, nil
}

var (
	NodeAddrNotContainedByNetmaskError = "verify: Node address not within configured netmask: %s"
)

// VerifyRegistrant performs appropriate registration-time checks to
// ensure that a Node is fit to be placed in the verification
// queue. If the given Node is acceptable, then no error will be
// returned.
func VerifyRegistrant(node *Node) error {
	// Ensure that the node's address is contained by the netmask.
	if Conf.Verify.Netmask != nil {
		if !(*net.IPNet)(Conf.Verify.Netmask).Contains(net.IP(node.Addr)) {
			return fmt.Errorf(NodeAddrNotContainedByNetmaskError,
				Conf.Verify.Netmask)
		}
	}
	return nil
}

var (
	RemoteAddressDoesNotMatchError = errors.New(
		"verify: remote address does not match Node address")
)

// VerifyRequest performs appropriate verification checks for a Node
// based on a received http.Request, as follows. Checks are only
// performed if they are enabled in the configuration. If all checks
// are successful, it returns nil.
//
// - Ensure that remote address matches the Node's address.
func VerifyRequest(node *Node, r *http.Request) error {
	// Ensure that r.RemoteAddr matches node.Addr.
	if Conf.Verify.FromNode {
		if !net.IP(node.Addr).Equal(net.ParseIP(r.RemoteAddr)) {
			// If the node address and remote address don't match,
			// then this verify step has failed.
			return RemoteAddressDoesNotMatchError
		}
	}
	return nil
}
