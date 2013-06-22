package main

import (
	"errors"
	"net/smtp"
	"time"
)

var SMTPDisabledError = errors.New("SMTP disabled in the configuration")

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

		// Authenticate using the password via CRAM-MD5. (Wait, we still
		// use MD5? Ugh.)
		if err = c.Auth(smtp.CRAMMD5Auth(Conf.SMTP.Username,
			Conf.SMTP.Password)); err != nil {
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
	data["Link"] = "http://" + Conf.Hostname + Conf.Prefix
	data["VerificationID"] = id

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
(id, address, owner, email, lat, lon, status, verifysent, expiration)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, []byte(node.Addr), node.OwnerName, node.OwnerEmail,
		node.Latitude, node.Longitude, node.Status,
		emailsent, time.Now().Add(time.Duration(grace)))
	return
}

// DeleteExpiredFromQueue removes expired nodes from the verify queue
// by checking if their expiration stamp is past the current time.
func (db DB) DeleteExpiredFromQueue() (err error) {
	_, err = db.Exec(`DELETE FROM nodes_verify_queue
WHERE expiration <= DATETIME('now');`)
	return
}

// VerifyQueuedNode removes a node (as identified by the id) from the
// queue and inserts it into the nodes table.
func (db DB) VerifyQueuedNode(id int64) (addr IP, err error) {
	// Get the node via the id.
	var node = new(Node)
	err = db.QueryRow(`SELECT address,owner,email,lat,lon,status
FROM nodes_verify_queue WHERE id = ?;`, id).Scan(
		&node.Addr, &node.OwnerName, &node.OwnerEmail,
		&node.Latitude, &node.Longitude, &node.Status)
	if err != nil {
		return
	}
	_, err = db.Exec(`DELETE FROM nodes_verify_queue
WHERE id = ?;`, id)
	if err != nil {
		l.Errf("Could not clear verified node %d: %s", id, err)
	}

	// Add the node to the regular database.
	return node.Addr, db.AddNode(node)
}
