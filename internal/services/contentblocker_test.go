// Package services provides unit tests for the content blocking functionality.
// These tests verify that domains are properly blocked based on schedules, lists, and rules.
package services

import (
	"database/sql"
	"testing"
	"time"

	schema "github.com/lemon3studio/leo/internal/db"
	"github.com/lemon3studio/leo/internal/model"
	"github.com/lemon3studio/leo/internal/repos"

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
//   - userRepo: Repository for user operations
//
// Returns:
//   - string: The hash/ID of the created user
func createTestUser(t *testing.T, userRepo repos.UserRepo) string {
	user := &model.User{
		Hash:         "test-user-hash",   // Fixed hash for predictable testing
		Name:         "Test User",        // Human-readable name
		Email:        "test@example.com", // Valid email format
		PasswordHash: "hashed-password",  // Pre-hashed password (not used in blocking tests)
	}

	if err := userRepo.Create(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user.Hash
}

// createTestDevice creates a test device in the database.
// Devices are the targets for DNS blocking - each device can be linked to schedules.
//
// Parameters:
//   - t: Testing context for error reporting
//   - deviceRepo: Repository for device operations
//   - userHash: Hash of the user who owns this device
//
// Returns:
//   - string: The hash/ID of the created device
func createTestDevice(t *testing.T, deviceRepo repos.DeviceRepo, userHash string) string {
	device := &model.Device{
		Hash:     "test-device-hash", // Fixed hash matching the one used in mobileconfig URLs
		Name:     "Test Device",      // Human-readable name for the device
		OS:       model.IOS,          // Device operating system (iOS/Android/etc.)
		UserHash: userHash,           // Links device to its owner
	}

	if err := deviceRepo.Create(device); err != nil {
		t.Fatalf("Failed to create test device: %v", err)
	}

	return device.Hash
}

// createTestList creates a test list with blocking rules for specified domains.
// Lists contain rules that define which domains should be blocked or allowed.
//
// Parameters:
//   - t: Testing context for error reporting
//   - listRepo: Repository for list operations
//   - ruleRepo: Repository for rule operations
//   - userHash: Hash of the user who owns this list
//   - domains: Slice of domain names to create blocking rules for
//
// Returns:
//   - int: The ID of the created list
func createTestList(t *testing.T, listRepo repos.ListRepo, ruleRepo repos.RuleRepo, userHash string, domains []string) int {
	list := &model.List{
		Name:        "Test Block List",                // Human-readable name for the list
		Description: "Test list for blocking domains", // Description of what this list blocks
		Source:      "test://source",                  // URL source (for external lists)
		UserHash:    userHash,                         // Links list to its owner
	}

	listID, err := listRepo.Create(list)
	if err != nil {
		t.Fatalf("Failed to create test list: %v", err)
	}

	// Add blocking rules for each domain in the list
	// Each rule defines whether a specific domain should be blocked or allowed
	for _, domain := range domains {
		rule := &model.Rule{
			Domain:  domain, // The domain this rule applies to
			Allowed: false,  // false = block, true = allow (whitelist)
		}
		ruleID, err := ruleRepo.CreateOrGet(rule)
		if err != nil {
			t.Fatalf("Failed to create blocking rule for domain %s: %v", domain, err)
		}

		// Link the rule to the list
		if err := ruleRepo.LinkToList(ruleID, listID); err != nil {
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
//   - scheduleRepo: Repository for schedule operations
//   - deviceHash: Hash of the device this schedule applies to
//   - listIDs: IDs of lists to link to this schedule
//   - directDomains: Domains to block directly (without lists)
//   - userHash: Hash of the user who owns this schedule
//
// Returns:
//   - int: The ID of the created schedule
func createTestSchedule(t *testing.T, scheduleRepo repos.ScheduleRepo, ruleRepo repos.RuleRepo, deviceHash string, listIDs []int, directDomains []string, userHash string) int {
	// Create a schedule that is active NOW for testing
	// This ensures that blocking rules will be applied during the test
	now := time.Now()
	startTime := now.Add(-1 * time.Hour).Format("15:04") // Started 1 hour ago
	endTime := now.Add(1 * time.Hour).Format("15:04")    // Ends in 1 hour

	schedule := &model.Schedule{
		Name:      "Test Schedule", // Human-readable name
		StartTime: startTime,       // Daily start time (HH:MM format)
		EndTime:   endTime,         // Daily end time (HH:MM format)
		UserHash:  userHash,        // Links schedule to its owner
		// Set only the current day to true so the schedule is active today
		// Days are stored as boolean flags for each day of the week
		Monday:    now.Weekday() == time.Monday,
		Tuesday:   now.Weekday() == time.Tuesday,
		Wednesday: now.Weekday() == time.Wednesday,
		Thursday:  now.Weekday() == time.Thursday,
		Friday:    now.Weekday() == time.Friday,
		Saturday:  now.Weekday() == time.Saturday,
		Sunday:    now.Weekday() == time.Sunday,
	}

	scheduleID, err := scheduleRepo.Create(schedule)
	if err != nil {
		t.Fatalf("Failed to create test schedule: %v", err)
	}

	// Link the target device to this schedule
	// Only devices linked to a schedule will have its rules applied
	if err := scheduleRepo.LinkDevice(scheduleID, deviceHash); err != nil {
		t.Fatalf("Failed to link device to schedule: %v", err)
	}

	// Link blocking lists to this schedule
	// All rules in these lists will be applied during the schedule's active time
	for _, listID := range listIDs {
		if err := scheduleRepo.LinkList(scheduleID, listID); err != nil {
			t.Fatalf("Failed to link list %d to schedule: %v", listID, err)
		}
	}

	// Create rules for direct domains and link them to this schedule
	// These rules don't belong to any list - they're directly linked to the schedule
	for _, domain := range directDomains {
		// Create a rule for this domain (blocked by default with allowed=false)
		rule := &model.Rule{
			Domain: domain,
			// ListID is 0/null - this rule doesn't belong to any list
			Allowed: false, // Blocked rule
		}
		ruleID, err := ruleRepo.CreateOrGet(rule)
		if err != nil {
			t.Fatalf("Failed to create rule for domain %s: %v", domain, err)
		}

		// Link the rule to the schedule
		if err := scheduleRepo.LinkRule(scheduleID, ruleID); err != nil {
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
	userRepo := repos.NewUserRepo(db)
	deviceRepo := repos.NewDeviceRepo(db)
	listRepo := repos.NewListRepo(db)
	ruleRepo := repos.NewRuleRepo(db)
	scheduleRepo := repos.NewScheduleRepo(db)

	// Create the content blocker instance we're testing
	contentBlocker := NewContentBlocker(scheduleRepo)

	// Setup test data hierarchy:
	// User -> Device -> Schedule -> Lists/Rules -> Blocked Domains
	userHash := createTestUser(t, userRepo)
	deviceHash := createTestDevice(t, deviceRepo, userHash)

	// Create a test list with several domains that should be blocked
	// These domains will be blocked via list rules
	blockedDomains := []string{"ads.example.com", "tracker.com", "malware.net"}
	listID := createTestList(t, listRepo, ruleRepo, userHash, blockedDomains)

	// Create domains that should be blocked directly (not via lists)
	// These are linked directly to the schedule
	directBlockedDomains := []string{"direct-blocked.com", "schedule-blocked.org"}

	// Create an active schedule that links everything together
	// This schedule is active NOW and applies to our test device
	createTestSchedule(t, scheduleRepo, ruleRepo, deviceHash, []int{listID}, directBlockedDomains, userHash)

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
	userRepo := repos.NewUserRepo(db)
	deviceRepo := repos.NewDeviceRepo(db)
	listRepo := repos.NewListRepo(db)
	ruleRepo := repos.NewRuleRepo(db)
	scheduleRepo := repos.NewScheduleRepo(db)

	// Create content blocker
	contentBlocker := NewContentBlocker(scheduleRepo)

	// Setup test data
	userHash := createTestUser(t, userRepo)
	deviceHash := createTestDevice(t, deviceRepo, userHash)

	// Create test list
	blockedDomains := []string{"time-blocked.com"}
	listID := createTestList(t, listRepo, ruleRepo, userHash, blockedDomains)

	// Create schedule that is NOT currently active (yesterday's time window)
	yesterday := time.Now().Add(-24 * time.Hour)
	startTime := yesterday.Format("15:04")
	endTime := yesterday.Add(1 * time.Hour).Format("15:04")

	schedule := &model.Schedule{
		Name:      "Inactive Schedule",
		StartTime: startTime,
		EndTime:   endTime,
		UserHash:  userHash,
		// Set current day to false, all other days to true
		Monday:    yesterday.Weekday() == time.Monday,
		Tuesday:   yesterday.Weekday() == time.Tuesday,
		Wednesday: yesterday.Weekday() == time.Wednesday,
		Thursday:  yesterday.Weekday() == time.Thursday,
		Friday:    yesterday.Weekday() == time.Friday,
		Saturday:  yesterday.Weekday() == time.Saturday,
		Sunday:    yesterday.Weekday() == time.Sunday,
	}

	scheduleID, err := scheduleRepo.Create(schedule)
	if err != nil {
		t.Fatalf("Failed to create inactive schedule: %v", err)
	}

	// Link device and list to inactive schedule
	if err := scheduleRepo.LinkDevice(scheduleID, deviceHash); err != nil {
		t.Fatalf("Failed to link device to inactive schedule: %v", err)
	}

	if err := scheduleRepo.LinkList(scheduleID, listID); err != nil {
		t.Fatalf("Failed to link list to inactive schedule: %v", err)
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

	scheduleRepo := repos.NewScheduleRepo(db)
	contentBlocker := NewContentBlocker(scheduleRepo)

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
