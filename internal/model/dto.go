package model

// UserDTO represents a user data transfer object
type UserDTO struct {
	Hash         string `json:"hash" msgpack:"hash"`
	Name         string `json:"name" msgpack:"name"`
	Email        string `json:"email" msgpack:"email"`
	PasswordHash string `json:"-" msgpack:"-"`
}

// DeviceDTO represents a device data transfer object
type DeviceDTO struct {
	Hash     string `json:"hash" msgpack:"hash"`
	Name     string `json:"name" msgpack:"name"`
	OS       OS     `json:"os" msgpack:"os"`
	UserHash string `json:"userHash" msgpack:"userHash"`
}

// ListDTO represents a list data transfer object
type ListDTO struct {
	Name        string `json:"name" msgpack:"name"`
	Description string `json:"description" msgpack:"description"`
	Source      string `json:"source" msgpack:"source"`
	UserHash    string `json:"userHash" msgpack:"userHash"`
	ID          int    `json:"id" msgpack:"id"`
}

// ScheduleDTO represents a schedule data transfer object
type ScheduleDTO struct {
	StartTime string `json:"startTime" msgpack:"startTime"`
	EndTime   string `json:"endTime" msgpack:"endTime"`
	Name      string `json:"name" msgpack:"name"`
	UserHash  string `json:"userHash" msgpack:"userHash"`
	Monday    bool   `json:"monday" msgpack:"monday"`
	Tuesday   bool   `json:"tuesday" msgpack:"tuesday"`
	Wednesday bool   `json:"wednesday" msgpack:"wednesday"`
	Thursday  bool   `json:"thursday" msgpack:"thursday"`
	Friday    bool   `json:"friday" msgpack:"friday"`
	Saturday  bool   `json:"saturday" msgpack:"saturday"`
	Sunday    bool   `json:"sunday" msgpack:"sunday"`
	ID        int    `json:"id" msgpack:"id"`
}
