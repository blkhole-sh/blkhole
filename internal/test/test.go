// Package test provides testing utilities and fixtures.
package test

import (
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

// Test creates a user and loads blocklists. Devices and schedules are configured manually via the dashboard.
func (t *test) Test() (int, error) {
	ui, err := t.AddUser(model.User{
		Name:  "Arian Gohari",
		Email: "arian@gohari.de",
	}, "SomePassword")
	if err != nil {
		return 0, err
	}

	_, err = t.AddList(model.List{
		Name:        "Hagezi Multi ULTIMATE",
		Description: "Ultimate Sweeper - Strictly cleans the Internet and protects your privacy! Blocks Ads, Affiliate, Tracking, Metrics, Telemetry, Phishing, Malware, Scam, Free Hoster, Fake, Cryptojacking and other Crap.",
		Source:      "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/adblock/pro.plus.mini.txt",
		UserID:      ui,
	}, []string{})
	if err != nil {
		return 0, err
	}

	_, err = t.AddList(model.List{
		Name:        "OISD Big",
		Description: "Blocks Ads, (Mobile) App Ads, Phishing, Malvertising, Malware, Spyware, Ransomware, CryptoJacking, Telemetry/Analytics/Tracking (where not needed for proper functionality).",
		Source:      "https://big.oisd.nl",
		UserID:      ui,
	}, []string{})
	if err != nil {
		return 0, err
	}

	_, err = t.AddList(model.List{
		Name:        "Anti Axel Springer",
		Description: "This list blocks all connections to sites which are from Axel Springer Verlag or have a connection with them.",
		Source:      "https://raw.githubusercontent.com/autinerd/anti-axelspringer-hosts/master/axelspringer-hosts",
		UserID:      ui,
	}, []string{})
	if err != nil {
		return 0, err
	}

	return ui, nil
}
