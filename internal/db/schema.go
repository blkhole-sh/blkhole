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

	// Migration: change list.name unique constraint from global to per-user UNIQUE(name, user_id).
	// SQLite can't drop constraints, so we recreate the table when the old definition is detected.
	var listSQL string
	err = db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='list'`).Scan(&listSQL)
	if err != nil {
		return err
	}
	if strings.Contains(listSQL, "name TEXT UNIQUE") {
		_, err = db.Exec(`PRAGMA foreign_keys = OFF`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`
			CREATE TABLE list_new (
				id INTEGER PRIMARY KEY,
				name TEXT NOT NULL,
				description TEXT,
				source TEXT,
				user_id INTEGER NOT NULL REFERENCES user (id) ON DELETE CASCADE,
				count INTEGER NOT NULL DEFAULT 0,
				UNIQUE(name, user_id)
			)
		`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`INSERT INTO list_new SELECT id, name, description, source, user_id, count FROM list`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`DROP TABLE list`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`ALTER TABLE list_new RENAME TO list`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`PRAGMA foreign_keys = ON`)
		if err != nil {
			return err
		}
	}

	return nil
}
