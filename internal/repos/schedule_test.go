package repos

import (
	"database/sql"
	"testing"

	"github.com/blkhole-sh/blkhole/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

func setupScheduleTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	queries := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE user (id INTEGER PRIMARY KEY, name TEXT NOT NULL, email TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL)`,
		`CREATE TABLE device (id INTEGER PRIMARY KEY, hash TEXT UNIQUE NOT NULL, name TEXT NOT NULL, os TEXT NOT NULL, user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE)`,
		`CREATE TABLE schedule (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			start_time TEXT NOT NULL,
			end_time TEXT NOT NULL,
			days INTEGER NOT NULL DEFAULT 127,
			active INTEGER NOT NULL DEFAULT 1,
			is_default INTEGER NOT NULL DEFAULT 0,
			user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
			CHECK (start_time < end_time)
		)`,
		`CREATE TABLE list (id INTEGER PRIMARY KEY, name TEXT UNIQUE NOT NULL, description TEXT, source TEXT, user_id INTEGER NOT NULL, count INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE rule (id INTEGER PRIMARY KEY, allowed INTEGER NOT NULL, domain_id INTEGER NOT NULL)`,
		`CREATE TABLE domain (id INTEGER PRIMARY KEY, name TEXT UNIQUE NOT NULL)`,
		`CREATE TABLE device_schedule (device_id INTEGER NOT NULL REFERENCES device(id) ON DELETE CASCADE, schedule_id INTEGER NOT NULL REFERENCES schedule(id) ON DELETE CASCADE, PRIMARY KEY(device_id, schedule_id))`,
		`CREATE TABLE list_schedule (list_id INTEGER NOT NULL REFERENCES list(id) ON DELETE CASCADE, schedule_id INTEGER NOT NULL REFERENCES schedule(id) ON DELETE CASCADE, PRIMARY KEY(list_id, schedule_id))`,
		`CREATE TABLE schedule_rule (schedule_id INTEGER NOT NULL REFERENCES schedule(id) ON DELETE CASCADE, rule_id INTEGER NOT NULL, PRIMARY KEY(schedule_id, rule_id))`,
		`CREATE TABLE list_rule (list_id INTEGER NOT NULL, rule_id INTEGER NOT NULL, PRIMARY KEY(list_id, rule_id))`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("schema: %q: %v", q, err)
		}
	}
	return db
}

func insertSchedUser(t *testing.T, db *sql.DB) int {
	t.Helper()
	res, err := db.Exec("INSERT INTO user (name, email, password_hash) VALUES ('Test', 'sched@example.com', 'hash')")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}

func newTestSchedule(userID int) *model.Schedule {
	return &model.Schedule{
		Name:      "Test Schedule",
		StartTime: "08:00",
		EndTime:   "17:00",
		Active:    true,
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		UserID:    userID,
	}
}

func TestScheduleRepo_Create(t *testing.T) {
	db := setupScheduleTestDB(t)
	defer db.Close()
	userID := insertSchedUser(t, db)

	repo := NewScheduleRepo(db)
	s := newTestSchedule(userID)
	id, err := repo.Create(s)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero schedule ID")
	}
}

func TestScheduleRepo_FindByID(t *testing.T) {
	db := setupScheduleTestDB(t)
	defer db.Close()
	userID := insertSchedUser(t, db)

	repo := NewScheduleRepo(db)
	s := newTestSchedule(userID)
	id, _ := repo.Create(s)

	found, err := repo.FindByID(id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != "Test Schedule" {
		t.Errorf("expected name 'Test Schedule', got %s", found.Name)
	}
	if !found.Monday {
		t.Error("expected Monday to be true")
	}
}

func TestScheduleRepo_FindByID_NotFound(t *testing.T) {
	db := setupScheduleTestDB(t)
	defer db.Close()

	repo := NewScheduleRepo(db)
	_, err := repo.FindByID(9999)
	if err == nil {
		t.Error("expected error for missing schedule, got nil")
	}
}

func TestScheduleRepo_Update(t *testing.T) {
	db := setupScheduleTestDB(t)
	defer db.Close()
	userID := insertSchedUser(t, db)

	repo := NewScheduleRepo(db)
	s := newTestSchedule(userID)
	id, _ := repo.Create(s)

	updated := &model.Schedule{
		Name:      "Updated Schedule",
		StartTime: "09:00",
		EndTime:   "18:00",
		Active:    false,
		Monday:    true,
	}
	if err := repo.Update(id, updated); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, _ := repo.FindByID(id)
	if found.Name != "Updated Schedule" {
		t.Errorf("expected 'Updated Schedule', got %s", found.Name)
	}
	if found.Active {
		t.Error("expected Active=false after update")
	}
}

func TestScheduleRepo_Delete(t *testing.T) {
	db := setupScheduleTestDB(t)
	defer db.Close()
	userID := insertSchedUser(t, db)

	repo := NewScheduleRepo(db)
	s := newTestSchedule(userID)
	id, _ := repo.Create(s)

	if err := repo.Delete(id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := repo.FindByID(id)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestScheduleRepo_FindByUser(t *testing.T) {
	db := setupScheduleTestDB(t)
	defer db.Close()
	userID := insertSchedUser(t, db)

	repo := NewScheduleRepo(db)
	s1 := newTestSchedule(userID)
	s2 := newTestSchedule(userID)
	s2.Name = "Second"
	s2.StartTime = "10:00"
	s2.EndTime = "20:00"
	repo.Create(s1)
	repo.Create(s2)

	schedules, err := repo.FindByUser(userID)
	if err != nil {
		t.Fatalf("FindByUser: %v", err)
	}
	if len(schedules) != 2 {
		t.Errorf("expected 2 schedules, got %d", len(schedules))
	}
}

func TestScheduleRepo_FindByUser_Empty(t *testing.T) {
	db := setupScheduleTestDB(t)
	defer db.Close()

	repo := NewScheduleRepo(db)
	schedules, err := repo.FindByUser(999)
	if err != nil {
		t.Fatalf("FindByUser: %v", err)
	}
	if len(schedules) != 0 {
		t.Errorf("expected empty slice, got %d", len(schedules))
	}
}
