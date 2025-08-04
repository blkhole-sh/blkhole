package repos

import (
	"context"
	"database/sql"
	"server/internal/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

type DeviceRepo interface {
	Create(d *model.Device) error
	Update(hash string, d *model.Device) error
	Delete(hash string) error
	LinkSchedule(hash string, scheduleID int) error
	LoadScheduleIDs(hash string) ([]int, error)
	FindByHash(hash string) (*model.Device, error)
	FindByUser(userHash string) ([]*model.Device, error)
	FindBySchedule(scheduleID int) ([]*model.Device, error)
}

// DeviceRepoImpl implements the DeviceRepo interface
type DeviceRepoImpl struct {
	db  *sql.DB
	ctx context.Context
}

// NewDeviceRepo creates a new DeviceRepo instance
func NewDeviceRepo(db *sql.DB) DeviceRepo {
	return &DeviceRepoImpl{db: db, ctx: context.Background()}
}

// Create stores a new device into the database
func (repo *DeviceRepoImpl) Create(d *model.Device) error {
	sql := "INSERT INTO device (hash, name, os, user_hash) VALUES (?, ?, ?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, d.Hash, d.Name, d.OS, d.UserHash)
	return err
}

// Update modifies an existing device given hash in the database
func (repo *DeviceRepoImpl) Update(hash string, d *model.Device) error {
	sql := "UPDATE device SET name=?, os=? WHERE hash=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, d.Hash, d.OS, hash)
	return err
}

// Delete removes an existing device given hash from the database
func (repo *DeviceRepoImpl) Delete(hash string) error {
	sql := "DELETE FROM device WHERE hash=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, hash)
	return err
}

// LinkSchedule links a schedule with given ID to a device with given hash
func (repo *DeviceRepoImpl) LinkSchedule(hash string, scheduleID int) error {
	sql := "INSERT INTO device_schedule (device_hash, schedule_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, hash, scheduleID)
	return err
}

// LoadScheduleIDs returns ids of all schedules linked to device with given hash
func (repo *DeviceRepoImpl) LoadScheduleIDs(hash string) ([]int, error) {
	sql := "SELECT schedule_id FROM device_schedule WHERE device_hash = ?"
	var scheduleIDs []int

	if err := sqlscan.Select(repo.ctx, repo.db, &scheduleIDs, sql, hash); err != nil {
		return []int{}, nil
	}

	// Ensure we return empty slice instead of nil
	if scheduleIDs == nil {
		return []int{}, nil
	}

	return scheduleIDs, nil
}

// FindByHash returns an existing device with given hash from the database
func (repo *DeviceRepoImpl) FindByHash(hash string) (*model.Device, error) {
	sql := "SELECT hash, name, os, user_hash FROM device WHERE hash=?"
	var d model.Device

	err := sqlscan.Get(repo.ctx, repo.db, &d, sql, hash)
	if err != nil {
		return nil, err
	}

	if d.ScheduleIds, err = repo.LoadScheduleIDs(d.Hash); err != nil {
		return nil, err
	}

	return &d, nil
}

// FindByUser returns all existing devices with given user hash from the database
func (repo *DeviceRepoImpl) FindByUser(userHash string) ([]*model.Device, error) {
	sql := "SELECT hash, name, os, user_hash FROM device WHERE user_hash=?"
	var devices []*model.Device

	err := sqlscan.Select(repo.ctx, repo.db, &devices, sql, userHash)
	if err != nil {
		return []*model.Device{}, nil
	}

	// Ensure we return empty slice instead of nil
	if devices == nil {
		return []*model.Device{}, nil
	}

	for _, d := range devices {
		if d.ScheduleIds, err = repo.LoadScheduleIDs(d.Hash); err != nil {
			return nil, err
		}
	}

	return devices, nil
}

func (repo *DeviceRepoImpl) FindBySchedule(scheduleID int) ([]*model.Device, error) {
	sql := "SELECT DISTINCT d.hash, d.name, d.os, d.user_hash FROM device d JOIN device_schedule ds ON d.hash = ds.device_hash WHERE ds.schedule_id = ?"
	var devices []*model.Device

	err := sqlscan.Select(repo.ctx, repo.db, &devices, sql, scheduleID)
	if err != nil {
		return []*model.Device{}, nil
	}

	// Ensure we return empty slice instead of nil
	if devices == nil {
		return []*model.Device{}, nil
	}

	for _, d := range devices {
		if d.ScheduleIds, err = repo.LoadScheduleIDs(d.Hash); err != nil {
			return nil, err
		}
	}

	return devices, nil
}
