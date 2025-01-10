package model

// Define UserDTO struct
type UserDTO struct {
	Hash         string `json:"hash"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

// Define DeviceDTO struct
type DeviceDTO struct {
	Hash     string `json:"hash"`
	Name     string `json:"name"`
	OS       OS     `json:"os"`
	UserHash string `json:"userHash"`
}

// Define ListDTO struct
type ListDTO struct {
	Name     string `json:"name"`
	UserHash string `json:"userHash"`
	Id       int    `json:"id"`
}

// Define ScheduleDTO struct
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
	Id        int    `json:"id"`
}
