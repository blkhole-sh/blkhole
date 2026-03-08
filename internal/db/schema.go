// Package db provides database schema initialization for the blkhole DNS blocker.
package db

import (
	"database/sql"
	_ "embed"
	"strings"

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

	// Enable WAL mode for better concurrent performance
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		return err
	}

	// Execute the embedded SQL script
	_, err = db.Exec(schemaSQL)
	if err != nil {
		return err
	}

	// Migration: Add count column to list table if it doesn't exist (for existing DBs)
	// Since SQLite doesn't support IF NOT EXISTS in ADD COLUMN, we'll try and ignore the error
	// if it says duplicate column name.
	_, err = db.Exec("ALTER TABLE list ADD COLUMN count INTEGER NOT NULL DEFAULT 0;")
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	// Calculate and update count for all lists to ensure consistency
	_, err = db.Exec("UPDATE list SET count = (SELECT COUNT(*) FROM list_rule WHERE list_id = list.id);")
	if err != nil {
		return err
	}

	return nil
}
