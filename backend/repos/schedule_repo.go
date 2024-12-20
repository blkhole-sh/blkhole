package repos

import (
	"context"
	"database/sql"
	"server/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// Define ScheduleRepo interface
type ScheduleRepo interface {
	Create(s *model.Schedule) (int, error)
	Update(id int, s *model.Schedule) error
	Delete(id int) error
	LinkDevice(id int, deviceHash string) error
	LinkDomain(id int, domainId int) error
	LinkList(id int, listId int) error
	LoadDeviceHashes(id int) ([]string, error)
	LoadDomainIds(id int) ([]int, error)
	LoadListIds(id int) ([]int, error)
	LoadRelations(s *model.Schedule) error
	FindById(id int) (*model.Schedule, error)
	FindByUser(userHash string) ([]*model.Schedule, error)
	FindByDevice(deviceHash string) ([]*model.Schedule, error)
	FindByDomain(domainId int) ([]*model.Schedule, error)
	FindByList(listId int) ([]*model.Schedule, error)
	DomainBlocked(domain string, deviceHash string) (bool, error)
}

// Define ScheduleRepoImpl struct
type ScheduleRepoImpl struct {
	db  *sql.DB
	ctx context.Context
}

// Create new ScheduleRepoImpl
func NewScheduleRepo(db *sql.DB) ScheduleRepo {
	return &ScheduleRepoImpl{db: db, ctx: context.Background()}
}

// Store a new schedule into the db
func (repo *ScheduleRepoImpl) Create(s *model.Schedule) (int, error) {
	sql := "INSERT INTO schedule (name, start_time, end_time, monday, tuesday, wednesday, thursday, friday, saturday, sunday, user_hash) VALUES  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id"
	err := repo.db.QueryRowContext(repo.ctx, sql, s.Name, s.StartTime, s.EndTime, s.Monday, s.Tuesday, s.Wednesday, s.Thursday, s.Friday, s.Saturday, s.Sunday, s.UserHash).Scan(&s.Id)
	return s.Id, err
}

// Update an existing schedule with given id in the db
func (repo *ScheduleRepoImpl) Update(id int, s *model.Schedule) error {
	sql := `UPDATE schedule SET name=?, startTime=? endTime=?, monday=?, tuesday=?, 
  wednesday=?, thursday=?, friday=?, saturday=?, sunday=? user_hash=? WHERE id=?`
	_, err := repo.db.ExecContext(repo.ctx, sql, s.Name, s.StartTime, s.EndTime, s.Monday, s.Tuesday, s.Wednesday, s.Thursday, s.Friday, s.Saturday, s.Sunday, s.UserHash, id)
	return err
}

// Delete an existing schedule with given id from the db
func (repo *ScheduleRepoImpl) Delete(id int) error {
	sql := "DELETE FROM schedule WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, id)
	return err
}

// Link a device with given hash to a schedule with given id
func (repo *ScheduleRepoImpl) LinkDevice(id int, deviceHash string) error {
	sql := "INSERT INTO device_schedule (device_hash, schedule_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, deviceHash, id)
	return err
}

// Link a domain with given id to a schedule with given id
func (repo *ScheduleRepoImpl) LinkDomain(id int, domainId int) error {
	sql := "INSERT INTO domain_schedule (domain_id, schedule_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, domainId, id)
	return err
}

// Link a list with given id to a schedule with given id
func (repo *ScheduleRepoImpl) LinkList(id int, list_id int) error {
	sql := "INSERT INTO list_schedule (list_id, schedule_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, list_id, id)
	return err
}

// Load hashes of all devices linked to schedule with given id
func (repo *ScheduleRepoImpl) LoadDeviceHashes(id int) ([]string, error) {
	sql := "SELECT DISTINCT ds.device_hash FROM device_schedule ds JOIN schedule s ON ds.schedule_id = s.id WHERE s.id = ?"
	var deviceHashes []string

	if err := sqlscan.Select(repo.ctx, repo.db, &deviceHashes, sql, id); err != nil {
		return nil, err
	}

	return deviceHashes, nil
}

// Load ids of all domains linked to schedule with given id
func (repo *ScheduleRepoImpl) LoadDomainIds(id int) ([]int, error) {
	sql := "SELECT DISTINCT ds.domain_id FROM domain_schedule ds JOIN schedule s ON ds.schedule_id = s.id WHERE s.id = ?"
	var domainIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &domainIds, sql, id); err != nil {
		return nil, err
	}

	return domainIds, nil
}

