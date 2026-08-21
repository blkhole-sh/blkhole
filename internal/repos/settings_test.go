package repos

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupSettingsTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE server_setting (key TEXT PRIMARY KEY, value TEXT NOT NULL)`); err != nil {
		db.Close()
		t.Fatalf("create settings table: %v", err)
	}
	return db
}

func TestSettingsRepoPersistsUpstreamDNS(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()
	repo := NewSettingsRepo(db)

	value, err := repo.GetUpstreamDNS("9.9.9.9:53")
	if err != nil || value != "9.9.9.9:53" {
		t.Fatalf("initial value = %q, %v", value, err)
	}
	if err := repo.UpdateUpstreamDNS("1.1.1.1:53"); err != nil {
		t.Fatalf("UpdateUpstreamDNS: %v", err)
	}
	value, err = repo.GetUpstreamDNS("9.9.9.9:53")
	if err != nil || value != "1.1.1.1:53" {
		t.Fatalf("persisted value = %q, %v", value, err)
	}
}
