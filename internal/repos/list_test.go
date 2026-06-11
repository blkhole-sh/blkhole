package repos_test

import (
	"database/sql"
	"testing"

	"github.com/blkhole-sh/blkhole/internal/db"
	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/blkhole-sh/blkhole/internal/repos"
	_ "github.com/mattn/go-sqlite3"
)

func TestListCount(t *testing.T) {
	// Setup DB
	d, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Init(d); err != nil {
		t.Fatal(err)
	}

	// Setup Repos
	listRepo := repos.NewListRepo(d)
	ruleRepo := repos.NewRuleRepo(d)
	userRepo := repos.NewUserRepo(d)
	domainRepo := repos.NewDomainRepo(d)

	// Create User
	user := &model.User{Name: "Test User", Email: "test@example.com", PasswordHash: "hash"}
	if err := userRepo.Create(user); err != nil {
		t.Fatal(err)
	}

	// Create List
	list := &model.List{Name: "Test List", UserID: user.ID}
	listID, err := listRepo.Create(list)
	if err != nil {
		t.Fatal(err)
	}

	// Verify initial count
	l, err := listRepo.FindByID(listID)
	if err != nil {
		t.Fatal(err)
	}
	if l.Count != 0 {
		t.Errorf("expected count 0, got %d", l.Count)
	}
	if len(l.Rules) != 0 {
		t.Errorf("expected rules empty, got %d", len(l.Rules))
	}

	// Add Rules
	domain := &model.Domain{Name: "example.com"}
	domainID, err := domainRepo.CreateOrGet(domain)
	if err != nil {
		t.Fatal(err)
	}
	rule := &model.Rule{DomainID: domainID, Allowed: false}
	ruleID, err := ruleRepo.CreateOrGet(rule)
	if err != nil {
		t.Fatal(err)
	}

	// Link Rule to List
	if err := ruleRepo.LinkToList(ruleID, listID); err != nil {
		t.Fatal(err)
	}

	// Verify count increased
	l, err = listRepo.FindByID(listID)
	if err != nil {
		t.Fatal(err)
	}
	if l.Count != 1 {
		t.Errorf("expected count 1, got %d", l.Count)
	}

	// Unlink Rule
	if err := ruleRepo.UnlinkFromList(ruleID, listID); err != nil {
		t.Fatal(err)
	}

	// Verify count decreased
	l, err = listRepo.FindByID(listID)
	if err != nil {
		t.Fatal(err)
	}
	if l.Count != 0 {
		t.Errorf("expected count 0, got %d", l.Count)
	}
}
