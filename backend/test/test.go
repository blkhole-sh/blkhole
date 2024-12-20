package test

import (
	"server/model"
	"server/repos"
	"server/services"
)

type Test interface {
	AddUser(name string, email string, password string) (string, error)
	AddDevice(name string, os model.OS, userHash string) (string, error)
	AddDomains(domains []string) ([]int, error)
	AddList(name string, domainIds []int, userHash string) (int, error)
	AddSchedule(name string, startTime string, endTime string, monday bool, tuesday bool, wednesday bool, thursday bool, friday bool, saturday bool, sunday bool, deviceHashes []string, domainIds []int, listIds []int, userHash string) (int, error)
	Test() error
}

type TestImpl struct {
	userRepo      repos.UserRepo
	deviceRepo    repos.DeviceRepo
	domainRepo    repos.DomainRepo
	listRepo      repos.ListRepo
	scheduleRepo  repos.ScheduleRepo
	cryptoService services.CryptoService
}

func NewTest(userRepo repos.UserRepo, deviceRepo repos.DeviceRepo, domainRepo repos.DomainRepo, listRepo repos.ListRepo, scheduleRepo repos.ScheduleRepo, cryptoService services.CryptoService) Test {
	return &TestImpl{
		userRepo:      userRepo,
		deviceRepo:    deviceRepo,
		domainRepo:    domainRepo,
		listRepo:      listRepo,
		scheduleRepo:  scheduleRepo,
		cryptoService: cryptoService,
	}
}

func (t TestImpl) AddUser(name string, email string, password string) (string, error) {
	hash, err := t.cryptoService.RandomHash()
	if err != nil {
		return "", err
	}

	passwordHash, err := t.cryptoService.HashPassword(password)
	if err != nil {
		return "", err
	}

	u := model.User{
		Hash:         hash,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
	}

	err = t.userRepo.Create(&u)
	if err != nil {
		return "", err
	}

	return u.Hash, nil
}

func (t TestImpl) AddDevice(name string, os model.OS, userHash string) (string, error) {
	hash, err := t.cryptoService.RandomHash()
	if err != nil {
		return "", err
	}

	d := model.Device{
		Hash:     hash,
		Name:     name,
		OS:       os,
		UserHash: userHash,
	}

	err = t.deviceRepo.Create(&d)
	if err != nil {
		return "", err
	}

	return d.Hash, nil
}

func (t TestImpl) AddDomains(domains []string) ([]int, error) {
	ids := []int{}
	for _, name := range domains {
		d := model.Domain{Name: name}
		id, err := t.domainRepo.Create(&d)
		if err != nil {
			return ids, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (t TestImpl) AddList(name string, domainIds []int, userHash string) (int, error) {
	l := model.List{
		Name:     name,
		UserHash: userHash,
	}

	id, err := t.listRepo.Create(&l)
	if err != nil {
		return -1, err
	}

	for _, di := range domainIds {
		err := t.listRepo.LinkDomain(id, di)
		if err != nil {
			return id, err
		}
	}

	return id, nil
}

func (t TestImpl) AddSchedule(name string, startTime string, endTime string, monday bool, tuesday bool, wednesday bool, thursday bool, friday bool, saturday bool, sunday bool, deviceHashes []string, domainIds []int, listIds []int, userHash string) (int, error) {
	s := model.Schedule{
		Name:      name,
		StartTime: startTime,
		EndTime:   endTime,
		Monday:    monday,
		Tuesday:   tuesday,
		Wednesday: wednesday,
		Thursday:  thursday,
		Friday:    friday,
		Saturday:  saturday,
		Sunday:    sunday,
		UserHash:  userHash,
	}

	id, err := t.scheduleRepo.Create(&s)
	if err != nil {
		return -1, err
	}

	for _, dh := range deviceHashes {
		err = t.scheduleRepo.LinkDevice(id, dh)
		if err != nil {
			return id, err
		}
	}

	for _, di := range domainIds {
		err = t.scheduleRepo.LinkDomain(id, di)
		if err != nil {
			return id, err
		}

	}

	for _, li := range listIds {
		err = t.scheduleRepo.LinkList(id, li)
		if err != nil {
			return id, err
		}
	}

	return id, nil
}

func (t TestImpl) Test() error {
	uh, err := t.AddUser("Arian Gohari", "arian@gohari.de", "SomePassword")
	if err != nil {
		return err
	}

	d1h, err := t.AddDevice("IPhone von Arian", model.IOS, uh)
	if err != nil {
		return err
	}

	d2h, err := t.AddDevice("MacBook Pro von Arian", model.MacOS, uh)
	if err != nil {
		return err
	}

	dis, err := t.AddDomains([]string{"reddit.com", "youtube.com", "linkedin.com", "instagram.com", "startmunich.de"})
	if err != nil {
		return err
	}

	l1i, err := t.AddList("Social Media", []int{dis[0], dis[2], dis[3]}, uh)
	if err != nil {
		return err
	}

	l2i, err := t.AddList("Addictive", []int{dis[1]}, uh)
	if err != nil {
		return err
	}

	l3i, err := t.AddList("Cringe", []int{dis[2], dis[4]}, uh)
	if err != nil {
		return err
	}

	_, err = t.AddSchedule("My Blocklist", "09:00", "17:00", true, true, true, false, true, true, true, []string{d1h, d2h}, dis, []int{l1i, l2i, l3i}, uh)
	if err != nil {
		return err
	}

	return nil
}
