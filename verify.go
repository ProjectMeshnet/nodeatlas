package main

import (
	"errors"
	"net/smtp"
)

var SMTPDisabledError = errors.New("SMTP not specified in the configuration")

// ConnectSMTP uses the global Conf to connect to the SMTP server and,
// unless disabled in the configuration, authenticates with STARTTLS
// if possible. Unless an error is returned, t is the caller's
// responsibility to close the client. If the Conf.SMTP is nil,
// STMPDisabledError will be returned.
func ConnectSMTP() (c *smtp.Client, err error) {
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
func SendVerificationEmail(id string, n *Node) (err error) {
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
