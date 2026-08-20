// Package db provides database schema initialization for the blkhole DNS blocker.
package db

import (
	"database/sql"
	"embed"
	"net/url"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Open opens a SQLite database with foreign keys enabled on every connection.
func Open(path string) (*sql.DB, error) {
	dsn := &url.URL{Scheme: "file", Path: path}
	query := dsn.Query()
	query.Set("_foreign_keys", "on")
	dsn.RawQuery = query.Encode()
	return sql.Open("sqlite3", dsn.String())
}

func Init(db *sql.DB) error {
	// Keep foreign keys enabled for callers that supply their own single connection.
	// Open configures the setting for every connection in the production pool.
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return err
	}

	// Enable WAL mode for better concurrent performance
	_, err := db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		return err
	}

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	return goose.Up(db, "migrations")
}
