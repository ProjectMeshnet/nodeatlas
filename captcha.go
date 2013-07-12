package main

import (
	// "github.com/dchest/captcha"
	"database/sql"
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
WHERE id = ? AND expiration >= ?;`, bid, time.Now())
	err := row.Scan(&digits)
	if err == sql.ErrNoRows {
		// If there are no rows, then the ID was not found.
		l.Debugln("CAPTCHA id not found")
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

// ClearExpired removes any expired CAPTCHA solutions from the
// database. It logs errors.
func (*CAPTCHAStore) ClearExpired() {
	_, err := Db.Exec(`DELETE FROM captcha
WHERE expiration < ?;`, time.Now())
	if err != nil {
		l.Err("Error deleting expired CAPTCHAs:", err)
	}
}
