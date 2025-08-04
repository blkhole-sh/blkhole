// Package test provides testing utilities for the Leo DNS blocker application.
package test

import (
	"server/internal/model"
	"server/internal/repos"
	"server/internal/services"
)

type Test interface {
	AddUser(user model.User, password string) (string, error)
	AddDevice(device model.Device) (string, error)
	AddRule(rule model.Rule) (int, error)
	AddList(list model.List, domains []string) (int, error)
	AddSchedule(schedule model.Schedule) (int, error)
	Test() (string, error)
}

type TestImpl struct {
	userRepo      repos.UserRepo
	deviceRepo    repos.DeviceRepo
	ruleRepo      repos.RuleRepo
	listRepo      repos.ListRepo
	listService   services.ListsService
	scheduleRepo  repos.ScheduleRepo
	cryptoService services.CryptoService
}

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

func (t TestImpl) AddUser(user model.User, password string) (string, error) {
	hash, err := t.cryptoService.RandomHash()
	if err != nil {
		return "", err
	}

	passwordHash, err := t.cryptoService.HashPassword(password)
	if err != nil {
		return "", err
	}

	user.Hash = hash
	user.PasswordHash = passwordHash

	err = t.userRepo.Create(&user)
	if err != nil {
		return "", err
	}

	return user.Hash, nil
}

func (t TestImpl) AddDevice(device model.Device) (string, error) {
	hash, err := t.cryptoService.RandomHash()
	if err != nil {
		return "", err
	}

	device.Hash = hash

	err = t.deviceRepo.Create(&device)
	if err != nil {
		return "", err
	}

	return device.Hash, nil
}

func (t TestImpl) AddRule(rule model.Rule) (int, error) {
	id, err := t.ruleRepo.Create(&rule)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (t TestImpl) AddList(list model.List, domains []string) (int, error) {
	id, err := t.listRepo.Create(&list)
	if err != nil {
		return -1, err
	}

	// Create rules for each domain (default to blocked)
	for _, domain := range domains {
		_, err := t.AddRule(model.Rule{
			Domain:  domain,
			ListID:  id,
			Allowed: false,
		})
		if err != nil {
			return id, err
		}
	}

	t.listService.LoadList(&list)

	return id, nil
}

func (t TestImpl) AddSchedule(schedule model.Schedule) (int, error) {
	id, err := t.scheduleRepo.Create(&schedule)
	if err != nil {
		return -1, err
	}

	for _, dh := range schedule.DeviceHashes {
		err = t.scheduleRepo.LinkDevice(id, dh)
		if err != nil {
			return id, err
		}
	}

	for _, domain := range schedule.Domains {
		err = t.scheduleRepo.LinkDomain(id, domain)
		if err != nil {
			return id, err
		}
	}

	for _, li := range schedule.ListIds {
		err = t.scheduleRepo.LinkList(id, li)
		if err != nil {
			return id, err
		}
	}

	return id, nil
}

func (t TestImpl) Test() (string, error) {
	uh, err := t.AddUser(model.User{
		Name:  "Arian Gohari",
		Email: "arian@gohari.de",
	}, "SomePassword")
	if err != nil {
		return "", err
	}

	d1h, err := t.AddDevice(model.Device{
		Name:     "IPhone von Arian",
		OS:       model.IOS,
		UserHash: uh,
	})
	if err != nil {
		return "", err
	}

	d2h, err := t.AddDevice(model.Device{
		Name:     "MacBook Pro von Arian",
		OS:       model.MacOS,
		UserHash: uh,
	})
	if err != nil {
		return "", err
	}

	l1i, err := t.AddList(model.List{
		Name:        "Hagezi Multi ULTIMATE",
		Description: "Ultimate Sweeper - Strictly cleans the Internet and protects your privacy! Blocks Ads, Affiliate, Tracking, Metrics, Telemetry, Phishing, Malware, Scam, Free Hoster, Fake, Cryptojacking and other Crap.",
		Source:      "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/adblock/pro.plus.mini.txt",
		UserHash:    uh,
	}, []string{})
	if err != nil {
		return "", err
	}

	l2i, err := t.AddList(model.List{
		Name:        "OISD Big",
		Description: "Blocks Ads, (Mobile) App Ads, Phishing, Malvertising, Malware, Spyware, Ransomware, CryptoJacking, Telemetry/Analytics/Tracking (where not needed for proper functionality).",
		Source:      "https://big.oisd.nl",
		UserHash:    uh,
	}, []string{})
	if err != nil {
		return "", err
	}

	l3i, err := t.AddList(model.List{
		Name:        "Anti Axel Springer",
		Description: "This list blocks all connections to sites which are from Axel Springer Verlag or have a connection with them.",
		Source:      "https://raw.githubusercontent.com/autinerd/anti-axelspringer-hosts/master/axelspringer-hosts",
		UserHash:    uh,
	}, []string{})
	if err != nil {
		return "", err
	}

	_, err = t.AddSchedule(model.Schedule{
		Name:         "Base Protection",
		StartTime:    "09:00",
		EndTime:      "17:00",
		Monday:       true,
		Tuesday:      true,
		Wednesday:    true,
		Thursday:     false,
		Friday:       true,
		Saturday:     true,
		Sunday:       true,
		DeviceHashes: []string{d1h, d2h},
		Domains:      []string{"example.com", "test.com"},
		ListIds:      []int{l1i, l2i, l3i},
		UserHash:     uh,
	})
	if err != nil {
		return "", err
	}

	return uh, nil
}
