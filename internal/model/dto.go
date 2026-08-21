package model

import "time"

// StatCount represents a query count at a specific timestamp
type StatCount struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

// QueryStatsDTO represents combined query statistics. QPS and BlockedQPS hold
// the peak queries-per-second sample of each 5-minute window over the last
// 24 hours, at the timestamp of the second where the peak occurred.
type QueryStatsDTO struct {
	Total      []StatCount         `json:"total"`
	Blocked    []StatCount         `json:"blocked"`
	QPS        []StatCount         `json:"qps,omitempty"`
	BlockedQPS []StatCount         `json:"blockedQps,omitempty"`
	Domains    []DomainStat        `json:"domains"`
	Activity   []DeviceActivityDTO `json:"activity"`
}

// DeviceActivityDTO contains one device's hourly query totals and last report time.
type DeviceActivityDTO struct {
	DeviceID    int        `json:"deviceId"`
	DeviceName  string     `json:"deviceName"`
	Hours       []int      `json:"hours"`
	LastQueryAt *time.Time `json:"lastQueryAt,omitempty"`
}

// UserDTO represents a user data transfer object
type UserDTO struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Devices      int    `json:"devices"`
	Lists        int    `json:"lists"`
	Schedules    int    `json:"schedules"`
}

// DeviceDTO represents a device data transfer object
type DeviceDTO struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	OS            OS         `json:"os"`
	Hash          string     `json:"hash"`
	UserID        int        `json:"userId"`
	ScheduleIDs   []int      `json:"scheduleIds"`
	ScheduleNames []string   `json:"scheduleNames"`
	LastQueryAt   *time.Time `json:"lastQueryAt,omitempty"`
}

// ListDTO represents a list data transfer object
type ListDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	UserID      int    `json:"userId"`
	Rules       int    `json:"rules"`
	Schedules   int    `json:"schedules"`
	IsDefault   bool   `json:"isDefault"`
}

// ScheduleDTO represents a schedule data transfer object
type ScheduleDTO struct {
	ID          int      `json:"id"`
	StartTime   string   `json:"startTime"`
	EndTime     string   `json:"endTime"`
	Name        string   `json:"name"`
	Active      bool     `json:"active"`
	UserID      int      `json:"userId"`
	DeviceIDs   []int    `json:"deviceIds"`
	RuleIDs     []int    `json:"ruleIds"`
	ListIDs     []int    `json:"listIds"`
	DeviceNames []string `json:"deviceNames"`
	Rules       int      `json:"rules"`
	ListNames   []string `json:"listNames"`
	Monday      bool     `json:"monday"`
	Tuesday     bool     `json:"tuesday"`
	Wednesday   bool     `json:"wednesday"`
	Thursday    bool     `json:"thursday"`
	Friday      bool     `json:"friday"`
	Saturday    bool     `json:"saturday"`
	Sunday      bool     `json:"sunday"`
	IsDefault   bool     `json:"isDefault"`
}
