package model

// UserDTO represents a user data transfer object
type UserDTO struct {
	Hash         string `json:"hash"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Devices      int    `json:"devices"`
	Lists        int    `json:"lists"`
	Schedules    int    `json:"schedules"`
}

// DeviceDTO represents a device data transfer object
type DeviceDTO struct {
	Hash      string `json:"hash"`
	Name      string `json:"name"`
	OS        OS     `json:"os"`
	UserHash  string `json:"userHash"`
	Schedules int    `json:"schedules"`
}

// ListDTO represents a list data transfer object
type ListDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	UserHash    string `json:"userHash"`
	Rules       int    `json:"rules"`
	Schedules   int    `json:"schedules"`
}

// ScheduleDTO represents a schedule data transfer object
type ScheduleDTO struct {
	ID        int    `json:"id"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Name      string `json:"name"`
	UserHash  string `json:"userHash"`
	Devices   int    `json:"devices"`
	Domains   int    `json:"domains"`
	Lists     int    `json:"lists"`
	Monday    bool   `json:"monday"`
	Tuesday   bool   `json:"tuesday"`
	Wednesday bool   `json:"wednesday"`
	Thursday  bool   `json:"thursday"`
	Friday    bool   `json:"friday"`
	Saturday  bool   `json:"saturday"`
	Sunday    bool   `json:"sunday"`
}

// ToDTO converts a User model to UserDTO with counts
func (u *User) ToDTO() UserDTO {
	return UserDTO{
		Hash:      u.Hash,
		Name:      u.Name,
		Email:     u.Email,
		Devices:   len(u.DeviceHashes),
		Lists:     len(u.ListIds),
		Schedules: len(u.ScheduleIds),
	}
}

// ToDTO converts a Device model to DeviceDTO with counts
func (d *Device) ToDTO() DeviceDTO {
	return DeviceDTO{
		Hash:      d.Hash,
		Name:      d.Name,
		OS:        d.OS,
		UserHash:  d.UserHash,
		Schedules: len(d.ScheduleIds),
	}
}

// ToDTO converts a List model to ListDTO with counts
func (l *List) ToDTO() ListDTO {
	return ListDTO{
		ID:          l.ID,
		Name:        l.Name,
		Description: l.Description,
		Source:      l.Source,
		UserHash:    l.UserHash,
		Rules:       len(l.Rules),
		Schedules:   len(l.ScheduleIds),
	}
}

// ToDTO converts a Schedule model to ScheduleDTO with counts
func (s *Schedule) ToDTO() ScheduleDTO {
	return ScheduleDTO{
		ID:        s.ID,
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
		Name:      s.Name,
		UserHash:  s.UserHash,
		Devices:   len(s.DeviceHashes),
		Domains:   len(s.Domains),
		Lists:     len(s.ListIds),
		Monday:    s.Monday,
		Tuesday:   s.Tuesday,
		Wednesday: s.Wednesday,
		Thursday:  s.Thursday,
		Friday:    s.Friday,
		Saturday:  s.Saturday,
		Sunday:    s.Sunday,
	}
}
