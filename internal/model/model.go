// Package model defines the data structures used throughout the Leo DNS blocker application.
package model

// User represents a user in the system
type User struct {
	Hash         string   `json:"hash"`
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	PasswordHash string   `json:"-"`
	DeviceHashes []string `json:"deviceHashes"`
	ListIds      []int    `json:"listIds"`
	ScheduleIds  []int    `json:"scheduleIds"`
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
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	OS          OS     `json:"os"`
	UserHash    string `json:"userHash"`
	ScheduleIds []int  `json:"scheduleIds"`
}

// Rule represents a domain rule with allow/block behavior
type Rule struct {
	Domain  string `json:"domain"`
	ListID  int    `json:"listId"`
	Allowed bool   `json:"allowed"`
	ID      int    `json:"id"`
}

// List represents a list of domain rules
type List struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	UserHash    string `json:"userHash"`
	Rules       []Rule `json:"rules"`
	ScheduleIds []int  `json:"scheduleIds"`
	ID          int    `json:"id"`
}

// Schedule represents a time-based schedule
type Schedule struct {
	StartTime    string   `json:"startTime"`
	EndTime      string   `json:"endTime"`
	Name         string   `json:"name"`
	UserHash     string   `json:"userHash"`
	DeviceHashes []string `json:"deviceHashes"`
	Domains      []string `json:"domains"`
	ListIds      []int    `json:"listIds"`
	Monday       bool     `json:"monday"`
	Tuesday      bool     `json:"tuesday"`
	Wednesday    bool     `json:"wednesday"`
	Thursday     bool     `json:"thursday"`
	Friday       bool     `json:"friday"`
	Saturday     bool     `json:"saturday"`
	Sunday       bool     `json:"sunday"`
	ID           int      `json:"id"`
}

// Quote represents a quote with author
type Quote struct {
	Quote  string `json:"quote"`
	Author string `json:"author"`
}
