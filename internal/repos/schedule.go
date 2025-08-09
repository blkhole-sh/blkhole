package repos

import (
	"context"
	"database/sql"

	"github.com/lemon3studio/leo/internal/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// ScheduleRepo defines the interface for schedule repository operations
type ScheduleRepo interface {
	Create(s *model.Schedule) (int, error)
	Update(id int, s *model.Schedule) error
	Delete(id int) error
	LinkDevice(id int, deviceHash string) error
	LinkRule(id int, ruleID int) error
	LinkList(id int, listID int) error
	LoadDeviceHashes(id int) ([]string, error)
	LoadRuleIDs(id int) ([]int, error)
	LoadListIDs(id int) ([]int, error)
	LoadRelations(s *model.Schedule) error
	FindByID(id int) (*model.Schedule, error)
	FindByUser(userHash string) ([]*model.Schedule, error)
	FindByDevice(deviceHash string) ([]*model.Schedule, error)
	FindByRule(ruleID int) ([]*model.Schedule, error)
	FindByList(listID int) ([]*model.Schedule, error)
	DomainBlocked(domain string, deviceHash string) (bool, error)
}

// scheduleRepo implements the ScheduleRepo interface
type scheduleRepo struct {
	db  *sql.DB
	ctx context.Context
}

// NewScheduleRepo creates a new ScheduleRepo instance
func NewScheduleRepo(db *sql.DB) ScheduleRepo {
	return &scheduleRepo{db: db, ctx: context.Background()}
}

// encodeDaysToInt converts boolean weekday fields to integer encoding
// Bit positions: Monday=0, Tuesday=1, Wednesday=2, Thursday=3, Friday=4, Saturday=5, Sunday=6
func encodeDaysToInt(schedule *model.Schedule) int {
	days := 0
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
	return days
}

// decodeDaysFromInt converts integer encoding to boolean weekday fields
// Bit positions: Monday=0, Tuesday=1, Wednesday=2, Thursday=3, Friday=4, Saturday=5, Sunday=6
func decodeDaysFromInt(days int, schedule *model.Schedule) {
	schedule.Monday = (days & (1 << 0)) != 0
	schedule.Tuesday = (days & (1 << 1)) != 0
	schedule.Wednesday = (days & (1 << 2)) != 0
	schedule.Thursday = (days & (1 << 3)) != 0
	schedule.Friday = (days & (1 << 4)) != 0
	schedule.Saturday = (days & (1 << 5)) != 0
	schedule.Sunday = (days & (1 << 6)) != 0
}

// scheduleDbRow represents a schedule row from the database with integer days encoding
type dbSchedule struct {
	ID        int    `db:"id"`
	Name      string `db:"name"`
	StartTime string `db:"start_time"`
	EndTime   string `db:"end_time"`
	Days      int    `db:"days"`
	UserHash  string `db:"user_hash"`
}

// toSchedule converts a scheduleDbRow to a model.Schedule
func (row *dbSchedule) toSchedule() *model.Schedule {
	s := &model.Schedule{
		ID:        row.ID,
		Name:      row.Name,
		StartTime: row.StartTime,
		EndTime:   row.EndTime,
		UserHash:  row.UserHash,
	}
	decodeDaysFromInt(row.Days, s)
	return s
}

// Create stores a new schedule into the database
func (repo *scheduleRepo) Create(s *model.Schedule) (int, error) {
	days := encodeDaysToInt(s)
	sql := "INSERT INTO schedule (name, start_time, end_time, days, user_hash) VALUES (?, ?, ?, ?, ?) RETURNING id"
	err := repo.db.QueryRowContext(repo.ctx, sql, s.Name, s.StartTime, s.EndTime, days, s.UserHash).Scan(&s.ID)
	return s.ID, err
}

// Update modifies an existing schedule with given ID in the database
func (repo *scheduleRepo) Update(id int, s *model.Schedule) error {
	days := encodeDaysToInt(s)
	sql := "UPDATE schedule SET name=?, start_time=?, end_time=?, days=?, user_hash=? WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, s.Name, s.StartTime, s.EndTime, days, s.UserHash, id)
	return err
}

// Delete removes an existing schedule with given ID from the database
func (repo *scheduleRepo) Delete(id int) error {
	sql := "DELETE FROM schedule WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, id)
	return err
}

// LinkDevice links a device with given hash to a schedule with given ID
func (repo *scheduleRepo) LinkDevice(id int, deviceHash string) error {
	sql := "INSERT INTO device_schedule (device_hash, schedule_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, deviceHash, id)
	return err
}

// LinkRule links a rule with given ID to a schedule with given ID
func (repo *scheduleRepo) LinkRule(id int, ruleID int) error {
	sql := "INSERT INTO schedule_rule (schedule_id, rule_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, id, ruleID)
	return err
}

// LinkList links a list with given ID to a schedule with given ID
func (repo *scheduleRepo) LinkList(id int, listID int) error {
	sql := "INSERT INTO list_schedule (list_id, schedule_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, listID, id)
	return err
}

// LoadDeviceHashes returns hashes of all devices linked to schedule with given id
func (repo *scheduleRepo) LoadDeviceHashes(id int) ([]string, error) {
	sql := "SELECT DISTINCT ds.device_hash FROM device_schedule ds JOIN schedule s ON ds.schedule_id = s.id WHERE s.id = ?"
	var deviceHashes []string

	if err := sqlscan.Select(repo.ctx, repo.db, &deviceHashes, sql, id); err != nil {
		return []string{}, nil
	}

	// Ensure we return empty slice instead of nil
	if deviceHashes == nil {
		return []string{}, nil
	}

	return deviceHashes, nil
}

// LoadRuleIDs returns rule IDs linked to schedule with given id
func (repo *scheduleRepo) LoadRuleIDs(id int) ([]int, error) {
	sql := "SELECT DISTINCT sr.rule_id FROM schedule_rule sr WHERE sr.schedule_id = ?"
	var ruleIDs []int

	if err := sqlscan.Select(repo.ctx, repo.db, &ruleIDs, sql, id); err != nil {
		return []int{}, nil
	}

	// Ensure we return empty slice instead of nil
	if ruleIDs == nil {
		return []int{}, nil
	}

	return ruleIDs, nil
}

// LoadListIDs returns ids of all lists linked to schedule with given id
func (repo *scheduleRepo) LoadListIDs(id int) ([]int, error) {
	sql := "SELECT DISTINCT ls.list_id FROM list_schedule ls JOIN schedule s ON ls.schedule_id = s.id WHERE s.id = ?"
	var listIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &listIds, sql, id); err != nil {
		return []int{}, nil
	}

	// Ensure we return empty slice instead of nil
	if listIds == nil {
		return []int{}, nil
	}

	return listIds, nil
}

