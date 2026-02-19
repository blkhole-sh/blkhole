// Package model defines the data structures used throughout the blkhole DNS blocker application.
package model

// User represents a user in the system
type User struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	DeviceIDs    []int  `json:"deviceIds"`
	ListIDs      []int  `json:"listIds"`
	ScheduleIDs  []int  `json:"scheduleIds"`
}

type OS string

// OS represents the operating system type
const (
	MacOS   OS = "macOS"
	IOS     OS = "iOS"
	Android OS = "Android"
	Linux   OS = "Linux"
	Windows OS = "Windows"
)

// Device represents a device in the system
type Device struct {
	ID          int    `json:"id"`
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	OS          OS     `json:"os"`
	UserID      int    `json:"userId"`
	ScheduleIDs []int  `json:"scheduleIds"`
}

// Domain represents a domain with its id in the system
type Domain struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Rule represents a domain rule with allow/block behavior
type Rule struct {
	ID       int  `json:"id"`
	DomainID int  `json:"domainId"`
	Allowed  bool `json:"allowed"`
}

// List represents a list of domain rules
type List struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	UserID      int    `json:"userId"`
	Rules       []Rule `json:"rules"`
	ScheduleIDs []int  `json:"scheduleIds"`
	ID          int    `json:"id"`
}

// Schedule represents a time-based schedule
type Schedule struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Name      string `json:"name"`
	UserID    int    `json:"userId"`
	DeviceIDs []int  `json:"deviceIds"`
	RuleIDs   []int  `json:"ruleIds"`
	ListIDs   []int  `json:"listIds"`
	Monday    bool   `json:"monday"`
	Tuesday   bool   `json:"tuesday"`
	Wednesday bool   `json:"wednesday"`
	Thursday  bool   `json:"thursday"`
	Friday    bool   `json:"friday"`
	Saturday  bool   `json:"saturday"`
	Sunday    bool   `json:"sunday"`
	ID        int    `json:"id"`
}

// DeviceSchedule represents the many-to-many relationship between devices and schedules
type DeviceSchedule struct {
	DeviceID   int `json:"deviceId"`
	ScheduleID int `json:"scheduleId"`
}

// ScheduleRule represents the many-to-many relationship between schedules and rules
type ScheduleRule struct {
	ScheduleID int `json:"scheduleId"`
	RuleID     int `json:"ruleId"`
}

// Quote represents a quote with author
type Quote struct {
	Quote  string `json:"quote"`
	Author string `json:"author"`
}

// ToDTO converts a User model to UserDTO with counts
func (u User) ToDTO() UserDTO {
	return UserDTO{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Devices:   len(u.DeviceIDs),
		Lists:     len(u.ListIDs),
		Schedules: len(u.ScheduleIDs),
	}
}

// ToDTO converts a Device model to DeviceDTO with schedule names
func (d Device) ToDTO(scheduleNames []string) DeviceDTO {
	if scheduleNames == nil {
		scheduleNames = []string{}
	}
	return DeviceDTO{
		ID:            d.ID,
		Hash:          d.Hash,
		Name:          d.Name,
		OS:            d.OS,
		UserID:        d.UserID,
		ScheduleNames: scheduleNames,
	}
}

// ToDTO converts a List model to ListDTO with counts
func (l List) ToDTO() ListDTO {
	return ListDTO{
		ID:          l.ID,
		Name:        l.Name,
		Description: l.Description,
		Source:      l.Source,
		UserID:      l.UserID,
		Rules:       len(l.Rules),
		Schedules:   len(l.ScheduleIDs),
	}
}

// ToDTO converts Schedule model to ScheduleDTO with device and list names
func (s Schedule) ToDTO(deviceNames []string, listNames []string) ScheduleDTO {
	if deviceNames == nil {
		deviceNames = []string{}
	}
	if listNames == nil {
		listNames = []string{}
	}
	return ScheduleDTO{
		StartTime:   s.StartTime,
		EndTime:     s.EndTime,
		Name:        s.Name,
		UserID:      s.UserID,
		DeviceIDs:   s.DeviceIDs,
		RuleIDs:     s.RuleIDs,
		ListIDs:     s.ListIDs,
		DeviceNames: deviceNames,
		Rules:       len(s.RuleIDs),
		ListNames:   listNames,
		Monday:      s.Monday,
		Tuesday:     s.Tuesday,
		Wednesday:   s.Wednesday,
		Thursday:    s.Thursday,
		Friday:      s.Friday,
		Saturday:    s.Saturday,
		Sunday:      s.Sunday,
		ID:          s.ID,
	}
}
