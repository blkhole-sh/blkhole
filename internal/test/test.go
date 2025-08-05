// Package test provides testing utilities and fixtures for the Leo DNS blocker application.
// This package contains helper functions for creating test data and a comprehensive
// integration test that sets up a realistic blocking scenario.
package test

import (
	"github.com/lemon3studio/leo/internal/model"
	"github.com/lemon3studio/leo/internal/repos"
	"github.com/lemon3studio/leo/internal/services"
)

// Test interface defines methods for creating test data fixtures.
// This interface provides a clean API for setting up test scenarios with users,
// devices, lists, rules, and schedules.
type Test interface {
	// AddUser creates a new user with the given password and returns the user hash
	AddUser(user model.User, password string) (string, error)

	// AddDevice creates a new device and returns the device hash
	AddDevice(device model.Device) (string, error)

	// AddRule creates a new blocking/allowing rule and returns the rule ID
	AddRule(rule model.Rule) (int, error)

	// AddList creates a new list with blocking rules for the given domains and returns the list ID
	AddList(list model.List, domains []string) (int, error)

	// AddSchedule creates a new schedule with all its relationships and returns the schedule ID
	AddSchedule(schedule model.Schedule) (int, error)

	// Test creates a complete test scenario with user, devices, lists, and schedules
	// Returns the user hash for the created test scenario
	Test() (string, error)
}

// TestImpl implements the Test interface and provides concrete methods for creating test data.
// It holds references to all the repositories and services needed to create a complete
// test scenario with proper relationships between users, devices, lists, and schedules.
type TestImpl struct {
	userRepo      repos.UserRepo         // Repository for user operations
	deviceRepo    repos.DeviceRepo       // Repository for device operations
	ruleRepo      repos.RuleRepo         // Repository for rule operations
	listRepo      repos.ListRepo         // Repository for list operations
	listService   services.ListsService  // Service for list loading operations
	scheduleRepo  repos.ScheduleRepo     // Repository for schedule operations
	cryptoService services.CryptoService // Service for generating hashes and passwords
}

// NewTest creates a new TestImpl instance with all required dependencies.
// This constructor function initializes the test helper with access to all
// repositories and services needed to create comprehensive test scenarios.
//
// Parameters:
//   - All repository and service dependencies needed for test data creation
//
// Returns:
//   - Test: Interface for creating test data fixtures
func NewTest(userRepo repos.UserRepo, deviceRepo repos.DeviceRepo, ruleRepo repos.RuleRepo, listRepo repos.ListRepo, listService services.ListsService, scheduleRepo repos.ScheduleRepo, cryptoService services.CryptoService) Test {
	return &TestImpl{
		userRepo:      userRepo,
		deviceRepo:    deviceRepo,
		ruleRepo:      ruleRepo,
		listRepo:      listRepo,
		listService:   listService,
		scheduleRepo:  scheduleRepo,
		cryptoService: cryptoService,
	}
}

// AddUser creates a new user in the database with a hashed password.
// This method generates a random hash for the user ID and securely hashes the password.
//
// Parameters:
//   - user: User model with name and email (hash and password will be generated)
//   - password: Plain text password to be hashed
//
// Returns:
//   - string: The generated user hash (used as the user's unique identifier)
//   - error: Any error that occurred during user creation
func (t TestImpl) AddUser(user model.User, password string) (string, error) {
	// Generate a cryptographically secure random hash for the user
	hash, err := t.cryptoService.RandomHash()
	if err != nil {
		return "", err
	}

	// Hash the password using a secure hashing algorithm (bcrypt)
	passwordHash, err := t.cryptoService.HashPassword(password)
	if err != nil {
		return "", err
	}

	// Set the generated hash and hashed password
	user.Hash = hash
	user.PasswordHash = passwordHash

	// Store the user in the database
	err = t.userRepo.Create(&user)
	if err != nil {
		return "", err
	}

	return user.Hash, nil
}

