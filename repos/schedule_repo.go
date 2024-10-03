package repos

import (
	"database/sql"
	"server/model"

	_ "github.com/mattn/go-sqlite3"
)

// Define ScheduleRepo interface
type ScheduleRepo interface {
	Create(s *model.Schedule) error
	Update(s *model.Schedule) error
	Delete(s *model.Schedule) error
}

// Define ScheduleRepoImpl struct
type ScheduleRepoImpl struct {
	DB *sql.DB
}

// Create new ScheduleRepoImpl
func NewScheduleRepo(db *sql.DB) *ScheduleRepoImpl {
	return &ScheduleRepoImpl{}
}

// Store a new schedule into the db
func (repo *ScheduleRepoImpl) Create(s *model.Schedule) error {
	sql := "INSERT INTO Schedule VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) WHERE id=?"
	_, err := repo.DB.Exec(sql, s.DeviceHash, s.StartTime, s.EndTime, s.Description, s.Monday, s.Tuesday, s.Wednesday, s.Thursday, s.Friday, s.Saturday, s.Sunday, s.Id)
	return err
}

// Update an existing schedule in the db
func (repo *ScheduleRepoImpl) Update(s *model.Schedule) error {
	sql := `UPDATE Schedule SET deviceHash=?, startTime=? endTime=?, description=?, monday=?, 
  tuesday=?, wednesday=?, thursday=?, friday=?, saturday=?, sunday=? WHERE id=?`
	_, err := repo.DB.Exec(sql, s.DeviceHash, s.StartTime, s.EndTime, s.Description, s.Monday, s.Tuesday, s.Wednesday, s.Thursday, s.Friday, s.Saturday, s.Sunday, s.Id)
	return err
}

// Delete an existing schedule from the db
func (repo *ScheduleRepoImpl) Delete(s *model.Schedule) error {
	sql := "DELETE FROM Schedule WHERE id=?"
	_, err := repo.DB.Exec(sql, s.Id)
	return err
}
