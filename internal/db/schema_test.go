package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestOpenEnablesForeignKeysOnEveryConnection(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "pool.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(4)

	ctx := context.Background()
	connections := make([]*sql.Conn, 0, 4)
	for i := 0; i < 4; i++ {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("connection %d: %v", i, err)
		}
		connections = append(connections, conn)
		defer conn.Close()

		var enabled int
		if err := conn.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&enabled); err != nil {
			t.Fatalf("foreign_keys on connection %d: %v", i, err)
		}
		if enabled != 1 {
			t.Fatalf("foreign_keys on connection %d = %d, want 1", i, enabled)
		}
	}
}

func legacySchema() []string {
	return []string{
		`CREATE TABLE user (id INTEGER PRIMARY KEY, name TEXT NOT NULL, email TEXT NOT NULL, password_hash TEXT NOT NULL)`,
		`CREATE TABLE device (id INTEGER PRIMARY KEY, name TEXT NOT NULL, os TEXT NOT NULL, hash TEXT NOT NULL, user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE)`,
		`CREATE TABLE domain (id INTEGER PRIMARY KEY, name TEXT UNIQUE NOT NULL)`,
		`CREATE TABLE rule (id INTEGER PRIMARY KEY, allowed INTEGER NOT NULL, domain_id INTEGER NOT NULL REFERENCES domain(id) ON DELETE CASCADE)`,
		`CREATE TABLE list (id INTEGER PRIMARY KEY, name TEXT NOT NULL, description TEXT, source TEXT, user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE, count INTEGER NOT NULL DEFAULT 0, UNIQUE(name, user_id))`,
		`CREATE TABLE schedule (id INTEGER PRIMARY KEY, name TEXT NOT NULL, start_time TEXT NOT NULL, end_time TEXT NOT NULL, days INTEGER NOT NULL CHECK (days >= 0 AND days < 128), active INTEGER NOT NULL DEFAULT 1, user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE, CHECK (start_time < end_time))`,
		`CREATE TABLE device_schedule (device_id INTEGER NOT NULL REFERENCES device(id) ON DELETE CASCADE, schedule_id INTEGER NOT NULL REFERENCES schedule(id) ON DELETE CASCADE, PRIMARY KEY(device_id, schedule_id))`,
		`CREATE TABLE list_schedule (list_id INTEGER NOT NULL REFERENCES list(id) ON DELETE CASCADE, schedule_id INTEGER NOT NULL REFERENCES schedule(id) ON DELETE CASCADE, PRIMARY KEY(list_id, schedule_id))`,
		`CREATE TABLE list_rule (list_id INTEGER NOT NULL REFERENCES list(id) ON DELETE CASCADE, rule_id INTEGER NOT NULL REFERENCES rule(id), PRIMARY KEY(list_id, rule_id))`,
		`CREATE TABLE schedule_rule (schedule_id INTEGER NOT NULL REFERENCES schedule(id) ON DELETE CASCADE, rule_id INTEGER NOT NULL REFERENCES rule(id), PRIMARY KEY(schedule_id, rule_id))`,
		`CREATE TABLE query_log (id INTEGER PRIMARY KEY, device_hash TEXT NOT NULL, domain TEXT NOT NULL, blocked INTEGER NOT NULL DEFAULT 0, timestamp INTEGER NOT NULL)`,
		`CREATE TABLE goose_db_version (id INTEGER PRIMARY KEY AUTOINCREMENT, version_id BIGINT NOT NULL, is_applied BOOLEAN NOT NULL, tstamp TIMESTAMP DEFAULT (datetime('now')))`,
		`INSERT INTO goose_db_version (version_id, is_applied) VALUES (1, 1)`,
	}
}

func TestInitUpgradesLegacyScheduleSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy.db")
	legacy, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}

	queries := append(legacySchema(),
		`INSERT INTO user (id, name, email, password_hash) VALUES (1, 'One', 'one@example.com', 'hash'), (2, 'Two', 'two@example.com', 'hash')`,
		`INSERT INTO device (id, name, os, hash, user_id) VALUES (1, 'One Device', 'test', 'one', 1), (2, 'Two Device', 'test', 'two', 2)`,
		`INSERT INTO domain (id, name) VALUES (1, 'example.com')`,
		`INSERT INTO rule (id, allowed, domain_id) VALUES (1, 0, 1)`,
		`INSERT INTO list (id, name, user_id) VALUES (1, 'One List', 1)`,
		`INSERT INTO schedule (id, name, start_time, end_time, days, active, user_id) VALUES (1, 'Existing', '08:00', '17:00', 31, 1, 1)`,
		`INSERT INTO device_schedule (device_id, schedule_id) VALUES (1, 1), (2, 1), (999, 1)`,
		`INSERT INTO list_schedule (list_id, schedule_id) VALUES (1, 1), (999, 1)`,
		`INSERT INTO list_rule (list_id, rule_id) VALUES (1, 1), (999, 1)`,
		`INSERT INTO schedule_rule (schedule_id, rule_id) VALUES (1, 1), (999, 1)`,
	)
	for _, query := range queries {
		if _, err := legacy.Exec(query); err != nil {
			legacy.Close()
			t.Fatalf("legacy setup %q: %v", query, err)
		}
	}
	if err := legacy.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if err := Init(db); err != nil {
		t.Fatalf("Init: %v", err)
	}

	var scheduleSQL string
	if err := db.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name='schedule'").Scan(&scheduleSQL); err != nil {
		t.Fatalf("read schedule schema: %v", err)
	}
	if strings.Contains(scheduleSQL, "start_time < end_time") {
		t.Fatalf("legacy time ordering constraint remains: %s", scheduleSQL)
	}
	if _, err := db.Exec(`INSERT INTO schedule (name, start_time, end_time, days, active, user_id, is_default) VALUES ('Always', '00:00', '00:00', 127, 1, 1, 1)`); err != nil {
		t.Fatalf("insert all-day schedule: %v", err)
	}

	var name string
	var isDefault int
	if err := db.QueryRow("SELECT name, is_default FROM schedule WHERE id=1").Scan(&name, &isDefault); err != nil {
		t.Fatalf("read existing schedule: %v", err)
	}
	if name != "Existing" || isDefault != 0 {
		t.Fatalf("existing schedule changed: name=%q is_default=%d", name, isDefault)
	}

	assertRelationCount(t, db, "device_schedule", 2)
	assertRelationCount(t, db, "list_schedule", 1)
	assertRelationCount(t, db, "list_rule", 1)
	assertRelationCount(t, db, "schedule_rule", 1)

	var violations int
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_foreign_key_check").Scan(&violations); err != nil {
		t.Fatalf("foreign key check: %v", err)
	}
	if violations != 0 {
		t.Fatalf("foreign key violations after migration = %d", violations)
	}
}

func TestInitRemovesRowsOrphanedByDeletedUser(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "orphans.db")
	legacy, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}

	queries := append(legacySchema(),
		`INSERT INTO user (id, name, email, password_hash) VALUES (1, 'One', 'one@example.com', 'hash')`,
		`INSERT INTO domain (id, name) VALUES (1, 'example.com')`,
		`INSERT INTO rule (id, allowed, domain_id) VALUES (1, 0, 1)`,
		// user 7 is gone; foreign keys were off when it was deleted, so its rows survived.
		`INSERT INTO device (id, name, os, hash, user_id) VALUES (1, 'Kept', 'test', 'kept', 1), (2, 'Orphan', 'test', 'orphan', 7)`,
		`INSERT INTO list (id, name, user_id) VALUES (1, 'Kept', 1), (16, 'Orphan', 7)`,
		`INSERT INTO schedule (id, name, start_time, end_time, days, active, user_id) VALUES (1, 'Kept', '08:00', '17:00', 31, 1, 1), (2, 'Orphan', '08:00', '17:00', 31, 1, 7)`,
		`INSERT INTO list_rule (list_id, rule_id) VALUES (1, 1), (16, 1)`,
		`INSERT INTO schedule_rule (schedule_id, rule_id) VALUES (1, 1), (2, 1)`,
		`INSERT INTO list_schedule (list_id, schedule_id) VALUES (16, 2)`,
		`INSERT INTO device_schedule (device_id, schedule_id) VALUES (2, 2)`,
	)
	for _, query := range queries {
		if _, err := legacy.Exec(query); err != nil {
			legacy.Close()
			t.Fatalf("legacy setup %q: %v", query, err)
		}
	}
	if err := legacy.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if err := Init(db); err != nil {
		t.Fatalf("Init: %v", err)
	}

	assertRelationCount(t, db, "device", 1)
	assertRelationCount(t, db, "list", 1)
	assertRelationCount(t, db, "schedule", 1)
	assertRelationCount(t, db, "list_rule", 1)
	assertRelationCount(t, db, "schedule_rule", 1)
	assertRelationCount(t, db, "list_schedule", 0)
	assertRelationCount(t, db, "device_schedule", 0)

	var violations int
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_foreign_key_check").Scan(&violations); err != nil {
		t.Fatalf("foreign key check: %v", err)
	}
	if violations != 0 {
		t.Fatalf("foreign key violations after migration = %d", violations)
	}
}

func assertRelationCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&got); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got != want {
		t.Fatalf("%s rows = %d, want %d", table, got, want)
	}
}
