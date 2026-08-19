package services

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/blkhole-sh/blkhole/internal/repos"

	_ "github.com/mattn/go-sqlite3"
)

type noopContentBlocker struct{}

func (n *noopContentBlocker) Init() error                            { return nil }
func (n *noopContentBlocker) Reload() error                          { return nil }
func (n *noopContentBlocker) IsBlocked(string, string) (bool, error) { return false, nil }
func (n *noopContentBlocker) EffectiveBlockedDomains(string) ([]string, error) {
	return []string{}, nil
}

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
			name:     "Empty file",
			content:  "",
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

	svc := NewListService(listRepo, ruleRepo, domainRepo, &noopContentBlocker{})

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

func TestNewListService(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	listRepo := repos.NewListRepo(db)
	ruleRepo := repos.NewRuleRepo(db)
	domainRepo := repos.NewDomainRepo(db)

	svc := NewListService(listRepo, ruleRepo, domainRepo, nil)
	if svc == nil {
		t.Fatal("expected non-nil ListsService")
	}
}

func TestLoadList_HTTPSFile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	listRepo := repos.NewListRepo(db)
	ruleRepo := repos.NewRuleRepo(db)
	domainRepo := repos.NewDomainRepo(db)
	userRepo := repos.NewUserRepo(db)

	svc := NewListService(listRepo, ruleRepo, domainRepo, &noopContentBlocker{})

	user := &model.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hash",
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Create test server
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/notfound" {
			http.NotFound(w, r)
			return
		}
		content := `
||https-example.com^
||https-test.com^
`
		w.Write([]byte(content))
	}))
	defer ts.Close()

	// Temporarily override http.DefaultClient to trust the test server certificate
	oldClient := http.DefaultClient
	http.DefaultClient = ts.Client()
	defer func() { http.DefaultClient = oldClient }()

	// Create list pointing to the test server
	list := &model.List{
		Name:        "Test HTTPS List",
		Description: "Test Description",
		Source:      ts.URL,
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

	if len(list.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(list.Rules))
	}

	rules, err := listRepo.LoadRules(listID)
	if err != nil {
		t.Fatalf("failed to load rules: %v", err)
	}
	if len(rules) != 2 {
		t.Errorf("expected 2 persisted rules, got %d", len(rules))
	}
}

func TestLoadList_EmptySource(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	svc := NewListService(nil, nil, nil, nil) // Repos not needed for this check

	list := &model.List{
		Source: "",
	}

	err := svc.LoadList(list)
	if err != nil {
		t.Errorf("expected nil error for empty source, got %v", err)
	}
}

func TestLoadList_InlineDomainsSource(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := repos.NewUserRepo(db)
	listRepo := repos.NewListRepo(db)
	ruleRepo := repos.NewRuleRepo(db)
	domainRepo := repos.NewDomainRepo(db)

	svc := NewListService(listRepo, ruleRepo, domainRepo, &noopContentBlocker{})

	user := &model.User{Name: "Test User", Email: "test@example.com", PasswordHash: "hash"}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	list := &model.List{Name: "Manual", Source: "example.com\nads.tracker.com", UserID: user.ID}
	listID, err := listRepo.Create(list)
	if err != nil {
		t.Fatalf("failed to create list: %v", err)
	}
	list.ID = listID

	if err := svc.LoadList(list); err != nil {
		t.Errorf("expected no error for inline domain source, got %v", err)
	}
	if list.Count == 0 {
		t.Errorf("expected domains to be loaded from inline text, got count 0")
	}
}

func TestLoadList_NotFoundHTTPSFile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	domainRepo := repos.NewDomainRepo(db)
	svc := NewListService(nil, nil, domainRepo, nil) // Only domain repo needed for readFromSource initially

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	// Temporarily override http.DefaultClient to trust the test server certificate
	oldClient := http.DefaultClient
	http.DefaultClient = ts.Client()
	defer func() { http.DefaultClient = oldClient }()

	list := &model.List{
		Source: ts.URL,
	}

	err := svc.LoadList(list)
	if err == nil {
		t.Errorf("expected error for not found HTTPS file, got nil")
	} else if !strings.Contains(err.Error(), "bad status: 404 Not Found") {
		t.Errorf("expected error containing 'bad status: 404', got %v", err)
	}
}

func TestLoadList_NotFoundLocalFile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	domainRepo := repos.NewDomainRepo(db)
	svc := NewListService(nil, nil, domainRepo, nil)

	list := &model.List{
		Source: "/tmp/this-file-does-not-exist-12345.txt",
	}

	err := svc.LoadList(list)
	if err == nil {
		t.Errorf("expected error for not found local file, got nil")
	} else if !strings.Contains(err.Error(), "failed to open local file") {
		t.Errorf("expected error containing 'failed to open local file', got %v", err)
	}
}
