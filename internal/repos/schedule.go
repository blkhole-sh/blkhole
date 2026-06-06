package repos

import (
	"context"
	"database/sql"

	"github.com/lemon3studio/blkhole/internal/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// ScheduleRepo defines the interface for schedule repository operations
type ScheduleRepo interface {
	Create(s *model.Schedule) (int, error)
	Update(id int, s *model.Schedule) error
	Delete(id int) error
	LinkDevice(scheduleID int, deviceID int) error
	LinkRule(id int, ruleID int) error
	LinkList(id int, listID int) error
	LoadDeviceIDs(id int) ([]int, error)
	LoadRuleIDs(id int) ([]int, error)
	LoadListIDs(id int) ([]int, error)
	LoadRelations(s *model.Schedule) error
	FindByID(id int) (*model.Schedule, error)
	FindByUser(userID int) ([]*model.Schedule, error)
	FindNamesByDeviceID(deviceID int) ([]string, error)
	FindByDevice(deviceHash string) ([]*model.Schedule, error)
	FindByRule(ruleID int) ([]*model.Schedule, error)
	FindByList(listID int) ([]*model.Schedule, error)
	FindAll() ([]*model.Schedule, error)
	FindScheduleRule() ([]*model.ScheduleRule, error)
	HasDefaultsForUser(userID int) (bool, error)
	SeedDefaults(userID int) error
}

// scheduleRepo implements the Schedulesr interface
type scheduleRepo struct {
	db  *sql.DB
	ctx context.Context
}

// NewScheduleRepo creates a new ScheduleRepo instance
func NewScheduleRepo(db *sql.DB) ScheduleRepo {
	return &scheduleRepo{db: db, ctx: context.Background()}
}

// encodeDaysToInt converts boolean days to integer bitmask for database storage
func encodeDaysToInt(schedule *model.Schedule) int {
	var days uint8
	if schedule.Monday {
		days |= 1 << 0
	}
	if schedule.Tuesday {
		days |= 1 << 1
	}
	if schedule.Wednesday {
		days |= 1 << 2
	}
	if schedule.Thursday {
		days |= 1 << 3
	}
	if schedule.Friday {
		days |= 1 << 4
	}
	if schedule.Saturday {
		days |= 1 << 5
	}
	if schedule.Sunday {
		days |= 1 << 6
	}
	return int(days)
}

// decodeDaysFromInt converts integer bitmask to boolean day fields
func decodeDaysFromInt(days int, schedule *model.Schedule) {
	dayBits := uint8(days)
	schedule.Monday = dayBits&(1<<0) != 0
	schedule.Tuesday = dayBits&(1<<1) != 0
	schedule.Wednesday = dayBits&(1<<2) != 0
	schedule.Thursday = dayBits&(1<<3) != 0
	schedule.Friday = dayBits&(1<<4) != 0
	schedule.Saturday = dayBits&(1<<5) != 0
	schedule.Sunday = dayBits&(1<<6) != 0
}

// scheduleDbRow represents a schedule row from the database with integer days encoding
type dbSchedule struct {
	ID        int    `db:"id"`
	Name      string `db:"name"`
	StartTime string `db:"start_time"`
	EndTime   string `db:"end_time"`
	Days      int    `db:"days"`
	Active    bool   `db:"active"`
	UserID    int    `db:"user_id"`
	IsDefault bool   `db:"is_default"`
}

// toSchedule converts a scheduleDbRow to a model.Schedule
func (row *dbSchedule) toSchedule() *model.Schedule {
	s := &model.Schedule{
		ID:        row.ID,
		Name:      row.Name,
		StartTime: row.StartTime,
		EndTime:   row.EndTime,
		Active:    row.Active,
		UserID:    row.UserID,
		IsDefault: row.IsDefault,
	}
	decodeDaysFromInt(row.Days, s)
	return s
}

// Create stores a new schedule into the database
func (sr *scheduleRepo) Create(s *model.Schedule) (int, error) {
	days := encodeDaysToInt(s)
	isDefault := 0
	if s.IsDefault {
		isDefault = 1
	}
	query := "INSERT INTO schedule (name, start_time, end_time, days, active, user_id, is_default) VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id"
	err := sr.db.QueryRowContext(sr.ctx, query, s.Name, s.StartTime, s.EndTime, days, s.Active, s.UserID, isDefault).Scan(&s.ID)
	if err != nil {
		return 0, err
	}

	// Link devices
	for _, deviceID := range s.DeviceIDs {
		if err := sr.LinkDevice(s.ID, deviceID); err != nil {
			return 0, err
		}
	}

	// Link lists
	for _, listID := range s.ListIDs {
		if err := sr.LinkList(s.ID, listID); err != nil {
			return 0, err
		}
	}

	return s.ID, nil
}

// Update modifies an existing schedule with given ID in the database
func (sr *scheduleRepo) Update(id int, s *model.Schedule) error {
	days := encodeDaysToInt(s)
	query := "UPDATE schedule SET name=?, start_time=?, end_time=?, days=?, active=? WHERE id=?"
	_, err := sr.db.ExecContext(sr.ctx, query, s.Name, s.StartTime, s.EndTime, days, s.Active, id)
	if err != nil {
		return err
	}

	// Update device relationships
	// Delete existing device links
	deleteDevicesQuery := "DELETE FROM device_schedule WHERE schedule_id=?"
	if _, err := sr.db.ExecContext(sr.ctx, deleteDevicesQuery, id); err != nil {
		return err
	}
	// Insert new device links
	for _, deviceID := range s.DeviceIDs {
		if err := sr.LinkDevice(id, deviceID); err != nil {
			return err
		}
	}

	// Update list relationships
	// Delete existing list links
	deleteListsQuery := "DELETE FROM list_schedule WHERE schedule_id=?"
	if _, err := sr.db.ExecContext(sr.ctx, deleteListsQuery, id); err != nil {
		return err
	}
	// Insert new list links
	for _, listID := range s.ListIDs {
		if err := sr.LinkList(id, listID); err != nil {
			return err
		}
	}

	return nil
}

// Delete removes an existing schedule with given ID from the database
func (sr *scheduleRepo) Delete(id int) error {
	query := "DELETE FROM schedule WHERE id=?"
	_, err := sr.db.ExecContext(sr.ctx, query, id)
	return err
}

// LinkDevice links a device with given ID to a schedule with given ID
func (sr *scheduleRepo) LinkDevice(scheduleID int, deviceID int) error {
	query := "INSERT INTO device_schedule (device_id, schedule_id) VALUES (?, ?)"
	_, err := sr.db.ExecContext(sr.ctx, query, deviceID, scheduleID)
	return err
}

// LinkRule links a rule with given ID to a schedule with given ID
func (sr *scheduleRepo) LinkRule(id int, ruleID int) error {
	query := "INSERT INTO schedule_rule (schedule_id, rule_id) VALUES (?, ?)"
	_, err := sr.db.ExecContext(sr.ctx, query, id, ruleID)
	return err
}

// LinkList links a list with given ID to a schedule with given ID
func (sr *scheduleRepo) LinkList(id int, listID int) error {
	query := "INSERT INTO list_schedule (list_id, schedule_id) VALUES (?, ?)"
	_, err := sr.db.ExecContext(sr.ctx, query, listID, id)
	return err
}

// LoadDeviceIDs returns ids of all devices linked to schedule with given id
func (sr *scheduleRepo) LoadDeviceIDs(id int) ([]int, error) {
	query := "SELECT DISTINCT ds.device_id FROM device_schedule ds WHERE ds.schedule_id = ?"
	var deviceIDs []int

	err := sqlscan.Select(sr.ctx, sr.db, &deviceIDs, query, id)
	if err != nil {
		return nil, err
	}

	if deviceIDs == nil {
		return []int{}, nil
	}

	return deviceIDs, nil
}

// LoadRuleIDs returns rule IDs linked to schedule with given id
func (sr *scheduleRepo) LoadRuleIDs(id int) ([]int, error) {
	query := "SELECT DISTINCT sr.rule_id FROM schedule_rule sr WHERE sr.schedule_id = ?"
	var ruleIDs []int

	err := sqlscan.Select(sr.ctx, sr.db, &ruleIDs, query, id)
	if err != nil {
		return nil, err
	}

	if ruleIDs == nil {
		return []int{}, nil
	}

	return ruleIDs, nil
}

// LoadListIDs returns ids of all lists linked to schedule with given id
func (sr *scheduleRepo) LoadListIDs(id int) ([]int, error) {
	query := "SELECT DISTINCT ls.list_id FROM list_schedule ls JOIN schedule s ON ls.schedule_id = s.id WHERE s.id = ?"
	var listIDs []int

	err := sqlscan.Select(sr.ctx, sr.db, &listIDs, query, id)
	if err != nil {
		return nil, err
	}

	if listIDs == nil {
		return []int{}, nil
	}

	return listIDs, nil
}

// LoadRelations loads all relations (hashes, domains, ids) for schedule with given id
func (sr *scheduleRepo) LoadRelations(s *model.Schedule) error {
	var err error

	if s.DeviceIDs, err = sr.LoadDeviceIDs(s.ID); err != nil {
		return err
	}

	if s.RuleIDs, err = sr.LoadRuleIDs(s.ID); err != nil {
		return err
	}

	if s.ListIDs, err = sr.LoadListIDs(s.ID); err != nil {
		return err
	}

	return nil
}

// FindByID returns an existing schedule with given id from the database
func (sr *scheduleRepo) FindByID(id int) (*model.Schedule, error) {
	query := "SELECT id, name, start_time, end_time, days, active, user_id, is_default FROM schedule WHERE id=?"
	var s model.Schedule
	var days int
	var isDefault int

	row := sr.db.QueryRowContext(sr.ctx, query, id)
	if err := row.Scan(&s.ID, &s.Name, &s.StartTime, &s.EndTime, &days, &s.Active, &s.UserID, &isDefault); err != nil {
		return nil, err
	}

	s.IsDefault = isDefault == 1
	decodeDaysFromInt(days, &s)

	if err := sr.LoadRelations(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

// HasDefaultsForUser returns true if the user already has any default schedules.
func (sr *scheduleRepo) HasDefaultsForUser(userID int) (bool, error) {
	var count int
	err := sr.db.QueryRowContext(sr.ctx, "SELECT COUNT(*) FROM schedule WHERE user_id=? AND is_default=1", userID).Scan(&count)
	return count > 0, err
}

// SeedDefaults creates the default "Base Protection" schedule for a user if they don't already have one.
func (sr *scheduleRepo) SeedDefaults(userID int) error {
	has, err := sr.HasDefaultsForUser(userID)
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	s := &model.Schedule{
		Name:      "Always",
		StartTime: "00:00",
		EndTime:   "00:00",
		Active:    true,
		UserID:    userID,
		IsDefault: true,
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  true,
		Sunday:    true,
	}
	_, err = sr.Create(s)
	return err
}

// FindByUser returns all existing schedules with given user ID from the database
func (sr *scheduleRepo) FindByUser(userID int) ([]*model.Schedule, error) {
	query := "SELECT id, name, start_time, end_time, days, active, user_id, is_default FROM schedule WHERE user_id=?"
	var dbRows []dbSchedule

	if err := sqlscan.Select(sr.ctx, sr.db, &dbRows, query, userID); err != nil {
		return nil, err
	}

	// Ensure we return empty slice instead of nil
	if dbRows == nil {
		return []*model.Schedule{}, nil
	}

	var schedules []*model.Schedule
	for _, row := range dbRows {
		s := row.toSchedule()
		if err := sr.LoadRelations(s); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}

	return schedules, nil
}

// FindAll returns all existing schedules from the database
func (sr *scheduleRepo) FindAll() ([]*model.Schedule, error) {
	query := "SELECT id, name, start_time, end_time, days, user_id FROM schedule"
	var dbRows []dbSchedule

	if err := sqlscan.Select(sr.ctx, sr.db, &dbRows, query); err != nil {
		return nil, err
	}

	// Ensure we return empty slice instead of nil
	if dbRows == nil {
		return []*model.Schedule{}, nil
	}

	var schedules []*model.Schedule
	for _, row := range dbRows {
		s := row.toSchedule()
		if err := sr.LoadRelations(s); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}

	return schedules, nil
}

// FindNamesByDeviceID returns the names of all schedules assigned to the device with given ID
func (sr *scheduleRepo) FindNamesByDeviceID(deviceID int) ([]string, error) {
	query := `SELECT DISTINCT s.name FROM device_schedule ds JOIN schedule s ON ds.schedule_id = s.id WHERE ds.device_id = ?`
	var names []string

	if err := sqlscan.Select(sr.ctx, sr.db, &names, query, deviceID); err != nil {
		return nil, err
	}

	if names == nil {
		return []string{}, nil
	}

	return names, nil
}

// FindByDevice returns all existing schedules that are assigned to device with given hash
func (sr *scheduleRepo) FindByDevice(deviceHash string) ([]*model.Schedule, error) {
	// First get device ID from hash
	var deviceID int
	deviceQuery := "SELECT id FROM device WHERE hash = ?"
	if err := sr.db.QueryRowContext(sr.ctx, deviceQuery, deviceHash).Scan(&deviceID); err != nil {
		return nil, err
	}

	query := `SELECT DISTINCT s.id, s.name, s.start_time, s.end_time, s.days, s.user_id 
          FROM device_schedule ds JOIN schedule s ON ds.schedule_id = s.id WHERE ds.device_id = ?`
	var dbRows []dbSchedule

	if err := sqlscan.Select(sr.ctx, sr.db, &dbRows, query, deviceID); err != nil {
		return nil, err
	}

	// Ensure we return empty slice instead of nil
	if dbRows == nil {
		return []*model.Schedule{}, nil
	}

	var schedules []*model.Schedule
	for _, row := range dbRows {
		s := row.toSchedule()
		if err := sr.LoadRelations(s); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}

	return schedules, nil
}

// FindByRule returns all existing schedules that are linked with rule with given ID
func (sr *scheduleRepo) FindByRule(ruleID int) ([]*model.Schedule, error) {
	query := `SELECT DISTINCT s.id, s.name, s.start_time, s.end_time, s.days, s.user_id
          FROM schedule_rule sr JOIN schedule s ON sr.schedule_id = s.id WHERE sr.rule_id = ?`
	var dbRows []dbSchedule

	if err := sqlscan.Select(sr.ctx, sr.db, &dbRows, query, ruleID); err != nil {
		return nil, err
	}

	// Ensure we return empty slice instead of nil
	if dbRows == nil {
		return []*model.Schedule{}, nil
	}

	var schedules []*model.Schedule
	for _, row := range dbRows {
		s := row.toSchedule()
		if err := sr.LoadRelations(s); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}

	return schedules, nil
}

// FindByList returns all existing schedules that are linked with list with given id
func (sr *scheduleRepo) FindByList(listID int) ([]*model.Schedule, error) {
	query := `SELECT DISTINCT s.id, s.name, s.start_time, s.end_time, s.days, s.user_id
          FROM list_schedule ls JOIN schedule s ON ls.schedule_id = s.id WHERE ls.list_id = ?`
	var dbRows []dbSchedule

	if err := sqlscan.Select(sr.ctx, sr.db, &dbRows, query, listID); err != nil {
		return nil, err
	}

	// Ensure we return empty slice instead of nil
	if dbRows == nil {
		return []*model.Schedule{}, nil
	}

	var schedules []*model.Schedule
	for _, row := range dbRows {
		s := row.toSchedule()
		if err := sr.LoadRelations(s); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}

	return schedules, nil
}

// FindScheduleRule returns all existing entries from schedule rule many-to-many table
func (sr *scheduleRepo) FindScheduleRule() ([]*model.ScheduleRule, error) {
	query := `SELECT schedule_id, rule_id FROM schedule_rule
		        UNION
	          SELECT ls.schedule_id, lr.rule_id
	          FROM list_schedule ls
	          JOIN list_rule lr ON ls.list_id = lr.list_id
	          ORDER BY schedule_id, rule_id`

	var scheduleRule []*model.ScheduleRule

	err := sqlscan.Select(sr.ctx, sr.db, &scheduleRule, query)
	if err != nil {
		return nil, err
	}

	return scheduleRule, nil
}
