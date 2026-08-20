package repos

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	appdb "github.com/blkhole-sh/blkhole/internal/db"
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

func TestScheduleRepo_SeedDefaultsAfterMigration(t *testing.T) {
	db, err := appdb.Open(filepath.Join(t.TempDir(), "schedules.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := appdb.Init(db); err != nil {
		t.Fatalf("init db: %v", err)
	}
	userID := insertSchedUser(t, db)

	repo := NewScheduleRepo(db)
	if err := repo.SeedDefaults(userID); err != nil {
		t.Fatalf("SeedDefaults: %v", err)
	}

	schedules, err := repo.FindByUser(userID)
	if err != nil {
		t.Fatalf("FindByUser: %v", err)
	}
	if len(schedules) != 1 {
		t.Fatalf("default schedules = %d, want 1", len(schedules))
	}
	if schedules[0].StartTime != "00:00" || schedules[0].EndTime != "00:00" || !schedules[0].IsDefault {
		t.Fatalf("default schedule = %+v", schedules[0])
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

func TestScheduleRepo_UpdateRollsBackInvalidRelations(t *testing.T) {
	db := setupScheduleTestDB(t)
	defer db.Close()
	userID := insertSchedUser(t, db)

	deviceResult, err := db.Exec("INSERT INTO device (hash, name, os, user_id) VALUES ('device-1', 'Device', 'test', ?)", userID)
	if err != nil {
		t.Fatalf("insert device: %v", err)
	}
	deviceID, _ := deviceResult.LastInsertId()
	listResult, err := db.Exec("INSERT INTO list (name, user_id) VALUES ('List', ?)", userID)
	if err != nil {
		t.Fatalf("insert list: %v", err)
	}
	listID, _ := listResult.LastInsertId()

	repo := NewScheduleRepo(db)
	s := newTestSchedule(userID)
	s.DeviceIDs = []int{int(deviceID)}
	s.ListIDs = []int{int(listID)}
	scheduleID, err := repo.Create(s)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated := newTestSchedule(userID)
	updated.Name = "Must Roll Back"
	updated.StartTime = "09:00"
	updated.EndTime = "18:00"
	updated.DeviceIDs = []int{int(deviceID)}
	updated.ListIDs = []int{999}
	if err := repo.Update(scheduleID, updated); !errors.Is(err, ErrInvalidScheduleRelation) {
		t.Fatalf("Update error = %v, want ErrInvalidScheduleRelation", err)
	}

	found, err := repo.FindByID(scheduleID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != "Test Schedule" || found.StartTime != "08:00" || found.EndTime != "17:00" {
		t.Fatalf("schedule was partially updated: %+v", found)
	}
	if len(found.DeviceIDs) != 1 || found.DeviceIDs[0] != int(deviceID) {
		t.Fatalf("device relations changed: %v", found.DeviceIDs)
	}
	if len(found.ListIDs) != 1 || found.ListIDs[0] != int(listID) {
		t.Fatalf("list relations changed: %v", found.ListIDs)
	}

	updated.DeviceIDs = []int{999}
	updated.ListIDs = []int{int(listID)}
	if err := repo.Update(scheduleID, updated); !errors.Is(err, ErrInvalidScheduleRelation) {
		t.Fatalf("Update error = %v, want ErrInvalidScheduleRelation", err)
	}
	found, err = repo.FindByID(scheduleID)
	if err != nil {
		t.Fatalf("FindByID after invalid device: %v", err)
	}
	if found.Name != "Test Schedule" || len(found.DeviceIDs) != 1 || found.DeviceIDs[0] != int(deviceID) || len(found.ListIDs) != 1 || found.ListIDs[0] != int(listID) {
		t.Fatalf("schedule changed after invalid device: %+v", found)
	}
}

func TestScheduleRepo_CreateRejectsCrossUserRelationsAtomically(t *testing.T) {
	db := setupScheduleTestDB(t)
	defer db.Close()
	userID := insertSchedUser(t, db)
	otherResult, err := db.Exec("INSERT INTO user (name, email, password_hash) VALUES ('Other', 'other@example.com', 'hash')")
	if err != nil {
		t.Fatalf("insert other user: %v", err)
	}
	otherUserID, _ := otherResult.LastInsertId()

	deviceResult, err := db.Exec("INSERT INTO device (hash, name, os, user_id) VALUES ('other-device', 'Other Device', 'test', ?)", otherUserID)
	if err != nil {
		t.Fatalf("insert device: %v", err)
	}
	deviceID, _ := deviceResult.LastInsertId()
	listResult, err := db.Exec("INSERT INTO list (name, user_id) VALUES ('Other List', ?)", otherUserID)
	if err != nil {
		t.Fatalf("insert list: %v", err)
	}
	listID, _ := listResult.LastInsertId()

	repo := NewScheduleRepo(db)
	invalidDevice := newTestSchedule(userID)
	invalidDevice.DeviceIDs = []int{int(deviceID)}
	if _, err := repo.Create(invalidDevice); !errors.Is(err, ErrInvalidScheduleRelation) {
		t.Fatalf("Create with cross-user device error = %v, want ErrInvalidScheduleRelation", err)
	}

	invalidList := newTestSchedule(userID)
	invalidList.ListIDs = []int{int(listID)}
	if _, err := repo.Create(invalidList); !errors.Is(err, ErrInvalidScheduleRelation) {
		t.Fatalf("Create with cross-user list error = %v, want ErrInvalidScheduleRelation", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schedule").Scan(&count); err != nil {
		t.Fatalf("count schedules: %v", err)
	}
	if count != 0 {
		t.Fatalf("partially created schedules = %d, want 0", count)
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

func TestScheduleRepo_FindScheduleRelationships(t *testing.T) {
	db := setupScheduleTestDB(t)
	defer db.Close()
	userID := insertSchedUser(t, db)

	repo := NewScheduleRepo(db)
	s := newTestSchedule(userID)
	scheduleID, err := repo.Create(s)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := db.Exec("INSERT INTO list (id, name, user_id) VALUES (10, 'Shared', ?)", userID); err != nil {
		t.Fatalf("insert list: %v", err)
	}
	if _, err := db.Exec("INSERT INTO rule (id, allowed, domain_id) VALUES (100, 0, 1), (200, 0, 2)"); err != nil {
		t.Fatalf("insert rules: %v", err)
	}
	if err := repo.LinkRule(scheduleID, 100); err != nil {
		t.Fatalf("LinkRule: %v", err)
	}
	if err := repo.LinkList(scheduleID, 10); err != nil {
		t.Fatalf("LinkList: %v", err)
	}
	if _, err := db.Exec("INSERT INTO list_rule (list_id, rule_id) VALUES (10, 200)"); err != nil {
		t.Fatalf("insert list_rule: %v", err)
	}

	scheduleRules, err := repo.FindScheduleRule()
	if err != nil {
		t.Fatalf("FindScheduleRule: %v", err)
	}
	if len(scheduleRules) != 1 || scheduleRules[0].ScheduleID != scheduleID || scheduleRules[0].RuleID != 100 {
		t.Fatalf("FindScheduleRule() = %+v; want only direct rule 100", scheduleRules)
	}

	scheduleLists, err := repo.FindScheduleList()
	if err != nil {
		t.Fatalf("FindScheduleList: %v", err)
	}
	if len(scheduleLists) != 1 || scheduleLists[0].ScheduleID != scheduleID || scheduleLists[0].ListID != 10 {
		t.Fatalf("FindScheduleList() = %+v; want schedule %d list 10", scheduleLists, scheduleID)
	}

	listRules, err := repo.FindListRule()
	if err != nil {
		t.Fatalf("FindListRule: %v", err)
	}
	if len(listRules) != 1 || listRules[0].ListID != 10 || listRules[0].RuleID != 200 {
		t.Fatalf("FindListRule() = %+v; want list 10 rule 200", listRules)
	}
}
