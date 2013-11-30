package main

// Copyright (C) 2013 Alexander Bauer, Luke Evers, Daniel Supernault,
// Dylan Whichard, and contributors; (GPLv3) see LICENSE or doc.go

import (
	"errors"
	"net"
	"net/http"
	"io/ioutil"
	"code.google.com/p/go.crypto/openpgp"
	"code.google.com/p/go.crypto/openpgp/packet"
	"code.google.com/p/go.crypto/openpgp/armor"
	"net/smtp"
	"strings"
	"time"
)

var (
	SMTPDisabledError = errors.New("SMTP disabled in the configuration")
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

// LookupKey uses http to pull a user's pgp key (ASCII armored) from
// a keyserver. In this case it uses keyserver.ubuntu.com but this could
// be replaced. It returns a pointer to an initialized openpgp.Entity.
func LookupKey(pgpsig string) (recipient *openpgp.Entity, err error) {
	resp, err := http.Get(Conf.PGP.Keyserver + pgpsig)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	beginning := strings.Index(string(body), "<pre>")
	end := strings.Index(string(body), "</pre>")

	if beginning < 0 || end < 0 {
		return nil, errors.New("Error: Key not found.")
	}
	
	keyBody, err := armor.Decode(strings.NewReader(string(body[beginning+6:end])))
	if err != nil {
		return nil, err
	}
	return openpgp.ReadEntity(packet.NewReader(keyBody.Body))
}

// PrepareEmail connects to the SMTP server configured in the Conf and
// prepares an email with the given fields. The caller should call the
// Data() method on the *smtp.Client and write the body directly to
// the io.WriteCloser it returns. To send, invoke the Client's Quit()
// method.
func PrepareEmail(from, to string) (c *smtp.Client, err error) {
	// Connect to the SMTP server and authenticate.
	c, err = ConnectSMTP()
	if err != nil {
		return
	}

	// TODO(DuoNoxSol): Check for correct-looking email addresses?

	// Set the sender and recipient.
	if err = c.Mail(from); err != nil {
		c.Quit()
		return nil, err
	}
	if err = c.Rcpt(to); err != nil {
		c.Quit()
		return nil, err
	}

	// Now, return the Client.
	return c, nil
}

// Email simplifies the process of crafting and sending emails via
// SMTP. It makes use of the global *template.Template t.
type Email struct {
	// To, From, and Subject are standard pieces of an Email template.
	To, From, Subject string

	// Data contains any additional, which should be referenced by
	// name in the template, e.g. '{{.Data.FieldName}}'.
	Data map[string]interface{}

	// Header contains data which is generated at Send() time. It does
	// not need to need to be filled out.
	Header struct {
		Date string
	}
}

func (e *Email) Send(templateName string, PGPsig PGPID) (err error) {
	var encrypted bool
	var PGPkey *openpgp.Entity
	if PGPsig.String() != "" {
		PGPkey, err = LookupKey(PGPsig.String())
		if err == nil {
			encrypted = true
		} else {
			encrypted = false
		}
	}
	c, err := PrepareEmail(Conf.SMTP.EmailAddress, e.To)
	if err != nil {
		return
	}
	defer c.Quit()

	e.Header.Date = time.Now().Format(time.RFC1123Z)

	// Tell the server we're about to send it the data.
	w, err := c.Data()
	if err != nil {
		return
	}

	// Execute the template verification.txt and write directly to the
	// connection.
	if encrypted {
		recipients := []*openpgp.Entity{PGPkey}
		armored, err := armor.Encode(w, "PGP MESSAGE", nil)
		if err != nil {
			return err
		}
		plaintext, err := openpgp.Encrypt(armored, recipients, nil, nil, nil)
		err = t.ExecuteTemplate(plaintext, templateName, e)
		if err != nil {
			return err
		}
		plaintext.Close()
		armored.Close()
		return err
	} else {
		return t.ExecuteTemplate(w, templateName, e)
	}
}
