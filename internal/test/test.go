// Package test provides testing utilities and fixtures.
package test

import (
	"time"

	"github.com/lemon3studio/blkhole/internal/cache"
	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/lemon3studio/blkhole/internal/repos"
	"github.com/lemon3studio/blkhole/internal/services"
)

// Test interface defines methods for creating test data fixtures.
type Test interface {
	AddUser(user model.User, password string) (int, error)
	AddDevice(device model.Device) (int, error)
	AddRule(rule model.Rule) (int, error)
	AddList(list model.List, domains []string) (int, error)
	AddSchedule(schedule model.Schedule) (int, error)
	Test() (int, error)
}

// test implements the Test interface.
type test struct {
	users         repos.UserRepo
	devices       repos.DeviceRepo
	rules         repos.RuleRepo
	lists         repos.ListRepo
	listService   services.ListsService
	schedules     repos.ScheduleRepo
	cryptoService services.CryptoService
	domains       repos.DomainRepo
	statsCache    cache.StatsCache
	deviceCache   cache.DeviceCache
}

// NewTest creates a new test instance.
func NewTest(users repos.UserRepo, devices repos.DeviceRepo, rules repos.RuleRepo, lists repos.ListRepo, listService services.ListsService, schedules repos.ScheduleRepo, cryptoService services.CryptoService, domains repos.DomainRepo, statsCache cache.StatsCache, deviceCache cache.DeviceCache) Test {
	return &test{
		users:         users,
		devices:       devices,
		rules:         rules,
		lists:         lists,
		listService:   listService,
		schedules:     schedules,
		cryptoService: cryptoService,
		domains:       domains,
		statsCache:    statsCache,
		deviceCache:   deviceCache,
	}
}