// AddDevice creates a new device in the database.
// Devices represent endpoints (phones, computers, etc.) that will use DNS blocking.
// Each device gets a unique hash that's used in mobileconfig URLs for DNS-over-HTTPS.
//
// Parameters:
//   - device: Device model with name, OS, and user hash (device hash will be generated)
//
// Returns:
//   - string: The generated device hash (used in DNS queries and mobileconfig URLs)
//   - error: Any error that occurred during device creation
func (t TestImpl) AddDevice(device model.Device) (string, error) {
	// Generate a cryptographically secure random hash for the device
	// This hash will be used in the mobileconfig URL: /devices/{hash}/config
	hash, err := t.cryptoService.RandomHash()
	if err != nil {
		return "", err
	}

	// Set the generated hash
	device.Hash = hash

	// Store the device in the database
	err = t.deviceRepo.Create(&device)
	if err != nil {
		return "", err
	}

	return device.Hash, nil
}

// AddRule creates a new blocking or allowing rule in the database.
// Rules define whether specific domains should be blocked (Allowed=false) or whitelisted (Allowed=true).
// Rules belong to lists and are applied when the list is linked to an active schedule.
//
// Parameters:
//   - rule: Rule model with domain, list ID, and allowed flag
//
// Returns:
//   - int: The ID of the created rule
//   - error: Any error that occurred during rule creation
func (t TestImpl) AddRule(rule model.Rule) (int, error) {
	// Create the rule in the database
	id, err := t.ruleRepo.Create(&rule)
	if err != nil {
		return -1, err
	}

	return id, nil
}

// AddList creates a new list with blocking rules for the specified domains.
// Lists are collections of rules that can be applied to schedules. This method
// creates a list and automatically adds blocking rules for all provided domains.
//
// Parameters:
//   - list: List model with name, description, source, and user hash
//   - domains: Slice of domain names to create blocking rules for
//
// Returns:
//   - int: The ID of the created list
//   - error: Any error that occurred during list or rule creation
func (t TestImpl) AddList(list model.List, domains []string) (int, error) {
	// Create the list in the database
	id, err := t.listRepo.Create(&list)
	if err != nil {
		return -1, err
	}

	// Create blocking rules for each domain in the list
	// By default, all rules are set to block (Allowed=false)
	for _, domain := range domains {
		_, err := t.AddRule(model.Rule{
			Domain:  domain, // The domain to block
			ListID:  id,     // Link to the parent list
			Allowed: false,  // Block this domain (false = block, true = allow/whitelist)
		})
		if err != nil {
			return id, err
		}
	}

	// Load the list into the list service (for caching/performance)
	t.listService.LoadList(&list)

	return id, nil
}

// AddSchedule creates a new schedule with all its relationships.
// Schedules define when and how blocking rules are applied. They specify:
// - Time windows (start/end times and days of week)
// - Which devices the rules apply to
// - Which lists of rules to apply
// - Which domains to block directly
//
// Parameters:
//   - schedule: Schedule model with time settings, device hashes, domains, and list IDs
//
// Returns:
//   - int: The ID of the created schedule
//   - error: Any error that occurred during schedule creation or linking
func (t TestImpl) AddSchedule(schedule model.Schedule) (int, error) {
	// Create the schedule in the database
	id, err := t.scheduleRepo.Create(&schedule)
	if err != nil {
		return -1, err
	}

	// Link all specified devices to this schedule
	// Only devices linked to a schedule will have its rules applied
	for _, dh := range schedule.DeviceHashes {
		err = t.scheduleRepo.LinkDevice(id, dh)
		if err != nil {
			return id, err
		}
	}

	// Link direct domains to this schedule
	// These domains are blocked directly without needing list rules
	for _, domain := range schedule.Domains {
		err = t.scheduleRepo.LinkDomain(id, domain)
		if err != nil {
			return id, err
		}
	}

	// Link lists to this schedule
	// All rules in these lists will be applied during the schedule's active time
	for _, li := range schedule.ListIds {
		err = t.scheduleRepo.LinkList(id, li)
		if err != nil {
			return id, err
		}
	}

	return id, nil
}

