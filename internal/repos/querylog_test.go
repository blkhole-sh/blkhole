package repos

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupQueryLogTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	for _, query := range []string{
		`CREATE TABLE user (id INTEGER PRIMARY KEY, name TEXT, email TEXT, password_hash TEXT)`,
		`CREATE TABLE device (id INTEGER PRIMARY KEY, hash TEXT, name TEXT, os TEXT, user_id INTEGER)`,
		`CREATE TABLE query_log (id INTEGER PRIMARY KEY, device_hash TEXT, domain TEXT, blocked INTEGER, timestamp INTEGER)`,
	} {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("schema setup: %v", err)
		}
	}
	return db
}

func TestQueryLogRepoFindFilteredByUser(t *testing.T) {
	db := setupQueryLogTestDB(t)
	defer db.Close()

	db.Exec(`INSERT INTO user (id, name) VALUES (1, 'A'), (2, 'B')`)
	db.Exec(`INSERT INTO device (id, hash, name, user_id) VALUES (1, 'mac', 'MacBook', 1), (2, 'tv', 'TV', 1), (3, 'other', 'Other', 2)`)
	db.Exec(`INSERT INTO query_log (device_hash, domain, blocked, timestamp) VALUES
		('mac', 'ads.example', 1, 300),
		('mac', 'allowed.example', 0, 200),
		('tv', 'tv.example', 0, 250),
		('other', 'private.example', 1, 300)`)

	repo := NewQueryLogRepo(db)
	logs, total, err := repo.FindFilteredByUser(1, QueryLogFilter{
		DeviceIDs: []int{1},
		From:      time.Unix(150, 0),
		To:        time.Unix(350, 0),
		Limit:     1,
		Offset:    1,
	})
	if err != nil {
		t.Fatalf("FindFilteredByUser: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(logs) != 1 || logs[0].Domain != "allowed.example" || logs[0].DeviceName != "MacBook" {
		t.Fatalf("unexpected filtered logs: %#v", logs)
	}
}

func TestQueryLogRepoDashboardStats(t *testing.T) {
	db := setupQueryLogTestDB(t)
	defer db.Close()

	db.Exec(`INSERT INTO user (id, name) VALUES (1, 'A')`)
	db.Exec(`INSERT INTO device (id, hash, name, user_id) VALUES (1, 'mac', 'MacBook', 1), (2, 'tv', 'TV', 1)`)
	from := time.Date(2026, 8, 20, 0, 0, 0, 0, time.Local)
	macMorning := from.Add(9 * time.Hour).Unix()
	macLater := from.Add(10 * time.Hour).Unix()
	tvEvening := from.Add(20 * time.Hour).Unix()
	db.Exec(`INSERT INTO query_log (device_hash, domain, blocked, timestamp) VALUES
		('mac', 'ads.example', 1, ?),
		('mac', 'ads.example', 1, ?),
		('mac', 'allowed.example', 0, ?),
		('tv', 'tv.example', 1, ?)`, macMorning, macMorning, macLater, tvEvening)

	repo := NewQueryLogRepo(db)
	domains, err := repo.GetDomainStats([]string{"mac", "tv"}, from, from.Add(24*time.Hour), true, 5)
	if err != nil {
		t.Fatalf("GetDomainStats: %v", err)
	}
	if len(domains) != 2 || domains[0].Domain != "ads.example" || domains[0].Count != 2 {
		t.Fatalf("unexpected domain stats: %#v", domains)
	}

	activity, err := repo.GetHourlyActivity([]string{"mac", "tv"}, from, from.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("GetHourlyActivity: %v", err)
	}
	if activity["mac"][9] != 2 || activity["mac"][10] != 1 || activity["tv"][20] != 1 {
		t.Fatalf("unexpected hourly activity: %#v", activity)
	}

	lastSeen, err := repo.GetLastQueries([]string{"mac", "tv"})
	if err != nil {
		t.Fatalf("GetLastQueries: %v", err)
	}
	if lastSeen["mac"].Unix() != macLater || lastSeen["tv"].Unix() != tvEvening {
		t.Fatalf("unexpected last queries: %#v", lastSeen)
	}
}
