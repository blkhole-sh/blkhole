package repos

import (
	"database/sql"
	"testing"

	"github.com/blkhole-sh/blkhole/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

func setupDomainTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	queries := []string{
		`CREATE TABLE domain (id INTEGER PRIMARY KEY, name TEXT UNIQUE NOT NULL)`,
		`CREATE TABLE rule (id INTEGER PRIMARY KEY, allowed INTEGER NOT NULL, domain_id INTEGER NOT NULL REFERENCES domain(id) ON DELETE CASCADE, UNIQUE(domain_id, allowed))`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

func TestDomainRepo_CreateOrGet_New(t *testing.T) {
	db := setupDomainTestDB(t)
	defer db.Close()

	repo := NewDomainRepo(db)
	d := &model.Domain{Name: "example.com"}
	id, err := repo.CreateOrGet(d)
	if err != nil {
		t.Fatalf("CreateOrGet: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestDomainRepo_CreateOrGet_Idempotent(t *testing.T) {
	db := setupDomainTestDB(t)
	defer db.Close()

	repo := NewDomainRepo(db)
	d := &model.Domain{Name: "example.com"}
	id1, _ := repo.CreateOrGet(d)
	id2, err := repo.CreateOrGet(&model.Domain{Name: "example.com"})
	if err != nil {
		t.Fatalf("CreateOrGet second call: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected same ID for duplicate insert, got %d vs %d", id1, id2)
	}
}

func TestDomainRepo_FindByID(t *testing.T) {
	db := setupDomainTestDB(t)
	defer db.Close()

	repo := NewDomainRepo(db)
	d := &model.Domain{Name: "find-me.com"}
	id, _ := repo.CreateOrGet(d)

	found, err := repo.FindByID(id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != "find-me.com" {
		t.Errorf("expected name find-me.com, got %s", found.Name)
	}
}

func TestDomainRepo_FindByID_NotFound(t *testing.T) {
	db := setupDomainTestDB(t)
	defer db.Close()

	repo := NewDomainRepo(db)
	_, err := repo.FindByID(9999)
	if err == nil {
		t.Error("expected error for missing domain, got nil")
	}
}

func TestDomainRepo_FindAll(t *testing.T) {
	db := setupDomainTestDB(t)
	defer db.Close()

	repo := NewDomainRepo(db)
	repo.CreateOrGet(&model.Domain{Name: "a.com"})
	repo.CreateOrGet(&model.Domain{Name: "b.com"})

	all, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 domains, got %d", len(all))
	}
}

func TestDomainRepo_Delete_Unreferenced(t *testing.T) {
	db := setupDomainTestDB(t)
	defer db.Close()

	repo := NewDomainRepo(db)
	d := &model.Domain{Name: "delete-me.com"}
	id, _ := repo.CreateOrGet(d)

	if err := repo.Delete(id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(id)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
