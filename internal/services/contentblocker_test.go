// Package services provides unit tests for the content blocking functionality.
// These tests verify that domains are properly blocked based on schedules, lists, and rules.
package services

import (
	"database/sql"
	"testing"
	"time"

	schema "github.com/lemon3studio/blkhole/internal/db"
	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/lemon3studio/blkhole/internal/repos"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for in-memory test database
)

// setupTestDB creates an in-memory SQLite database for testing.
// This provides a clean, isolated database for each test run.
//
// Returns:
//   - *sql.DB: Configured database connection with initialized schema
//
// The database is automatically cleaned up when the connection is closed.
func setupTestDB(t *testing.T) *sql.DB {
	// Create in-memory SQLite database for testing
	// ":memory:" creates a temporary database that exists only in RAM
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Initialize schema with all required tables (users, devices, schedules, etc.)
	if err := schema.Init(db); err != nil {
		t.Fatalf("Failed to initialize test database schema: %v", err)
	}

	return db
}

// createTestUser creates a test user in the database.
// Users are required as owners for devices, lists, and schedules.
//
// Parameters:
//   - t: Testing context for error reporting
//   - users: Repository for user operations
//
// Returns:
//   - int: The ID of the created user
func createTestUser(t *testing.T, users repos.UserRepo) int {
	user := &model.User{
		Name:         "Test User",        // Human-readable name
		Email:        "test@example.com", // Valid email format
		PasswordHash: "hashed-password",  // Pre-hashed password (not used in blocking tests)
	}

	if err := users.Create(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user.ID
}

// createTestDevice creates a test device in the database.
// Devices are the targets for DNS blocking - each device can be linked to schedules.
//
// Parameters:
//   - t: Testing context for error reporting
//   - devices: Repository for device operations
//   - userID: ID of the user who owns this device
//
// Returns:
//   - string: The hash of the created device (still used for DNS queries)
func createTestDevice(t *testing.T, devices repos.DeviceRepo, userID int) string {
	device := &model.Device{
		Hash:   "test-device-hash", // Fixed hash matching the one used in mobileconfig URLs
		Name:   "Test Device",      // Human-readable name for the device
		OS:     model.IOS,          // Device operating system (iOS/Android/etc.)
		UserID: userID,             // Links device to its owner
	}

	if err := devices.Create(device); err != nil {
		t.Fatalf("Failed to create test device: %v", err)
	}

	return device.Hash
}

// createTestList creates a test list with blocking rules for specified domains.
// Lists contain rules that define which domains should be blocked or allowed.
//
// Parameters:
//   - t: Testing context for error reporting
//   - lists: Repository for list operations
//   - rules: Repository for rule operations
//   - userID: ID of the user who owns this list
//   - domains: Slice of domain names to create blocking rules for
//
// Returns:
//   - int: The ID of the created list
func createTestList(t *testing.T, lists repos.ListRepo, rules repos.RuleRepo, domainRepo repos.DomainRepo, userID int, domains []string) int {
	list := &model.List{
		Name:        "Test Block List",                // Human-readable name for the list
		Description: "Test list for blocking domains", // Description of what this list blocks
		Source:      "test://source",                  // URL source (for external lists)
		UserID:      userID,                           // Links list to its owner
	}

	listID, err := lists.Create(list)
	if err != nil {
		t.Fatalf("Failed to create test list: %v", err)
	}

	// Add blocking rules for each domain in the list
	// Each rule defines whether a specific domain should be blocked or allowed
	for _, domain := range domains {
		// Create or get domain first
		domainModel := &model.Domain{Name: domain}
		domainID, err := domainRepo.CreateOrGet(domainModel)
		if err != nil {
			t.Fatalf("Failed to create domain %s: %v", domain, err)
		}

		rule := &model.Rule{
			DomainID: domainID, // The domain ID this rule applies to
			Allowed:  false,    // false = block, true = allow (whitelist)
		}
		ruleID, err := rules.CreateOrGet(rule)
		if err != nil {
			t.Fatalf("Failed to create blocking rule for domain %s: %v", domain, err)
		}

		// Link the rule to the list
		if err := rules.LinkToList(ruleID, listID); err != nil {
			t.Fatalf("Failed to link rule %d to list %d: %v", ruleID, listID, err)
		}
	}

	return listID
}

// createTestSchedule creates a test schedule that is currently active.
// Schedules define when blocking rules are applied based on time and day of week.
// This function creates a schedule that spans the current time to ensure blocking works.
//
// Parameters:
//   - t: Testing context for error reporting
//   - schedules: Repository for schedule operations
//   - devices: Repository for device operations
//   - deviceHash: Hash of the device this schedule applies to
//   - listIDs: IDs of lists to link to this schedule
//   - directDomains: Domains to block directly (without lists)
//   - userID: ID of the user who owns this schedule
//
// Returns:
//   - int: The ID of the created schedule
func createTestSchedule(t *testing.T, schedules repos.ScheduleRepo, rules repos.RuleRepo, domainRepo repos.DomainRepo, devices repos.DeviceRepo, deviceHash string, listIDs []int, directDomains []string, userID int) int {
	// Create a schedule that is active NOW for testing
	// This ensures that blocking rules will be applied during the test
	now := time.Now()

	schedule := &model.Schedule{
		Name:      "Test Schedule",              // Human-readable name
		UserID:    userID,                       // Links schedule to its owner
		StartTime: "00:00",                      // Start time in HH:MM format
		EndTime:   "23:55",                      // End time in HH:MM format
		Monday:    now.Weekday() == time.Monday, // Active on current day
		Tuesday:   now.Weekday() == time.Tuesday,
		Wednesday: now.Weekday() == time.Wednesday,
		Thursday:  now.Weekday() == time.Thursday,
		Friday:    now.Weekday() == time.Friday,
		Saturday:  now.Weekday() == time.Saturday,
		Sunday:    now.Weekday() == time.Sunday,
	}

	scheduleID, err := schedules.Create(schedule)
	if err != nil {
		t.Fatalf("Failed to create test schedule: %v", err)
	}

	// Link the target device to this schedule
	// Only devices linked to a schedule will have its rules applied
	// First get the device by hash to get its ID
	device, err := devices.FindByHash(deviceHash)
	if err != nil {
		t.Fatalf("Failed to find device by hash: %v", err)
	}
	if err := schedules.LinkDevice(scheduleID, device.ID); err != nil {
		t.Fatalf("Failed to link device to schedule: %v", err)
	}

	// Link blocking lists to this schedule
	// All rules in these lists will be applied during the schedule's active time
	for _, listID := range listIDs {
		if err := schedules.LinkList(scheduleID, listID); err != nil {
			t.Fatalf("Failed to link list %d to schedule: %v", listID, err)
		}
	}

	// Create rules for direct domains and link them to this schedule
	// These rules don't belong to any list - they're directly linked to the schedule
	for _, domain := range directDomains {
		// Create or get domain first
		domainModel := &model.Domain{Name: domain}
		domainID, err := domainRepo.CreateOrGet(domainModel)
		if err != nil {
			t.Fatalf("Failed to create domain %s: %v", domain, err)
		}

		// Create a rule for this domain (blocked by default with allowed=false)
		rule := &model.Rule{
			DomainID: domainID,
			Allowed:  false, // Blocked rule
		}
		ruleID, err := rules.CreateOrGet(rule)
		if err != nil {
			t.Fatalf("Failed to create rule for domain %s: %v", domain, err)
		}

		// Link the rule to the schedule
		if err := schedules.LinkRule(scheduleID, ruleID); err != nil {
			t.Fatalf("Failed to link rule %d to schedule: %v", ruleID, err)
		}
	}

	return scheduleID
}

// TestContentBlocker_IsBlocked tests the main domain blocking functionality.
// This test verifies that domains are correctly blocked or allowed based on:
// - Blocking rules in lists (via schedules)
// - Direct domain blocking (via schedules)
// - Unknown devices and domains
// - Invalid domain formats
func TestContentBlocker_IsBlocked(t *testing.T) {
	// Setup test database with clean schema
	db := setupTestDB(t)
	defer db.Close()

	// Create all required repositories for testing
	users := repos.NewUserRepo(db)
	devices := repos.NewDeviceRepo(db)
	lists := repos.NewListRepo(db)
	rules := repos.NewRuleRepo(db)
	schedules := repos.NewScheduleRepo(db)
	domains := repos.NewDomainRepo(db)

	// Create the content blocker instance we're testing
	contentBlocker := NewContentBlocker(devices, rules, schedules, domains)

	// Setup test data hierarchy:
	// User -> Device -> Schedule -> Lists/Rules -> Blocked Domains
	userID := createTestUser(t, users)
	deviceHash := createTestDevice(t, devices, userID)

	// Create a test list with several domains that should be blocked
	// These domains will be blocked via list rules
	blockedDomains := []string{"ads.example.com", "tracker.com", "malware.net", "statsig.anthropic.com", "evil.com"}
	listID := createTestList(t, lists, rules, domains, userID, blockedDomains)

	// Create domains that should be blocked directly (not via lists)
	// These are linked directly to the schedule
	directBlockedDomains := []string{"direct-blocked.com", "schedule-blocked.org"}

	// Create an active schedule that links everything together
	// This schedule is active NOW and applies to our test device
	createTestSchedule(t, schedules, rules, domains, devices, deviceHash, []int{listID}, directBlockedDomains, userID)

	// Initialize the content blocker with data from database
	if err := contentBlocker.Init(); err != nil {
		t.Fatalf("Failed to initialize content blocker: %v", err)
	}

	// Define test cases covering different blocking scenarios
	tests := []struct {
		name     string // Test case description
		domain   string // Domain to test for blocking
		device   string // Device hash to test with
		expected bool   // Expected blocking result (true = blocked, false = allowed)
		wantErr  bool   // Whether we expect an error
	}{
		{
			name:     "Block domain from list", // Should block: domain exists in our test list
			domain:   "ads.example.com",        // Domain from blockedDomains slice
			device:   deviceHash,               // Our test device
			expected: true,                     // Should be blocked
			wantErr:  false,                    // No error expected
		},
		{
			name:     "Block another domain from list", // Should block: another domain from list
			domain:   "tracker.com",                    // Another domain from blockedDomains slice
			device:   deviceHash,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "Block direct domain from schedule", // Should block: directly linked to schedule
			domain:   "direct-blocked.com",                // Domain from directBlockedDomains slice
			device:   deviceHash,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "Allow non-blocked domain", // Should allow: domain not in any blocking rules
			domain:   "google.com",               // Random domain not in our test data
			device:   deviceHash,
			expected: false, // Should NOT be blocked
			wantErr:  false,
		},
		{
			name:     "Allow domain for unknown device", // Should allow: device not in database
			domain:   "ads.example.com",                 // Even blocked domain should be allowed
			device:   "unknown-device-hash",             // Device that doesn't exist
			expected: false,                             // Should NOT be blocked
			wantErr:  false,
		},
		{
			name:     "Handle invalid domain format", // Should error: invalid domain format
			domain:   "invalid..domain",              // Domain with double dots (invalid)
			device:   deviceHash,
			expected: false, // Not blocked, but...
			wantErr:  true,  // Should return an error
		},
		{
			name:     "Handle empty domain", // Should error: empty domain string
			domain:   "",                    // Empty string
			device:   deviceHash,
			expected: false,
			wantErr:  true, // Should return an error
		},
		{
			name:     "Block subdomain statsig.anthropic.com", // Should block: exact match in list
			domain:   "statsig.anthropic.com",
			device:   deviceHash,
			expected: true, // Should be blocked
			wantErr:  false,
		},
		{
			name:     "Check parent domain anthropic.com", // Should NOT block: only subdomain is in list
			domain:   "anthropic.com",
			device:   deviceHash,
			expected: false, // Should NOT be blocked (only subdomain is blocked)
			wantErr:  false,
		},
		{
			name:     "Check unrelated subdomain api.anthropic.com", // Should NOT block: different subdomain
			domain:   "api.anthropic.com",
			device:   deviceHash,
			expected: false, // Should NOT be blocked (only statsig subdomain is blocked)
			wantErr:  false,
		},
		{
			name:     "Block parent domain evil.com", // Should block: exact match
			domain:   "evil.com",
			device:   deviceHash,
			expected: true, // Should be blocked
			wantErr:  false,
		},
		{
			name:     "Block subdomain of blocked parent sub.evil.com", // Should block: parent is blocked
			domain:   "sub.evil.com",
			device:   deviceHash,
			expected: true, // Should be blocked (parent domain is blocked)
			wantErr:  false,
		},
	}

	// Run each test case as a subtest
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function we're testing
			blocked, err := contentBlocker.IsBlocked(tt.domain, tt.device)

			// Check error handling first
			if tt.wantErr {
				if err == nil {
					t.Errorf("IsBlocked() expected error but got none")
				}
				return // Don't check blocking result if we expected an error
			}

			// If we didn't expect an error, make sure we didn't get one
			if err != nil {
				t.Errorf("IsBlocked() unexpected error: %v", err)
				return
			}

			// Check that the blocking result matches expectations
			if blocked != tt.expected {
				t.Errorf("IsBlocked() = %v, expected %v for domain %s", blocked, tt.expected, tt.domain)
			}
		})
	}
}

