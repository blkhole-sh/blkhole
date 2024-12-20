package model

// Define User struct
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

// Define OS enum
const (
	MacOS   OS = "macOS"
	IOS     OS = "iOS"
	Android    = "Android"
	Linux      = "Linux"
	Windows    = "Windows"
)

// Define Device struct
type Device struct {
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	OS          OS     `json:"os"`
	UserHash    string `json:"userHash"`
	ScheduleIds []int  `json:"scheduleIds"`
}

// Define Domain struct
type Domain struct {
	Name        string `json:"name"`
	ListIds     []int  `json:"listIds"`
	ScheduleIds []int  `json:"scheduleIds"`
	Id          int    `json:"id"`
}

// Define List struct
type List struct {
	Name        string `json:"name"`
	UserHash    string `json:"userHash"`
	DomainIds   []int  `json:"domainIds"`
	ScheduleIds []int  `json:"scheduleIds"`
	Id          int    `json:"id"`
}

// Define Schedule struct
type Schedule struct {
	StartTime    string   `json:"startTime"`
	EndTime      string   `json:"endTime"`
	Name         string   `json:"name"`
	UserHash     string   `json:"userHash"`
	DeviceHashes []string `json:"deviceHashes"`
	DomainIds    []int    `json:"domainIds"`
	ListIds      []int    `json:"listIds"`
	Monday       bool     `json:"monday"`
	Tuesday      bool     `json:"tuesday"`
	Wednesday    bool     `json:"wednesday"`
	Thursday     bool     `json:"thursday"`
	Friday       bool     `json:"friday"`
	Saturday     bool     `json:"saturday"`
	Sunday       bool     `json:"sunday"`
	Id           int      `json:"id"`
}

// Define Quote struct
type Quote struct {
	Quote  string `json:"quote"`
	Author string `json:"author"`
}
