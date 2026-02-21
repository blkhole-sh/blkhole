package model

import "time"

// StatCount represents a query count at a specific timestamp
type StatCount struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

// QueryStatsDTO represents combined query statistics
type QueryStatsDTO struct {
	Total   []StatCount `json:"total"`
	Blocked []StatCount `json:"blocked"`
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
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	OS            OS       `json:"os"`
	Hash          string   `json:"hash"`
	UserID        int      `json:"userId"`
	ScheduleIDs   []int    `json:"scheduleIds"`
	ScheduleNames []string `json:"scheduleNames"`
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
}