// TestContentBlocker_TimeBasedBlocking tests that blocking only occurs during active schedule times.
// This test verifies that schedules with inactive time windows don't block domains.
func TestContentBlocker_TimeBasedBlocking(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	// Create repositories
	users := repos.NewUserRepo(db)
	devices := repos.NewDeviceRepo(db)
	lists := repos.NewListRepo(db)
	rules := repos.NewRuleRepo(db)
	schedules := repos.NewScheduleRepo(db)
	domains := repos.NewDomainRepo(db)

	// Create content blocker
	contentBlocker := NewContentBlocker(devices, rules, schedules, domains)

	// Setup test data
	userID := createTestUser(t, users)
	deviceHash := createTestDevice(t, devices, userID)

	// Create test list
	blockedDomains := []string{"time-blocked.com"}
	listID := createTestList(t, lists, rules, domains, userID, blockedDomains)

	// Create schedule that is NOT currently active (yesterday's day pattern)
	yesterday := time.Now().Add(-24 * time.Hour)

	// Round times to 5-minute boundaries to satisfy SQL constraints
	startMinute := (yesterday.Minute() / 5) * 5
	endTime := yesterday.Add(1 * time.Hour)
	endMinute := (endTime.Minute() / 5) * 5
	yesterdayStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), yesterday.Hour(), startMinute, 0, 0, yesterday.Location())
	yesterdayEnd := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), endTime.Hour(), endMinute, 0, 0, endTime.Location())

	schedule := &model.Schedule{
		Name:      "Inactive Schedule",
		UserID:    userID,
		StartTime: yesterdayStart.Format("15:04"),     // Use yesterday's time
		EndTime:   yesterdayEnd.Format("15:04"),       // End time 1 hour later
		Monday:    yesterday.Weekday() == time.Monday, // Active only on yesterday's day
		Tuesday:   yesterday.Weekday() == time.Tuesday,
		Wednesday: yesterday.Weekday() == time.Wednesday,
		Thursday:  yesterday.Weekday() == time.Thursday,
		Friday:    yesterday.Weekday() == time.Friday,
		Saturday:  yesterday.Weekday() == time.Saturday,
		Sunday:    yesterday.Weekday() == time.Sunday,
	}

	scheduleID, err := schedules.Create(schedule)
	if err != nil {
		t.Fatalf("Failed to create inactive schedule: %v", err)
	}

	// Link device and list to inactive schedule
	// First get the device by hash to get its ID
	device, err := devices.FindByHash(deviceHash)
	if err != nil {
		t.Fatalf("Failed to find device by hash: %v", err)
	}
	if err := schedules.LinkDevice(scheduleID, device.ID); err != nil {
		t.Fatalf("Failed to link device to inactive schedule: %v", err)
	}

	if err := schedules.LinkList(scheduleID, listID); err != nil {
		t.Fatalf("Failed to link list to inactive schedule: %v", err)
	}

	// Initialize the content blocker
	if err := contentBlocker.Init(); err != nil {
		t.Fatalf("Failed to initialize content blocker: %v", err)
	}

	// Test that domain is NOT blocked when schedule is inactive
	blocked, err := contentBlocker.IsBlocked("time-blocked.com", deviceHash)
	if err != nil {
		t.Fatalf("IsBlocked() unexpected error: %v", err)
	}

	if blocked {
		t.Errorf("IsBlocked() = true, expected false for inactive schedule")
	}
}