// LoadRelations loads all relations (hashes, domains, ids) for schedule with given id
func (repo *scheduleRepo) LoadRelations(s *model.Schedule) error {
	var err error

	if s.DeviceHashes, err = repo.LoadDeviceHashes(s.ID); err != nil {
		return err
	}

	if s.RuleIDs, err = repo.LoadRuleIDs(s.ID); err != nil {
		return err
	}

	if s.ListIds, err = repo.LoadListIDs(s.ID); err != nil {
		return err
	}

	return nil
}

// FindByID returns an existing schedule with given id from the database
func (repo *scheduleRepo) FindByID(id int) (*model.Schedule, error) {
	sql := "SELECT id, name, start_time, end_time, days, user_hash FROM schedule WHERE id=?"
	var s model.Schedule
	var days int

	row := repo.db.QueryRowContext(repo.ctx, sql, id)
	if err := row.Scan(&s.ID, &s.Name, &s.StartTime, &s.EndTime, &days, &s.UserHash); err != nil {
		return nil, err
	}

	decodeDaysFromInt(days, &s)

	if err := repo.LoadRelations(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

// FindByUser returns all existing schedules with given user hash from the database
func (repo *scheduleRepo) FindByUser(userHash string) ([]*model.Schedule, error) {
	sql := "SELECT id, name, start_time, end_time, days, user_hash FROM schedule WHERE user_hash=?"
	var dbRows []dbSchedule

	if err := sqlscan.Select(repo.ctx, repo.db, &dbRows, sql, userHash); err != nil {
		return []*model.Schedule{}, nil
	}

	// Ensure we return empty slice instead of nil
	if dbRows == nil {
		return []*model.Schedule{}, nil
	}

	var schedules []*model.Schedule
	for _, row := range dbRows {
		s := row.toSchedule()
		if err := repo.LoadRelations(s); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}

	return schedules, nil
}

// FindByDevice returns all existing schedules that are assigned to device with given hash
func (repo *scheduleRepo) FindByDevice(deviceHash string) ([]*model.Schedule, error) {
	sql := `SELECT DISTINCT s.id, s.name, s.start_time, s.end_time, s.days, s.user_hash 
          FROM device_schedule ds JOIN schedule s ON ds.schedule_id = s.id WHERE ds.device_hash = ?`
	var dbRows []dbSchedule

	if err := sqlscan.Select(repo.ctx, repo.db, &dbRows, sql, deviceHash); err != nil {
		return []*model.Schedule{}, nil
	}

	// Ensure we return empty slice instead of nil
	if dbRows == nil {
		return []*model.Schedule{}, nil
	}

	var schedules []*model.Schedule
	for _, row := range dbRows {
		s := row.toSchedule()
		if err := repo.LoadRelations(s); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}

	return schedules, nil
}

// FindByRule returns all existing schedules that are linked with rule with given ID
func (repo *scheduleRepo) FindByRule(ruleID int) ([]*model.Schedule, error) {
	sql := `SELECT DISTINCT s.id, s.name, s.start_time, s.end_time, s.days, s.user_hash
          FROM schedule_rule sr JOIN schedule s ON sr.schedule_id = s.id WHERE sr.rule_id = ?`
	var dbRows []dbSchedule

	if err := sqlscan.Select(repo.ctx, repo.db, &dbRows, sql, ruleID); err != nil {
		return []*model.Schedule{}, nil
	}

	// Ensure we return empty slice instead of nil
	if dbRows == nil {
		return []*model.Schedule{}, nil
	}

	var schedules []*model.Schedule
	for _, row := range dbRows {
		s := row.toSchedule()
		if err := repo.LoadRelations(s); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}

	return schedules, nil
}

// FindByList returns all existing schedules that are linked with list with given id
func (repo *scheduleRepo) FindByList(listID int) ([]*model.Schedule, error) {
	sql := `SELECT DISTINCT s.id, s.name, s.start_time, s.end_time, s.days, s.user_hash
          FROM list_schedule ls JOIN schedule s ON ls.schedule_id = s.id WHERE ls.list_id = ?`
	var dbRows []dbSchedule

	if err := sqlscan.Select(repo.ctx, repo.db, &dbRows, sql, listID); err != nil {
		return []*model.Schedule{}, nil
	}

	// Ensure we return empty slice instead of nil
	if dbRows == nil {
		return []*model.Schedule{}, nil
	}

	var schedules []*model.Schedule
	for _, row := range dbRows {
		s := row.toSchedule()
		if err := repo.LoadRelations(s); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}

	return schedules, nil
}

func (repo *scheduleRepo) DomainBlocked(domain string, deviceHash string) (bool, error) {
	query := `WITH CurrentDateTime AS (
              SELECT 
                time('now', 'localtime') AS current_time_only,
                (CAST(strftime('%w', 'now', 'localtime') AS INTEGER) + 6) % 7 AS current_day
            )
            SELECT COUNT(*) as blocked_count
            FROM (
              -- Domains blocked via rules directly linked to schedule
              SELECT r.domain
              FROM rule r
              JOIN schedule_rule sr ON r.id = sr.rule_id
              JOIN schedule s ON sr.schedule_id = s.id
              JOIN device_schedule ds ON s.id = ds.schedule_id
              CROSS JOIN CurrentDateTime
              WHERE 
                r.allowed = 0
                AND r.domain = ?
                AND CurrentDateTime.current_time_only BETWEEN s.start_time AND s.end_time
                AND (s.days & (1 << CAST(CurrentDateTime.current_day AS INTEGER))) != 0
                AND ds.device_hash = ?
              UNION
              -- Domains blocked via rules in lists linked to schedule
              SELECT r.domain
              FROM rule r
              JOIN list_rule lr ON r.id = lr.rule_id
              JOIN list l ON lr.list_id = l.id
              JOIN list_schedule ls ON l.id = ls.list_id
              JOIN schedule s ON ls.schedule_id = s.id
              JOIN device_schedule ds ON s.id = ds.schedule_id
              CROSS JOIN CurrentDateTime
              WHERE 
                r.allowed = 0
                AND r.domain = ?
                AND CurrentDateTime.current_time_only BETWEEN s.start_time AND s.end_time
                AND (s.days & (1 << CAST(CurrentDateTime.current_day AS INTEGER))) != 0
                AND ds.device_hash = ?
            )`

	var count int
	err := repo.db.QueryRowContext(repo.ctx, query, domain, deviceHash, domain, deviceHash).Scan(&count)
	
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
