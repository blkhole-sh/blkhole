// Package model defines the data structures used throughout the Leo DNS blocker application.
package model

// User represents a user in the system
type User struct {
	Hash         string   `json:"hash" msgpack:"hash"`
	Name         string   `json:"name" msgpack:"name"`
	Email        string   `json:"email" msgpack:"email"`
	PasswordHash string   `json:"-" msgpack:"-"`
	DeviceHashes []string `json:"deviceHashes" msgpack:"deviceHashes"`
	ListIds      []int    `json:"listIds" msgpack:"listIds"`
	ScheduleIds  []int    `json:"scheduleIds" msgpack:"scheduleIds"`
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
	Hash        string `json:"hash" msgpack:"hash"`
	Name        string `json:"name" msgpack:"name"`
	OS          OS     `json:"os" msgpack:"os"`
	UserHash    string `json:"userHash" msgpack:"userHash"`
	ScheduleIds []int  `json:"scheduleIds" msgpack:"scheduleIds"`
}

// Rule represents a domain rule with allow/block behavior
type Rule struct {
	Domain  string `json:"domain" msgpack:"domain"`
	ListID  int    `json:"listId" msgpack:"listId"`
	Allowed bool   `json:"allowed" msgpack:"allowed"`
	ID      int    `json:"id" msgpack:"id"`
}

// List represents a list of domain rules
type List struct {
	Name        string `json:"name" msgpack:"name"`
	Description string `json:"description" msgpack:"description"`
	Source      string `json:"source" msgpack:"source"`
	UserHash    string `json:"userHash" msgpack:"userHash"`
	Rules       []Rule `json:"rules" msgpack:"rules"`
	ScheduleIds []int  `json:"scheduleIds" msgpack:"scheduleIds"`
	ID          int    `json:"id" msgpack:"id"`
}

// Schedule represents a time-based schedule
type Schedule struct {
	StartTime    string   `json:"startTime" msgpack:"startTime"`
	EndTime      string   `json:"endTime" msgpack:"endTime"`
	Name         string   `json:"name" msgpack:"name"`
	UserHash     string   `json:"userHash" msgpack:"userHash"`
	DeviceHashes []string `json:"deviceHashes" msgpack:"deviceHashes"`
	Domains      []string `json:"domains" msgpack:"domains"`
	ListIds      []int    `json:"listIds" msgpack:"listIds"`
	Monday       bool     `json:"monday" msgpack:"monday"`
	Tuesday      bool     `json:"tuesday" msgpack:"tuesday"`
	Wednesday    bool     `json:"wednesday" msgpack:"wednesday"`
	Thursday     bool     `json:"thursday" msgpack:"thursday"`
	Friday       bool     `json:"friday" msgpack:"friday"`
	Saturday     bool     `json:"saturday" msgpack:"saturday"`
	Sunday       bool     `json:"sunday" msgpack:"sunday"`
	ID           int      `json:"id" msgpack:"id"`
}

// Quote represents a quote with author
type Quote struct {
	Quote  string `json:"quote" msgpack:"quote"`
	Author string `json:"author" msgpack:"author"`
}
