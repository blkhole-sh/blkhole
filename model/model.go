package model

import "time"

// Define user struct
type User struct {
	Hash         string
	Name         string
	Email        string
	PasswordHash string
	Devices      []Device
}

// Define device struct
type Device struct {
	Hash     string
	Name     string
	UserHash string
}

// Define domain struct
type Domain struct {
	Name string
	Id   int
}

// Define category struct
type Category struct {
	Name string
	Id   int
}

// Define domain rule struct
type DomainRule struct {
	Domain     *Domain
	Blocked    bool
	Id         int
	ScheduleId int
}

// Define category rule struct
type CategoryRule struct {
	Category   *Category
	Blocked    bool
	Id         int
	ScheduleId int
}

// Define schedule struct
type Schedule struct {
	StartTime             time.Time
	EndTime               time.Time
	Description           string
	DeviceHash            string
	DomainBlockingRules   []DomainRule
	CategoryBlockingRules []CategoryRule
	Monday                bool
	Tuesday               bool
	Wednesday             bool
	Thursday              bool
	Friday                bool
	Saturday              bool
	Sunday                bool
	Id                    int
}
