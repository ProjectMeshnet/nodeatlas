package main

// Copyright (C) 2013 Alexander Bauer, Luke Evers, Daniel Supernault,
// Dylan Whichard, and contributors; (GPLv3) see LICENSE or doc.go

import (
	"github.com/dchest/captcha"
	"database/sql"
	"net/http"
	"strings"
	"time"
)

const (
	CAPTCHAGracePeriod = time.Minute * 10
)

type CAPTCHAStore struct{}

// Set inserts a new CAPTCHA ID and solution into the captcha table in
// the database. It logs errors.
func (CAPTCHAStore) Set(id string, digits []byte) {
	_, err := Db.Exec(`INSERT INTO captcha
(id, solution, expiration)
VALUES(?, ?, ?);`,
		[]byte(id), digits, time.Now().Add(CAPTCHAGracePeriod))
	if err != nil {
		l.Err("Error registering CAPTCHA:", err)
	}
}

// Get retrieves a CAPTCHA solution from the database and clears the
// row if appropriate. It logs errors.
func (CAPTCHAStore) Get(id string, clear bool) (digits []byte) {
	bid := []byte(id)
	row := Db.QueryRow(`SELECT solution
FROM captcha
WHERE id = ? AND expiration > ?;`, bid, time.Now())
	err := row.Scan(&digits)
	if err == sql.ErrNoRows {
		// If there are no rows, then the ID was not found.
		return nil
	} else if err != nil {
		l.Err("Error retrieving CAPTCHA:", err)
		return nil
	}

	// If we're supposed to remove the CAPTCHA from the database, then
	// do so.
	if clear {
		_, err = Db.Exec(`DELETE FROM captcha
WHERE id = ?;`, bid)
		if err != nil {
			l.Err("Error deleting CAPTCHA:", err)
		}
	}
	return
}

// ClearExpiredCAPTCHA removes any expired CAPTCHA solutions from the
// database. It logs errors.
func ClearExpiredCAPTCHA() {
	_, err := Db.Exec(`DELETE FROM captcha
WHERE expiration <= ?;`, time.Now())
	if err != nil {
		l.Err("Error deleting expired CAPTCHAs:", err)
	}
}


// VerifyCAPTCHA accepts a *http.Request and verifies that the given
// 'captcha' form is valid. This is a string of the form
// "id:solution". It will return IncorrectCAPTCHAError if the solution
// or ID is invalid.
func VerifyCAPTCHA(req *http.Request) error {
	// Get the "captcha" form value.
	solution := req.FormValue("captcha")
	
	// Find the point to split the form value at. If it's not found in
	// the string, return the InvalidCAPTCHAFormat error.
	index := strings.Index(solution, ":")
	if index < 0 {
		return InvalidCAPTCHAFormat
	}
	
	// If that was successful, try to verify it. If it returns false,
	// the ID or solution was invalid.
	if !captcha.VerifyString(solution[:index], solution[index+1:]) {
		return IncorrectCAPTCHA
	}
	
	// If we get to this point, then it was successfully validated and
	// we can return nil.
	return nil
}
