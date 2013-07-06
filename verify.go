package main

import (
	"database/sql"
	"errors"
	"net"
	"net/http"
	"net/smtp"
	"time"
)

var (
	SMTPDisabledError = errors.New("SMTP disabled in the configuration")

	RemoteAddressDoesNotMatchError = errors.New(
		"verify: remote address does not match Node address")
)

// ConnectSMTP uses the global Conf to connect to the SMTP server and,
// unless disabled in the configuration, authenticates with STARTTLS
// if possible. Unless an error is returned, t is the caller's
// responsibility to close the client. If the Conf.SMTP is nil,
// STMPDisabledError will be returned.
func ConnectSMTP() (c *smtp.Client, err error) {
	// If the SMTP portion of the config isn't specified, we don't
	// know what to connect to.
	if Conf.SMTP == nil {
		return nil, SMTPDisabledError
	}

	// Connect to the server.
	c, err = smtp.Dial(Conf.SMTP.ServerAddress)
	if err != nil {
		return
	}

	// If NoAuthenticate is true, then skip the authentication step.
	if !Conf.SMTP.NoAuthenticate {
		// Upgrade to STARTTLS if supported.
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err = c.StartTLS(nil); err != nil {
				c.Quit()
				return
			}
		}

		// Authenticate using the password via plain auth.
		host, _, err := net.SplitHostPort(Conf.SMTP.ServerAddress)
		if err != nil {
			return nil, err
		}

		if err = c.Auth(smtp.PlainAuth("", Conf.SMTP.Username,
			Conf.SMTP.Password, host)); err != nil {
			// If the authentication fails, close the client and exit.
			c.Quit()
			return nil, err
		}
	}

	// If all is successful, return.
	return
}

// SendVerificationEmail uses the fields in Conf.SMTP to send a
// templated email (verification.txt) to the email address specified
// by the given node. If the email could not be sent, it returns an
// error.
func SendVerificationEmail(id int64, n *Node) (err error) {
	// Connect to the SMTP server and authenticate.
	c, err := ConnectSMTP()
	if err != nil {
		return
	}
	defer c.Quit()

	// TODO(DuoNoxSol): Check for correct-looking email addresses?

	// Set the sender and recipient.
	if err = c.Mail(Conf.SMTP.EmailAddress); err != nil {
		return
	}
	if err = c.Rcpt(n.OwnerEmail); err != nil {
		return
	}

	// Prepare the template data.
	data := make(map[string]interface{}, 3)
	data["InstanceName"] = Conf.Name
	data["Link"] = "http://" + Conf.Web.Hostname + Conf.Web.Prefix
	data["VerificationID"] = id
	data["From"] = Conf.SMTP.EmailAddress
	data["To"] = n.OwnerEmail
	data["Date"] = time.Now().Format(time.RFC1123Z)

	// Tell the server we're about to send it the data.
	w, err := c.Data()
	if err != nil {
		return
	}

	// Execute the template verification.txt and write directly to the
	// connection.
	return t.ExecuteTemplate(w, "verification.txt", data)
}

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

// VerifyRequest performs appropriate verification checks for a Node
// based on a received http.Request, as follows. Checks are only
// performed if they are enabled in the configuration. If all checks
// are successful, it returns nil.
//
// - Ensure that remote address matches the Node's address.
func VerifyRequest(node *Node, r *http.Request) error {
	// Ensure that r.RemoteAddr matches node.Addr.
	if Conf.Verify.FromNode {
		remoteAddr, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// If we encounter an error here, it is probably something
			// to do with the reverse proxy. As such, log it as an
			// internal error, and return it.
			l.Errf("Could not parse input address while verifying %q: %s\n",
				node.Addr, err)
			return err
		}
		if !net.IP(node.Addr).Equal(net.ParseIP(remoteAddr)) {
			// If the node address and remote address don't match,
			// then this verify step has failed.
			return RemoteAddressDoesNotMatchError
		}
	}
	return nil
}
