package repos

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/lemon3studio/blkhole/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create tables
	queries := []string{
		`CREATE TABLE domain (
			id INTEGER PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE rule (
			id INTEGER PRIMARY KEY,
			allowed INTEGER NOT NULL,
			domain_id INTEGER NOT NULL REFERENCES domain (id) ON DELETE CASCADE,
			UNIQUE(domain_id, allowed)
		)`,
		`CREATE TABLE list (
			id INTEGER PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			source TEXT,
			user_id INTEGER NOT NULL
		)`,
		`CREATE TABLE list_rule (
			list_id INTEGER NOT NULL REFERENCES list (id) ON DELETE CASCADE,
			rule_id INTEGER NOT NULL REFERENCES rule (id),
			PRIMARY KEY (list_id, rule_id)
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("Failed to execute query %s: %v", query, err)
		}
	}

	return db
}

func TestBatchCreateOrGet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRuleRepo(db)

	// Create some domains
	domains := []string{"example.com", "test.com", "foo.com", "bar.com"}
	domainIDs := make([]int, len(domains))
	for i, d := range domains {
		res, err := db.Exec("INSERT INTO domain (name) VALUES (?)", d)
		if err != nil {
			t.Fatalf("Failed to create domain %s: %v", d, err)
		}
		id, _ := res.LastInsertId()
		domainIDs[i] = int(id)
	}

	// Prepare rules
	rules := []*model.Rule{
		{DomainID: domainIDs[0], Allowed: false},
		{DomainID: domainIDs[1], Allowed: true},
		{DomainID: domainIDs[2], Allowed: false},
	}

	// Test 1: Insert new rules
	if err := repo.BatchCreateOrGet(rules); err != nil {
		t.Fatalf("BatchCreateOrGet failed: %v", err)
	}

	// Verify IDs are populated
	for i, r := range rules {
		if r.ID == 0 {
			t.Errorf("Rule %d ID not populated", i)
		}
	}

	// Verify count
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM rule").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rules: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 rules, got %d", count)
	}

	// Test 2: Insert mixed existing and new rules
	newRule := &model.Rule{DomainID: domainIDs[3], Allowed: false}
	mixedRules := []*model.Rule{
		rules[0], // Existing
		newRule,  // New
		rules[1], // Existing
	}

	// Reset IDs for existing rules to ensure they are re-fetched
	rules[0].ID = 0
	rules[1].ID = 0

	if err := repo.BatchCreateOrGet(mixedRules); err != nil {
		t.Fatalf("BatchCreateOrGet mixed failed: %v", err)
	}

	if mixedRules[0].ID == 0 {
		t.Errorf("Existing rule 0 ID not populated")
	}
	if mixedRules[1].ID == 0 {
		t.Errorf("New rule ID not populated")
	}
	if mixedRules[2].ID == 0 {
		t.Errorf("Existing rule 1 ID not populated")
	}

	// Verify count
	err = db.QueryRow("SELECT COUNT(*) FROM rule").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rules: %v", err)
	}
	if count != 4 {
		t.Errorf("Expected 4 rules, got %d", count)
	}
}

func TestBatchLinkToList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRuleRepo(db)

	listID := 1
	_, err := db.Exec("INSERT INTO list (id, name, user_id) VALUES (?, 'Test List', 1)", listID)
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Create rules
	ruleIDs := []int{}
	for i := 1; i <= 10; i++ {
		// Create domain and rule manually
		_, err := db.Exec("INSERT INTO domain (name) VALUES (?)", fmt.Sprintf("d%d.com", i))
		if err != nil {
			t.Fatal(err)
		}
		res, err := db.Exec("INSERT INTO rule (domain_id, allowed) VALUES (?, 0)", i)
		if err != nil {
			t.Fatal(err)
		}
		id, _ := res.LastInsertId()
		ruleIDs = append(ruleIDs, int(id))
	}

	// Test BatchLinkToList
	if err := repo.BatchLinkToList(ruleIDs, listID); err != nil {
		t.Fatalf("BatchLinkToList failed: %v", err)
	}

	// Verify links
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM list_rule WHERE list_id = ?", listID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count list rules: %v", err)
	}
	if count != 10 {
		t.Errorf("Expected 10 links, got %d", count)
	}

	// Test idempotency (should handle existing links gracefully)
	if err := repo.BatchLinkToList(ruleIDs, listID); err != nil {
		t.Fatalf("BatchLinkToList idempotency failed: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM list_rule WHERE list_id = ?", listID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count list rules: %v", err)
	}
	if count != 10 {
		t.Errorf("Expected 10 links, got %d", count)
	}
}
