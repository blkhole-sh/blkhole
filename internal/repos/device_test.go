package repos

import (
	"database/sql"
	"testing"

	"github.com/lemon3studio/blkhole/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

func setupDeviceTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	queries := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE user (id INTEGER PRIMARY KEY, name TEXT NOT NULL, email TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL)`,
		`CREATE TABLE device (id INTEGER PRIMARY KEY, hash TEXT UNIQUE NOT NULL, name TEXT NOT NULL, os TEXT NOT NULL, user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE)`,
		`CREATE TABLE schedule (id INTEGER PRIMARY KEY, name TEXT NOT NULL, start_time TEXT NOT NULL, end_time TEXT NOT NULL, days INTEGER NOT NULL DEFAULT 127, active INTEGER NOT NULL DEFAULT 1, user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE)`,
		`CREATE TABLE device_schedule (device_id INTEGER NOT NULL REFERENCES device(id) ON DELETE CASCADE, schedule_id INTEGER NOT NULL REFERENCES schedule(id) ON DELETE CASCADE, PRIMARY KEY(device_id, schedule_id))`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("schema setup: %v", err)
		}
	}
	return db
}

func insertTestUser(t *testing.T, db *sql.DB) int {
	t.Helper()
	res, err := db.Exec("INSERT INTO user (name, email, password_hash) VALUES ('Test', 'test@example.com', 'hash')")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}

func TestDeviceRepo_Create(t *testing.T) {
	db := setupDeviceTestDB(t)
	defer db.Close()
	userID := insertTestUser(t, db)

	repo := NewDeviceRepo(db)
	d := &model.Device{Hash: "abc", Name: "iPhone", OS: model.IOS, UserID: userID}
	if err := repo.Create(d); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if d.ID == 0 {
		t.Error("expected non-zero ID after Create")
	}
}

func TestDeviceRepo_FindByID(t *testing.T) {
	db := setupDeviceTestDB(t)
	defer db.Close()
	userID := insertTestUser(t, db)

	repo := NewDeviceRepo(db)
	d := &model.Device{Hash: "xyz", Name: "MacBook", OS: model.MacOS, UserID: userID}
	repo.Create(d)

	found, err := repo.FindByID(d.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != "MacBook" {
		t.Errorf("expected name MacBook, got %s", found.Name)
	}
}

func TestDeviceRepo_FindByID_NotFound(t *testing.T) {
	db := setupDeviceTestDB(t)
	defer db.Close()

	repo := NewDeviceRepo(db)
	_, err := repo.FindByID(9999)
	if err == nil {
		t.Error("expected error for missing device, got nil")
	}
}

func TestDeviceRepo_FindByHash(t *testing.T) {
	db := setupDeviceTestDB(t)
	defer db.Close()
	userID := insertTestUser(t, db)

	repo := NewDeviceRepo(db)
	d := &model.Device{Hash: "myhash", Name: "Linux Box", OS: model.Linux, UserID: userID}
	repo.Create(d)

	found, err := repo.FindByHash("myhash")
	if err != nil {
		t.Fatalf("FindByHash: %v", err)
	}
	if found.ID != d.ID {
		t.Errorf("expected id %d, got %d", d.ID, found.ID)
	}
}

func TestDeviceRepo_Update(t *testing.T) {
	db := setupDeviceTestDB(t)
	defer db.Close()
	userID := insertTestUser(t, db)

	repo := NewDeviceRepo(db)
	d := &model.Device{Hash: "updhash", Name: "Old Name", OS: model.Windows, UserID: userID}
	repo.Create(d)

	updated := &model.Device{Name: "New Name", OS: model.MacOS}
	if err := repo.Update(d.ID, updated); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, _ := repo.FindByID(d.ID)
	if found.Name != "New Name" {
		t.Errorf("expected name 'New Name', got %s", found.Name)
	}
}

func TestDeviceRepo_Delete(t *testing.T) {
	db := setupDeviceTestDB(t)
	defer db.Close()
	userID := insertTestUser(t, db)

	repo := NewDeviceRepo(db)
	d := &model.Device{Hash: "delhash", Name: "To Delete", OS: model.Android, UserID: userID}
	repo.Create(d)

	if err := repo.Delete(d.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := repo.FindByID(d.ID)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestDeviceRepo_FindByUser(t *testing.T) {
	db := setupDeviceTestDB(t)
	defer db.Close()
	userID := insertTestUser(t, db)

	repo := NewDeviceRepo(db)
	repo.Create(&model.Device{Hash: "h1", Name: "D1", OS: model.IOS, UserID: userID})
	repo.Create(&model.Device{Hash: "h2", Name: "D2", OS: model.IOS, UserID: userID})

	devices, err := repo.FindByUser(userID)
	if err != nil {
		t.Fatalf("FindByUser: %v", err)
	}
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
}

func TestDeviceRepo_FindByUser_Empty(t *testing.T) {
	db := setupDeviceTestDB(t)
	defer db.Close()

	repo := NewDeviceRepo(db)
	devices, err := repo.FindByUser(999)
	if err != nil {
		t.Fatalf("FindByUser with no results: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("expected empty slice, got %d", len(devices))
	}
}
