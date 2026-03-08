package repos

import (
	"database/sql"
	"testing"

	"github.com/lemon3studio/blkhole/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

func setupUserTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create tables needed for user repo tests
	queries := []string{
		`CREATE TABLE user (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL
		)`,
		`CREATE TABLE device (
			id INTEGER PRIMARY KEY,
			hash TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			os TEXT NOT NULL,
			user_id INTEGER NOT NULL REFERENCES user (id) ON DELETE CASCADE
		)`,
		`CREATE TABLE list (
			id INTEGER PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			source TEXT,
			user_id INTEGER NOT NULL REFERENCES user (id) ON DELETE CASCADE
		)`,
		`CREATE TABLE schedule (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			start_time TEXT NOT NULL,
			end_time TEXT NOT NULL,
			active INTEGER NOT NULL DEFAULT 1,
			monday INTEGER NOT NULL DEFAULT 1,
			tuesday INTEGER NOT NULL DEFAULT 1,
			wednesday INTEGER NOT NULL DEFAULT 1,
			thursday INTEGER NOT NULL DEFAULT 1,
			friday INTEGER NOT NULL DEFAULT 1,
			saturday INTEGER NOT NULL DEFAULT 1,
			sunday INTEGER NOT NULL DEFAULT 1,
			user_id INTEGER NOT NULL REFERENCES user (id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("Failed to execute query %s: %v", query, err)
		}
	}

	return db
}

func TestUserRepo_Create(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)

	user := &model.User{
		Name:         "John Doe",
		Email:        "john.doe@example.com",
		PasswordHash: "hashed_password",
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Errorf("Expected user ID to be populated, got 0")
	}

	// Verify insertion
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM user WHERE id = ?", user.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to verify user creation: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 user, got %d", count)
	}
}

func TestUserRepo_Update(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)

	// Create initial user
	user := &model.User{
		Name:         "Jane Doe",
		Email:        "jane.doe@example.com",
		PasswordHash: "hash1",
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	// Update user
	user.Name = "Jane Smith"
	user.Email = "jane.smith@example.com"
	user.PasswordHash = "hash2"

	err := repo.Update(user.ID, user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify update
	var name, email, pass string
	err = db.QueryRow("SELECT name, email, password_hash FROM user WHERE id = ?", user.ID).Scan(&name, &email, &pass)
	if err != nil {
		t.Fatalf("Failed to verify user update: %v", err)
	}

	if name != "Jane Smith" || email != "jane.smith@example.com" || pass != "hash2" {
		t.Errorf("User details not updated correctly. Got Name: %s, Email: %s, Pass: %s", name, email, pass)
	}
}

func TestUserRepo_Delete(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)

	user := &model.User{
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hash",
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	err := repo.Delete(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify deletion
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM user WHERE id = ?", user.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to verify user deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 users after deletion, got %d", count)
	}
}

func TestUserRepo_FindByIDAndEmail(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)

	user := &model.User{
		Name:         "Bob",
		Email:        "bob@example.com",
		PasswordHash: "hash",
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	// Add relations to verify LoadRelations
	if _, err := db.Exec("INSERT INTO device (id, hash, name, os, user_id) VALUES (1, 'hash1', 'dev1', 'macOS', ?)", user.ID); err != nil {
		t.Fatalf("Failed to insert device: %v", err)
	}
	if _, err := db.Exec("INSERT INTO list (id, name, user_id) VALUES (1, 'list1', ?)", user.ID); err != nil {
		t.Fatalf("Failed to insert list: %v", err)
	}
	if _, err := db.Exec("INSERT INTO schedule (id, name, start_time, end_time, user_id) VALUES (1, 'sched1', '08:00', '17:00', ?)", user.ID); err != nil {
		t.Fatalf("Failed to insert schedule: %v", err)
	}

	// Test FindByID
	foundUser, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if foundUser.Name != user.Name || foundUser.Email != user.Email || foundUser.PasswordHash != user.PasswordHash {
		t.Errorf("FindByID returned wrong user data")
	}

	if len(foundUser.DeviceIDs) != 1 || foundUser.DeviceIDs[0] != 1 {
		t.Errorf("FindByID failed to load DeviceIDs correctly. Got: %v", foundUser.DeviceIDs)
	}
	if len(foundUser.ListIDs) != 1 || foundUser.ListIDs[0] != 1 {
		t.Errorf("FindByID failed to load ListIDs correctly. Got: %v", foundUser.ListIDs)
	}
	if len(foundUser.ScheduleIDs) != 1 || foundUser.ScheduleIDs[0] != 1 {
		t.Errorf("FindByID failed to load ScheduleIDs correctly. Got: %v", foundUser.ScheduleIDs)
	}

	// Test FindByEmail
	foundUserByEmail, err := repo.FindByEmail(user.Email)
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}

	if foundUserByEmail.ID != user.ID {
		t.Errorf("FindByEmail returned wrong user ID")
	}

	// Verify NotFound errors
	_, err = repo.FindByID(999)
	if err == nil {
		t.Errorf("FindByID for non-existent user should return error")
	}

	_, err = repo.FindByEmail("nonexistent@example.com")
	if err == nil {
		t.Errorf("FindByEmail for non-existent user should return error")
	}
}

func TestUserRepo_LoadEmptyRelations(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepo(db)

	user := &model.User{
		Name:         "Empty User",
		Email:        "empty@example.com",
		PasswordHash: "hash",
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	// Load relations for a user without devices, lists, or schedules
	err := repo.LoadRelations(user)
	if err != nil {
		t.Fatalf("Failed to load empty relations: %v", err)
	}

	if user.DeviceIDs == nil || len(user.DeviceIDs) != 0 {
		t.Errorf("Expected empty device IDs slice, got %v", user.DeviceIDs)
	}
	if user.ListIDs == nil || len(user.ListIDs) != 0 {
		t.Errorf("Expected empty list IDs slice, got %v", user.ListIDs)
	}
	if user.ScheduleIDs == nil || len(user.ScheduleIDs) != 0 {
		t.Errorf("Expected empty schedule IDs slice, got %v", user.ScheduleIDs)
	}
}
