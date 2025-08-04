// Package db provides database schema initialization for the Leo DNS blocker.
package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func Init(db *sql.DB) error {
	// Enable foreign keys
	_, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return err
	}

	// Read the init.sql file
	initSQL, err := os.ReadFile("./internal/db/sql/schema.sql")
	if err != nil {
		return err
	}

	// Execute the SQL script
	_, err = db.Exec(string(initSQL))
	if err != nil {
		return err
	}

	fmt.Println("database schema initialized successfully.")
	return nil
}