// TestContentBlocker_DomainNormalization tests domain preprocessing and validation.
// This test verifies that domains are properly normalized and invalid formats are caught.
func TestContentBlocker_DomainNormalization(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	schedules := repos.NewScheduleRepo(db)
	devices := repos.NewDeviceRepo(db)
	rules := repos.NewRuleRepo(db)
	domains := repos.NewDomainRepo(db)
	contentBlocker := NewContentBlocker(devices, rules, schedules, domains)

	tests := []struct {
		name        string
		inputDomain string
		wantErr     bool
	}{
		{
			name:        "Domain with trailing dot",
			inputDomain: "example.com.",
			wantErr:     false,
		},
		{
			name:        "Domain with .localdomain suffix",
			inputDomain: "myserver.example.com.localdomain",
			wantErr:     false, // After stripping .localdomain, becomes "myserver.example.com" which is valid
		},
		{
			name:        "Mixed case domain",
			inputDomain: "ExAmPlE.CoM",
			wantErr:     false,
		},
		{
			name:        "Invalid domain with double dots",
			inputDomain: "invalid..domain.com",
			wantErr:     true,
		},
		{
			name:        "Invalid domain - just dots",
			inputDomain: "...",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := contentBlocker.IsBlocked(tt.inputDomain, "non-existent-device")

			if tt.wantErr {
				if err == nil {
					t.Errorf("IsBlocked() expected error for invalid domain %s but got none", tt.inputDomain)
				}
			} else {
				if err != nil {
					t.Errorf("IsBlocked() unexpected error for valid domain %s: %v", tt.inputDomain, err)
				}
			}
		})
	}
}