// Load ids of all lists linked to schedule with given id
func (repo *ScheduleRepoImpl) LoadListIds(id int) ([]int, error) {
	sql := "SELECT DISTINCT ls.list_id FROM list_schedule ls JOIN schedule s ON ls.schedule_id = s.id WHERE s.id = ?"
	var listIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &listIds, sql, id); err != nil {
		return nil, err
	}

	return listIds, nil
}

// Load all relations (hashes, ids) for scheduke with given id
func (repo *ScheduleRepoImpl) LoadRelations(s *model.Schedule) error {
	var err error

	if s.DeviceHashes, err = repo.LoadDeviceHashes(s.Id); err != nil {
		return err
	}

	if s.DomainIds, err = repo.LoadDomainIds(s.Id); err != nil {
		return err
	}

	if s.ListIds, err = repo.LoadListIds(s.Id); err != nil {
		return err
	}

	return nil
}

// Find an existing schedules with given id in the db
func (repo *ScheduleRepoImpl) FindById(id int) (*model.Schedule, error) {
	sql := "SELECT id, name, start_time, end_time, monday, tuesday, wednesday, thursday, friday, saturday, sunday FROM schedule WHERE id=?"
	var s model.Schedule

	if err := sqlscan.Get(repo.ctx, repo.db, &s, sql, id); err != nil {
		return nil, err
	}

	if err := repo.LoadRelations(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

// Find all existing schedules with given user hash in the db
func (repo *ScheduleRepoImpl) FindByUser(userHash string) ([]*model.Schedule, error) {
	sql := "SELECT id, name, start_time, end_time, monday, tuesday, wednesday, thursday, friday, saturday, sunday, user_hash FROM schedule WHERE user_hash=?"
	var schedules []*model.Schedule

	if err := sqlscan.Select(repo.ctx, repo.db, &schedules, sql, userHash); err != nil {
		return nil, err
	}

	for _, s := range schedules {
		if err := repo.LoadRelations(s); err != nil {
			return nil, err
		}
	}

	return schedules, nil
}

// Find all existing schedules that are asigned to device with given hash in the db
func (repo *ScheduleRepoImpl) FindByDevice(deviceHash string) ([]*model.Schedule, error) {
	sql := `SELECT DISTINCT s.id, s.name, s.start_time, s.end_time, s.monday, s.tuesday, s.wednesday, s.thursday, s.friday, s.saturday, s.sunday, s.user_hash 
          FROM device_schedule ds JOIN schedule s ON ds.schedule_id = s.id WHERE ds.device_hash = ?`
	var schedules []*model.Schedule

	if err := sqlscan.Select(repo.ctx, repo.db, &schedules, sql, deviceHash); err != nil {
		return nil, err
	}

	for _, s := range schedules {
		if err := repo.LoadRelations(s); err != nil {
			return nil, err
		}
	}

	return schedules, nil
}

// Find all existing schedules that are linked with domain with given id in the db
func (repo *ScheduleRepoImpl) FindByDomain(domainId int) ([]*model.Schedule, error) {
	sql := `SELECT DISTINCT s.id, s.name, s.start_time, s.end_time, s.monday, s.tuesday, s.wednesday, s.thursday, s.friday, s.saturday, s.sunday, s.user_hash
          FROM domain_schedule ds JOIN schedule s ON ds.schedule_id = s.id WHERE ds.domain_id = ?`
	var schedules []*model.Schedule

	if err := sqlscan.Select(repo.ctx, repo.db, &schedules, sql, domainId); err != nil {
		return nil, err
	}

	for _, s := range schedules {
		if err := repo.LoadRelations(s); err != nil {
			return nil, err
		}
	}

	return schedules, nil
}

// Find all existing schedules that are linked with list with given id in the db
func (repo *ScheduleRepoImpl) FindByList(listId int) ([]*model.Schedule, error) {
	sql := `SELECT DISTINCT s.id, s.name, s.start_time, s.end_time, s.monday, s.tuesday, s.wednesday, s.thursday, s.friday, s.saturday, s.sunday, s.user_hash
          FROM list_schedule ls JOIN schedule s ON ls.schedule_id = s.id WHERE ls.list_id = ?`
	var schedules []*model.Schedule

	if err := sqlscan.Select(repo.ctx, repo.db, &schedules, sql, listId); err != nil {
		return nil, err
	}

	for _, s := range schedules {
		if err := repo.LoadRelations(s); err != nil {
			return nil, err
		}
	}

	return schedules, nil
}

func (repo *ScheduleRepoImpl) DomainBlocked(domain string, deviceHash string) (bool, error) {
	query := `WITH CurrentDateTime AS (
              SELECT 
                time('now', 'localtime') AS current_time_only,  -- Current local time in HH:MM
                strftime('%w', 'now', 'localtime') AS current_day  -- Current day of the week (0 = Sunday, 1 = Monday, ...)
            )
            SELECT DISTINCT d.name AS domain_name
            FROM (
              -- Domains directly blocked via schedule linked with the device
              SELECT d.name
              FROM domain d
              JOIN schedule s ON s.id = s.id  -- Replace with actual join condition for schedule
              JOIN device_schedule ds ON s.id = ds.schedule_id
              JOIN device dv ON ds.device_hash = dv.hash
              CROSS JOIN CurrentDateTime
              WHERE 
                CurrentDateTime.current_time_only BETWEEN s.start_time AND s.end_time
              AND (
                (s.monday = 1 AND CurrentDateTime.current_day = '1') OR
                (s.tuesday = 1 AND CurrentDateTime.current_day = '2') OR
                (s.wednesday = 1 AND CurrentDateTime.current_day = '3') OR
                (s.thursday = 1 AND CurrentDateTime.current_day = '4') OR
                (s.friday = 1 AND CurrentDateTime.current_day = '5') OR
                (s.saturday = 1 AND CurrentDateTime.current_day = '6') OR
                (s.sunday = 1 AND CurrentDateTime.current_day = '0')
              )
              AND dv.hash = ?  -- Bind the device hash here
              UNION
              -- Domains indirectly blocked via lists in schedule linked with the device
              SELECT d.name
              FROM domain d
              JOIN domain_list dl ON d.id = dl.domain_id
              JOIN schedule s ON s.id = s.id  -- Replace with actual join condition for schedule
              JOIN device_schedule ds ON s.id = ds.schedule_id
              JOIN device dv ON ds.device_hash = dv.hash
              CROSS JOIN CurrentDateTime
              WHERE 
                CurrentDateTime.current_time_only BETWEEN s.start_time AND s.end_time
              AND (
                (s.monday = 1 AND CurrentDateTime.current_day = '1') OR
                (s.tuesday = 1 AND CurrentDateTime.current_day = '2') OR
                (s.wednesday = 1 AND CurrentDateTime.current_day = '3') OR
                (s.thursday = 1 AND CurrentDateTime.current_day = '4') OR
                (s.friday = 1 AND CurrentDateTime.current_day = '5') OR
                (s.saturday = 1 AND CurrentDateTime.current_day = '6') OR
                (s.sunday = 1 AND CurrentDateTime.current_day = '0')
              )
              AND dv.hash = ?  -- Bind the device hash here
            ) AS blocked_domains
            WHERE domain_name = ?;  -- Bind the domain name here`

	var result string
	err := repo.db.QueryRowContext(repo.ctx, query, deviceHash, deviceHash, domain).Scan(&result)

	// Domain is not blocked
	if err == sql.ErrNoRows {
		return false, nil
	}

	// An actual error occured
	if err != nil {
		return false, err
	}

	// Domain is blocked
	return true, nil
}
