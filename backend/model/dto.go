package model

// UserDTO represents a user data transfer object
type UserDTO struct {
	Hash         string `json:"hash"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

// DeviceDTO represents a device data transfer object
type DeviceDTO struct {
	Hash     string `json:"hash"`
	Name     string `json:"name"`
	OS       OS     `json:"os"`
	UserHash string `json:"userHash"`
}

// ListDTO represents a list data transfer object
type ListDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	UserHash    string `json:"userHash"`
	ID          int    `json:"id"`
}

// ScheduleDTO represents a schedule data transfer object
type ScheduleDTO struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Name      string `json:"name"`
	UserHash  string `json:"userHash"`
	Monday    bool   `json:"monday"`
	Tuesday   bool   `json:"tuesday"`
	Wednesday bool   `json:"wednesday"`
	Thursday  bool   `json:"thursday"`
	Friday    bool   `json:"friday"`
	Saturday  bool   `json:"saturday"`
	Sunday    bool   `json:"sunday"`
	ID        int    `json:"id"`
}
