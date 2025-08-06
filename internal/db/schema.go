// Package db provides database schema initialization for the Leo DNS blocker.
package db

import (
	"database/sql"
	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

func Init(db *sql.DB) error {
	// Enable foreign keys
	_, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return err
	}

	// Execute the embedded SQL script
	_, err = db.Exec(schemaSQL)
	if err != nil {
		return err
	}

	return nil
}
