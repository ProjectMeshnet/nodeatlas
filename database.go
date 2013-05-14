package main

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"database/sql"
)

var (
	DB *sql.DB
)