// TestContentBlocker_SubdomainBlocking tests that blocking a parent domain also blocks its subdomains.
// This verifies the hierarchical domain blocking feature using the radix tree's longest-prefix matching.
func TestContentBlocker_SubdomainBlocking(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	// Create repositories
	users := repos.NewUserRepo(db)
	devices := repos.NewDeviceRepo(db)
	lists := repos.NewListRepo(db)
	rules := repos.NewRuleRepo(db)
	schedules := repos.NewScheduleRepo(db)
	domains := repos.NewDomainRepo(db)

	// Create content blocker
	contentBlocker := NewContentBlocker(devices, rules, schedules, domains)

	// Setup test data
	userID := createTestUser(t, users)
	deviceHash := createTestDevice(t, devices, userID)

	// Create a test list that blocks the parent domain "google.com"
	blockedDomains := []string{"google.com"}
	listID := createTestList(t, lists, rules, domains, userID, blockedDomains)

	// Create an active schedule
	createTestSchedule(t, schedules, rules, domains, devices, deviceHash, []int{listID}, []string{}, userID)

	// Initialize the content blocker
	if err := contentBlocker.Init(); err != nil {
		t.Fatalf("Failed to initialize content blocker: %v", err)
	}

	// Test cases: parent domain should block all subdomains
	tests := []struct {
		name     string
		domain   string
		expected bool
	}{
		{
			name:     "Block parent domain",
			domain:   "google.com",
			expected: true,
		},
		{
			name:     "Block first-level subdomain",
			domain:   "mail.google.com",
			expected: true,
		},
		{
			name:     "Block multi-level subdomain",
			domain:   "accounts.mail.google.com",
			expected: true,
		},
		{
			name:     "Allow unrelated domain",
			domain:   "example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked, err := contentBlocker.IsBlocked(tt.domain, deviceHash)
			if err != nil {
				t.Fatalf("IsBlocked() unexpected error: %v", err)
			}

			if blocked != tt.expected {
				t.Errorf("IsBlocked() = %v, expected %v for domain %s", blocked, tt.expected, tt.domain)
			}
		})
	}
}

