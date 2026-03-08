package services

import (
	"os"
	"strings"
	"testing"

	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/lemon3studio/blkhole/internal/repos"

	_ "github.com/mattn/go-sqlite3"
)

func TestReadAdblockFile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	domainRepo := repos.NewDomainRepo(db)

	tests := []struct {
		name     string
		content  string
		expected int // number of rules
	}{
		{
			name: "Basic adblock rules",
			content: `
! Title: EasyList
||example.com^
||ads.google.com^
`,
			expected: 2,
		},
		{
			name: "Whitelist rule",
			content: `
@@||good-site.com^
||bad-site.com^
`,
			expected: 2,
		},
		{
			name: "Comments and empty lines",
			content: `
! Comment 1
# Comment 2
[Adblock Plus 2.0]

||example.com^
`,
			expected: 1,
		},
		{
			name: "Invalid domain format",
			content: `
||invalid^
||valid.com^
`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			rules, err := readAdblockFile(reader, domainRepo)
			if err != nil {
				t.Fatalf("readAdblockFile failed: %v", err)
			}

			if len(rules) != tt.expected {
				t.Errorf("expected %d rules, got %d", tt.expected, len(rules))
			}
		})
	}
}

func TestReadHostFile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	domainRepo := repos.NewDomainRepo(db)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "Basic hosts file",
			content: `
127.0.0.1 localhost
127.0.0.1 localhost.localdomain
0.0.0.0 ads.example.com
127.0.0.1 tracker.com
`,
			expected: 2, // localhost entries should be skipped
		},
		{
			name: "Comments",
			content: `
# This is a comment
0.0.0.0 bad.com
`,
			expected: 1,
		},
		{
			name: "Invalid format",
			content: `
invalid line
0.0.0.0 valid.com
`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			rules, err := readHostFile(reader, domainRepo)
			if err != nil {
				t.Fatalf("readHostFile failed: %v", err)
			}

			if len(rules) != tt.expected {
				t.Errorf("expected %d rules, got %d", tt.expected, len(rules))
			}
		})
	}
}

func TestReadDomainsFile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	domainRepo := repos.NewDomainRepo(db)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "Basic domains file",
			content: `
example.com
ads.google.com
`,
			expected: 2,
		},
		{
			name: "Comments and whitespace",
			content: `
# Comment
  example.com

bad.com
`,
			expected: 2,
		},
		{
			name: "Invalid domain",
			content: `
invalid
valid.com
`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			rules, err := readDomainsFile(reader, domainRepo)
			if err != nil {
				t.Fatalf("readDomainsFile failed: %v", err)
			}

			if len(rules) != tt.expected {
				t.Errorf("expected %d rules, got %d", tt.expected, len(rules))
			}
		})
	}
}

func TestDetectAndReadFile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	domainRepo := repos.NewDomainRepo(db)

	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "Detect Adblock",
			content: `
[Adblock Plus 2.0]
||example.com^
`,
			expected: 1,
		},
		{
			name: "Detect Hosts",
			content: `
127.0.0.1 localhost
0.0.0.0 example.com
`,
			expected: 1,
		},
		{
			name: "Detect Domains",
			content: `
example.com
test.com
`,
			expected: 2,
		},
		{
			name: "Empty file",
			content: "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			rules, err := detectAndReadFile(reader, domainRepo)
			if err != nil {
				t.Fatalf("detectAndReadFile failed: %v", err)
			}

			if len(rules) != tt.expected {
				t.Errorf("expected %d rules, got %d", tt.expected, len(rules))
			}
		})
	}
}

func TestLoadList_LocalFile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	listRepo := repos.NewListRepo(db)
	ruleRepo := repos.NewRuleRepo(db)
	domainRepo := repos.NewDomainRepo(db)
	userRepo := repos.NewUserRepo(db)

	svc := NewListsService(listRepo, ruleRepo, domainRepo)

	// Create user
	user := &model.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hash",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Create temp file
	tmpfile, err := os.CreateTemp("", "list_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := `
||example.com^
||test.com^
`
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Create list
	list := &model.List{
		Name:        "Test List",
		Description: "Test Description",
		Source:      tmpfile.Name(),
		UserID:      user.ID,
	}
	listID, err := listRepo.Create(list)
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	list.ID = listID

	// Load list
	if err := svc.LoadList(list); err != nil {
		t.Fatalf("LoadList failed: %v", err)
	}

	// Verify rules loaded
	if len(list.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(list.Rules))
	}

	// Verify rules persisted
	// Check linked rules
	rules, err := listRepo.LoadRules(listID)
	if err != nil {
		t.Fatalf("failed to load rules: %v", err)
	}

	if len(rules) != 2 {
		t.Errorf("expected 2 persisted rules, got %d", len(rules))
	}
}