// Test creates a comprehensive integration test scenario with realistic data.
// This method sets up a complete blocking environment including:
// - A test user (Arian Gohari)
// - Two devices (iPhone and MacBook)
// - Three popular blocking lists (Hagezi, OISD, Anti Axel Springer)
// - An active schedule that applies blocking rules
//
// This is useful for integration testing and demonstrating the full system functionality.
// The created schedule blocks domains Monday-Wednesday, Friday-Sunday from 09:00-22:00.
// Thursday is intentionally left unblocked for testing time-based rules.
//
// Returns:
//   - string: The user hash for the created test scenario
//   - error: Any error that occurred during test setup
func (t TestImpl) Test() (string, error) {
	// Create a test user to own all the test data
	uh, err := t.AddUser(model.User{
		Name:  "Arian Gohari",    // Test user name
		Email: "arian@gohari.de", // Test user email
	}, "SomePassword") // Test password (will be hashed)
	if err != nil {
		return "", err
	}

	// Create two test devices representing different platforms
	// Device 1: iPhone (iOS device)
	d1h, err := t.AddDevice(model.Device{
		Name:     "IPhone von Arian", // German naming convention
		OS:       model.IOS,          // iOS operating system
		UserHash: uh,                 // Links device to the test user
	})
	if err != nil {
		return "", err
	}

	// Device 2: MacBook (macOS device)
	d2h, err := t.AddDevice(model.Device{
		Name:     "MacBook Pro von Arian", // German naming convention
		OS:       model.MacOS,             // macOS operating system
		UserHash: uh,                      // Links device to the test user
	})
	if err != nil {
		return "", err
	}

	// Create blocking lists based on popular real-world DNS blocklists
	// List 1: Hagezi Multi ULTIMATE - Comprehensive ad/tracking blocker
	l1i, err := t.AddList(model.List{
		Name:        "Hagezi Multi ULTIMATE",
		Description: "Ultimate Sweeper - Strictly cleans the Internet and protects your privacy! Blocks Ads, Affiliate, Tracking, Metrics, Telemetry, Phishing, Malware, Scam, Free Hoster, Fake, Cryptojacking and other Crap.",
		Source:      "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/adblock/pro.plus.mini.txt",
		UserHash:    uh,
	}, []string{}) // Empty domains array - would be populated from the source URL in production
	if err != nil {
		return "", err
	}

	// List 2: OISD Big - Another popular comprehensive blocker
	l2i, err := t.AddList(model.List{
		Name:        "OISD Big",
		Description: "Blocks Ads, (Mobile) App Ads, Phishing, Malvertising, Malware, Spyware, Ransomware, CryptoJacking, Telemetry/Analytics/Tracking (where not needed for proper functionality).",
		Source:      "https://big.oisd.nl",
		UserHash:    uh,
	}, []string{}) // Empty domains array - would be populated from the source URL in production
	if err != nil {
		return "", err
	}

	// List 3: Anti Axel Springer - Blocks specific media company domains
	l3i, err := t.AddList(model.List{
		Name:        "Anti Axel Springer",
		Description: "This list blocks all connections to sites which are from Axel Springer Verlag or have a connection with them.",
		Source:      "https://raw.githubusercontent.com/autinerd/anti-axelspringer-hosts/master/axelspringer-hosts",
		UserHash:    uh,
	}, []string{}) // Empty domains array - would be populated from the source URL in production
	if err != nil {
		return "", err
	}

	// Create an active schedule that applies all blocking lists to both devices
	// This schedule is active most days but leaves Thursday unblocked for testing
	_, err = t.AddSchedule(model.Schedule{
		Name:      "Base Protection", // Human-readable schedule name
		StartTime: "09:00",           // Daily start time (9 AM)
		EndTime:   "22:00",           // Daily end time (10 PM)
		// Day configuration - Thursday is false for testing time-based blocking
		Monday:    true,  // Active on Monday
		Tuesday:   true,  // Active on Tuesday
		Wednesday: true,  // Active on Wednesday
		Thursday:  false, // INACTIVE on Thursday (for testing)
		Friday:    true,  // Active on Friday
		Saturday:  true,  // Active on Saturday
		Sunday:    true,  // Active on Sunday
		// Link both devices to this schedule
		DeviceHashes: []string{d1h, d2h},
		// Block some domains directly (not via lists)
		Domains: []string{"example.com", "test.com"},
		// Apply all three blocking lists
		ListIds:  []int{l1i, l2i, l3i},
		UserHash: uh, // Links schedule to the test user
	})
	if err != nil {
		return "", err
	}

	// Return the user hash so tests can reference the created data
	return uh, nil
}