// TestContentBlocker_SubdomainVsParent tests that blocking a subdomain does NOT block its parent domain.
// This verifies that the blocking works only in the downward direction (parent → subdomain).
func TestContentBlocker_SubdomainVsParent(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	// Create repositories
	users := repos.NewUserRepo(db)
	devices := repos.NewDeviceRepo(db)
	lists := repos.NewListRepo(db)
	rules := repos.NewRuleRepo(db)
	schedules := repos.NewScheduleRepo(db)
	domains := repos.NewDomainRepo(db)

	// Create content blocker
	contentBlocker := NewContentBlocker(devices, rules, schedules, domains)

	// Setup test data
	userID := createTestUser(t, users)
	deviceHash := createTestDevice(t, devices, userID)

	// Create a test list that blocks only the subdomain "ads.example.com"
	blockedDomains := []string{"ads.example.com"}
	listID := createTestList(t, lists, rules, domains, userID, blockedDomains)

	// Create an active schedule
	createTestSchedule(t, schedules, rules, domains, devices, deviceHash, []int{listID}, []string{}, userID)

	// Initialize the content blocker
	if err := contentBlocker.Init(); err != nil {
		t.Fatalf("Failed to initialize content blocker: %v", err)
	}

	// Test cases: subdomain blocking should NOT affect parent domains
	tests := []struct {
		name     string
		domain   string
		expected bool
	}{
		{
			name:     "Block the specific subdomain",
			domain:   "ads.example.com",
			expected: true,
		},
		{
			name:     "Allow parent domain",
			domain:   "example.com",
			expected: false,
		},
		{
			name:     "Allow other subdomains",
			domain:   "www.example.com",
			expected: false,
		},
		{
			name:     "Allow nested subdomains under blocked domain",
			domain:   "tracking.ads.example.com",
			expected: true, // This should be blocked as it's under ads.example.com
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked, err := contentBlocker.IsBlocked(tt.domain, deviceHash)
			if err != nil {
				t.Fatalf("IsBlocked() unexpected error: %v", err)
			}

			if blocked != tt.expected {
				t.Errorf("IsBlocked() = %v, expected %v for domain %s", blocked, tt.expected, tt.domain)
			}
		})
	}
}