// AddUser creates a new user with hashed password.
func (t *test) AddUser(user model.User, password string) (int, error) {
	passwordHash, err := t.cryptoService.HashPassword(password)
	if err != nil {
		return 0, err
	}

	user.PasswordHash = passwordHash

	err = t.users.Create(&user)
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

// AddDevice creates a new device.
func (t *test) AddDevice(device model.Device) (int, error) {
	hash, err := t.cryptoService.RandomHash()
	if err != nil {
		return 0, err
	}

	device.Hash = hash

	err = t.devices.Create(&device)
	if err != nil {
		return 0, err
	}

	// Reload device cache to include the new device
	allDevices, err := t.devices.FindAll()
	if err != nil {
		return device.ID, err
	}
	t.deviceCache.LoadDevices(allDevices)

	return device.ID, nil
}

// AddRule creates a new rule.
func (t *test) AddRule(rule model.Rule) (int, error) {
	id, err := t.rules.CreateOrGet(&rule)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// AddList creates a new list with blocking rules for the given domains.
func (t *test) AddList(list model.List, domains []string) (int, error) {
	id, err := t.lists.Create(&list)
	if err != nil {
		return 0, err
	}

	for _, domain := range domains {
		// Create or get domain first
		domainModel := &model.Domain{Name: domain}
		domainID, err := t.domains.CreateOrGet(domainModel)
		if err != nil {
			return id, err
		}

		ruleID, err := t.AddRule(model.Rule{
			DomainID: domainID,
			Allowed:  false,
		})
		if err != nil {
			return id, err
		}

		if err := t.rules.LinkToList(ruleID, id); err != nil {
			return id, err
		}
	}

	list.ID = id
	err = t.listService.LoadList(&list)
	if err != nil {
		return id, err
	}

	return id, nil
}

// AddSchedule creates a new schedule with linked devices, rules, and lists.
func (t *test) AddSchedule(schedule model.Schedule) (int, error) {
	// Create now automatically links devices and lists
	id, err := t.schedules.Create(&schedule)
	if err != nil {
		return -1, err
	}

	// Only need to manually link rules since Create doesn't handle those
	for _, ruleID := range schedule.RuleIDs {
		err = t.schedules.LinkRule(id, ruleID)
		if err != nil {
			return id, err
		}
	}

	return id, nil
}

// Test creates a comprehensive test scenario with a user, devices, lists, and schedule.
// Thursday is left unblocked for testing time-based rules.
func (t *test) Test() (int, error) {
	ui, err := t.AddUser(model.User{
		Name:  "Arian Gohari",
		Email: "arian@gohari.de",
	}, "SomePassword")
	if err != nil {
		return 0, err
	}

	device1 := model.Device{
		Name:   "iPhone von Arian",
		OS:     model.IOS,
		UserID: ui,
	}
	d1i, err := t.AddDevice(device1)
	if err != nil {
		return 0, err
	}

	device2 := model.Device{
		Name:   "MacBook Pro von Arian",
		OS:     model.MacOS,
		UserID: ui,
	}
	d2i, err := t.AddDevice(device2)
	if err != nil {
		return 0, err
	}

	l1i, err := t.AddList(model.List{
		Name:        "Hagezi Multi ULTIMATE",
		Description: "Ultimate Sweeper - Strictly cleans the Internet and protects your privacy! Blocks Ads, Affiliate, Tracking, Metrics, Telemetry, Phishing, Malware, Scam, Free Hoster, Fake, Cryptojacking and other Crap.",
		Source:      "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/adblock/pro.plus.mini.txt",
		UserID:      ui,
	}, []string{})
	if err != nil {
		return 0, err
	}

	l2i, err := t.AddList(model.List{
		Name:        "OISD Big",
		Description: "Blocks Ads, (Mobile) App Ads, Phishing, Malvertising, Malware, Spyware, Ransomware, CryptoJacking, Telemetry/Analytics/Tracking (where not needed for proper functionality).",
		Source:      "https://big.oisd.nl",
		UserID:      ui,
	}, []string{})
	if err != nil {
		return 0, err
	}

	l3i, err := t.AddList(model.List{
		Name:        "Anti Axel Springer",
		Description: "This list blocks all connections to sites which are from Axel Springer Verlag or have a connection with them.",
		Source:      "https://raw.githubusercontent.com/autinerd/anti-axelspringer-hosts/master/axelspringer-hosts",
		UserID:      ui,
	}, []string{})
	if err != nil {
		return 0, err
	}

	// Create domains first
	exampleDomain := &model.Domain{Name: "example.com"}
	exampleDomainID, err := t.domains.CreateOrGet(exampleDomain)
	if err != nil {
		return 0, err
	}

	testDomain := &model.Domain{Name: "test.com"}
	testDomainID, err := t.domains.CreateOrGet(testDomain)
	if err != nil {
		return 0, err
	}

	r1i, err := t.AddRule(model.Rule{
		DomainID: exampleDomainID,
		Allowed:  false,
	})
	if err != nil {
		return 0, err
	}

	r2i, err := t.AddRule(model.Rule{
		DomainID: testDomainID,
		Allowed:  false,
	})
	if err != nil {
		return 0, err
	}

	_, err = t.AddSchedule(model.Schedule{
		Name:      "Malicious Content",
		StartTime: "00:00",
		EndTime:   "23:55",
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  false,
		Sunday:    true,
		DeviceIDs: []int{d1i, d2i},
		RuleIDs:   []int{r1i, r2i},
		ListIDs:   []int{l1i, l2i, l3i},
		UserID:    ui,
	})
	if err != nil {
		return -1, err
	}

	// Add realistic query stats for the last 24 hours
	d1, err := t.devices.FindByID(d1i)
	if err != nil {
		return -1, err
	}

	d2, err := t.devices.FindByID(d2i)
	if err != nil {
		return -1, err
	}

	// Generate realistic query patterns for last 24 hours
	now := time.Now()
	for hoursAgo := 23; hoursAgo >= 0; hoursAgo-- {
		hour := now.Add(-time.Duration(hoursAgo) * time.Hour)

		// Simulate realistic usage patterns (more queries during daytime)
		var queriesPerHour, blockedPerHour int
		hourOfDay := hour.Hour()

		switch {
		case hourOfDay >= 0 && hourOfDay < 6: // Night (low activity)
			queriesPerHour = 5 + hoursAgo%3
			blockedPerHour = 1 + hoursAgo%2
		case hourOfDay >= 6 && hourOfDay < 9: // Morning (medium activity)
			queriesPerHour = 30 + hoursAgo%15
			blockedPerHour = 7 + hoursAgo%4
		case hourOfDay >= 9 && hourOfDay < 17: // Work hours (high activity)
			queriesPerHour = 80 + hoursAgo%30
			blockedPerHour = 18 + hoursAgo%8
		case hourOfDay >= 17 && hourOfDay < 22: // Evening (medium activity)
			queriesPerHour = 45 + hoursAgo%20
			blockedPerHour = 10 + hoursAgo%5
		default: // Late evening (low-medium activity)
			queriesPerHour = 20 + hoursAgo%10
			blockedPerHour = 4 + hoursAgo%3
		}

		// Add stats for iPhone (more mobile usage)
		t.statsCache.IncrementAt(d1.Hash, hour, queriesPerHour)
		t.statsCache.IncrementBlockedAt(d1.Hash, hour, blockedPerHour)

		// Add stats for MacBook (more work-related usage during work hours)
		macQueriesPerHour := queriesPerHour
		macBlockedPerHour := blockedPerHour
		if hourOfDay >= 9 && hourOfDay < 17 {
			macQueriesPerHour = int(float64(queriesPerHour) * 1.5) // 50% more during work
			macBlockedPerHour = int(float64(blockedPerHour) * 1.3)
		}
		t.statsCache.IncrementAt(d2.Hash, hour, macQueriesPerHour)
		t.statsCache.IncrementBlockedAt(d2.Hash, hour, macBlockedPerHour)
	}

	return ui, nil
}