// TestContentBlocker_RealWorldData tests the content blocker with real-world data similar to test/test.go
// This test mimics the production scenario with multiple devices, lists, and schedules.
func TestContentBlocker_RealWorldData(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	// Create repositories
	users := repos.NewUserRepo(db)
	devices := repos.NewDeviceRepo(db)
	lists := repos.NewListRepo(db)
	rules := repos.NewRuleRepo(db)
	schedules := repos.NewScheduleRepo(db)
	domains := repos.NewDomainRepo(db)

	// Create content blocker
	contentBlocker := NewContentBlocker(devices, rules, schedules, domains)

	// Create user - matching test/test.go
	user := &model.User{
		Name:         "Arian Gohari",
		Email:        "arian@gohari.de",
		PasswordHash: "hashed-password",
	}
	if err := users.Create(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	userID := user.ID

	// Create devices - matching test/test.go
	device1 := &model.Device{
		Hash:   "iphone-test-hash",
		Name:   "IPhone von Arian",
		OS:     model.IOS,
		UserID: userID,
	}
	if err := devices.Create(device1); err != nil {
		t.Fatalf("Failed to create device1: %v", err)
	}

	device2 := &model.Device{
		Hash:   "macbook-test-hash", 
		Name:   "MacBook Pro von Arian",
		OS:     model.MacOS,
		UserID: userID,
	}
	if err := devices.Create(device2); err != nil {
		t.Fatalf("Failed to create device2: %v", err)
	}

	// Create lists with some sample blocked domains (simplified from test/test.go)
	// List 1: Ad blocking
	list1 := &model.List{
		Name:        "Ad Blocker",
		Description: "Blocks ads and trackers",
		Source:      "test://adblocker",
		UserID:      userID,
	}
	list1ID, err := lists.Create(list1)
	if err != nil {
		t.Fatalf("Failed to create list1: %v", err)
	}

	// Add some ad domains to list1
	adDomains := []string{"doubleclick.net", "googleadservices.com", "googlesyndication.com"}
	for _, domain := range adDomains {
		domainModel := &model.Domain{Name: domain}
		domainID, err := domains.CreateOrGet(domainModel)
		if err != nil {
			t.Fatalf("Failed to create domain %s: %v", domain, err)
		}

		rule := &model.Rule{
			DomainID: domainID,
			Allowed:  false,
		}
		ruleID, err := rules.CreateOrGet(rule)
		if err != nil {
			t.Fatalf("Failed to create rule for %s: %v", domain, err)
		}

		if err := rules.LinkToList(ruleID, list1ID); err != nil {
			t.Fatalf("Failed to link rule to list1: %v", err)
		}
	}

	// List 2: Malware blocking  
	list2 := &model.List{
		Name:        "Malware Blocker",
		Description: "Blocks malware and phishing",
		Source:      "test://malware",
		UserID:      userID,
	}
	list2ID, err := lists.Create(list2)
	if err != nil {
		t.Fatalf("Failed to create list2: %v", err)
	}

	// Add direct blocking rules (like in test/test.go)
	directDomains := []string{"example.com", "test.com"}
	var directRuleIDs []int
	for _, domain := range directDomains {
		domainModel := &model.Domain{Name: domain}
		domainID, err := domains.CreateOrGet(domainModel)
		if err != nil {
			t.Fatalf("Failed to create domain %s: %v", domain, err)
		}

		rule := &model.Rule{
			DomainID: domainID,
			Allowed:  false,
		}
		ruleID, err := rules.CreateOrGet(rule)
		if err != nil {
			t.Fatalf("Failed to create rule for %s: %v", domain, err)
		}
		directRuleIDs = append(directRuleIDs, ruleID)
	}

	// Create schedule - Base Protection (00:00-23:55, all days) like in test/test.go
	schedule := &model.Schedule{
		Name:      "Base Protection",
		UserID:    userID,
		StartTime: "00:00",
		EndTime:   "23:55",
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  true,
		Sunday:    true,
	}

	scheduleID, err := schedules.Create(schedule)
	if err != nil {
		t.Fatalf("Failed to create schedule: %v", err)
	}

	// Link devices to schedule
	if err := schedules.LinkDevice(scheduleID, device1.ID); err != nil {
		t.Fatalf("Failed to link device1 to schedule: %v", err)
	}
	if err := schedules.LinkDevice(scheduleID, device2.ID); err != nil {
		t.Fatalf("Failed to link device2 to schedule: %v", err)
	}

	// Link lists to schedule
	if err := schedules.LinkList(scheduleID, list1ID); err != nil {
		t.Fatalf("Failed to link list1 to schedule: %v", err)
	}
	if err := schedules.LinkList(scheduleID, list2ID); err != nil {
		t.Fatalf("Failed to link list2 to schedule: %v", err)
	}

	// Link direct rules to schedule
	for _, ruleID := range directRuleIDs {
		if err := schedules.LinkRule(scheduleID, ruleID); err != nil {
			t.Fatalf("Failed to link rule %d to schedule: %v", ruleID, err)
		}
	}

	// Initialize the content blocker
	if err := contentBlocker.Init(); err != nil {
		t.Fatalf("Failed to initialize content blocker: %v", err)
	}

	// Test cases
	tests := []struct {
		name       string
		domain     string
		deviceHash string
		expected   bool
	}{
		// Test direct blocked domains
		{
			name:       "Block example.com on iPhone",
			domain:     "example.com",
			deviceHash: "iphone-test-hash",
			expected:   true,
		},
		{
			name:       "Block test.com on MacBook",
			domain:     "test.com",
			deviceHash: "macbook-test-hash",
			expected:   true,
		},
		// Test list blocked domains
		{
			name:       "Block ad domain on iPhone",
			domain:     "doubleclick.net",
			deviceHash: "iphone-test-hash",
			expected:   true,
		},
		{
			name:       "Block ad subdomain on MacBook",
			domain:     "stats.doubleclick.net",
			deviceHash: "macbook-test-hash",
			expected:   true,
		},
		// Test allowed domains
		{
			name:       "Allow google.com on iPhone",
			domain:     "google.com",
			deviceHash: "iphone-test-hash",
			expected:   false,
		},
		{
			name:       "Allow github.com on MacBook",
			domain:     "github.com",
			deviceHash: "macbook-test-hash",
			expected:   false,
		},
		// Test unknown device
		{
			name:       "Unknown device should not block",
			domain:     "example.com",
			deviceHash: "unknown-device-hash",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked, err := contentBlocker.IsBlocked(tt.domain, tt.deviceHash)
			if err != nil {
				t.Fatalf("IsBlocked() unexpected error: %v", err)
			}

			if blocked != tt.expected {
				t.Errorf("IsBlocked() = %v, expected %v for domain %s with device %s", 
					blocked, tt.expected, tt.domain, tt.deviceHash)
			}
		})
	}
}
