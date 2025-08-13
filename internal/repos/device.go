package repos

import (
	"context"
	"database/sql"

	"github.com/lemon3studio/leo/internal/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// DeviceRepo defines the interface for device repository operations
type DeviceRepo interface {
	Create(d *model.Device) error
	Update(id int, d *model.Device) error
	Delete(id int) error
	LinkSchedule(id int, scheduleID int) error
	LoadScheduleIDs(id int) ([]int, error)
	FindByID(id int) (*model.Device, error)
	FindByHash(hash string) (*model.Device, error)
	FindByUser(userID int) ([]*model.Device, error)
	FindBySchedule(scheduleID int) ([]*model.Device, error)
	FindAll() ([]*model.Device, error)
	FindDeviceSchedule() ([]*model.DeviceSchedule, error)
}

// deviceRepo implements the DeviceRepo interface
type deviceRepo struct {
	db  *sql.DB
	ctx context.Context
}

// NewDeviceRepo creates a new DeviceRepo instance
func NewDeviceRepo(db *sql.DB) DeviceRepo {
	return &deviceRepo{db: db, ctx: context.Background()}
}

// Create stores a new device into the database
func (dr *deviceRepo) Create(d *model.Device) error {
	query := "INSERT INTO device (hash, name, os, user_id) VALUES (?, ?, ?, ?)"
	result, err := dr.db.ExecContext(dr.ctx, query, d.Hash, d.Name, d.OS, d.UserID)
	if err != nil {
		return err
	}

	// Get the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	d.ID = int(id)

	return nil
}

// Update modifies an existing device given id in the database
func (dr *deviceRepo) Update(id int, d *model.Device) error {
	query := "UPDATE device SET name=?, os=?, hash=? WHERE id=?"
	_, err := dr.db.ExecContext(dr.ctx, query, d.Name, d.OS, d.Hash, id)
	return err
}

// Delete removes an existing device given id from the database
func (dr *deviceRepo) Delete(id int) error {
	query := "DELETE FROM device WHERE id=?"
	_, err := dr.db.ExecContext(dr.ctx, query, id)
	return err
}

// LinkSchedule links a schedule with given ID to a device with given id
func (dr *deviceRepo) LinkSchedule(id int, scheduleID int) error {
	query := "INSERT INTO device_schedule (device_id, schedule_id) VALUES (?, ?)"
	_, err := dr.db.ExecContext(dr.ctx, query, id, scheduleID)
	return err
}

// LoadScheduleIDs returns ids of all schedules linked to device with given id
func (dr *deviceRepo) LoadScheduleIDs(id int) ([]int, error) {
	query := "SELECT schedule_id FROM device_schedule WHERE device_id = ?"
	var scheduleIDs []int

	err := sqlscan.Select(dr.ctx, dr.db, &scheduleIDs, query, id)

	if err != nil || scheduleIDs == nil {
		return []int{}, nil
	}

	return scheduleIDs, nil
}

// FindByID returns an existing device with given id from the database
func (dr *deviceRepo) FindByID(id int) (*model.Device, error) {
	query := "SELECT id, hash, name, os, user_id FROM device WHERE id=?"
	var d model.Device

	err := sqlscan.Get(dr.ctx, dr.db, &d, query, id)
	if err != nil {
		return nil, err
	}

	if d.ScheduleIDs, err = dr.LoadScheduleIDs(d.ID); err != nil {
		return nil, err
	}

	return &d, nil
}

// FindByHash returns an existing device with given hash from the database
func (dr *deviceRepo) FindByHash(hash string) (*model.Device, error) {
	query := "SELECT id, hash, name, os, user_id FROM device WHERE hash=?"
	var d model.Device

	err := sqlscan.Get(dr.ctx, dr.db, &d, query, hash)
	if err != nil {
		return nil, err
	}

	if d.ScheduleIDs, err = dr.LoadScheduleIDs(d.ID); err != nil {
		return nil, err
	}

	return &d, nil
}

// FindByUser returns all existing devices with given user id from the database
func (dr *deviceRepo) FindByUser(userID int) ([]*model.Device, error) {
	query := "SELECT id, hash, name, os, user_id FROM device WHERE user_id=?"
	var devices []*model.Device

	err := sqlscan.Select(dr.ctx, dr.db, &devices, query, userID)

	if err != nil || devices == nil {
		return []*model.Device{}, nil
	}

	for _, d := range devices {
		if d.ScheduleIDs, err = dr.LoadScheduleIDs(d.ID); err != nil {
			return nil, err
		}
	}

	return devices, nil
}

// FindBySchedule returns all existing devices with given schedule id from the database
func (dr *deviceRepo) FindBySchedule(scheduleID int) ([]*model.Device, error) {
	query := "SELECT DISTINCT d.id, d.hash, d.name, d.os, d.user_id FROM device d JOIN device_schedule ds ON d.id = ds.device_id WHERE ds.schedule_id = ?"
	var devices []*model.Device

	err := sqlscan.Select(dr.ctx, dr.db, &devices, query, scheduleID)
	if err != nil {
		return []*model.Device{}, nil
	}

	if devices == nil {
		return []*model.Device{}, nil
	}

	for _, d := range devices {
		if d.ScheduleIDs, err = dr.LoadScheduleIDs(d.ID); err != nil {
			return nil, err
		}
	}

	return devices, nil
}

// FindAll returns all existing devices from the database
func (dr *deviceRepo) FindAll() ([]*model.Device, error) {
	query := "SELECT id, hash, name, os, user_id FROM device"
	var devices []*model.Device

	err := sqlscan.Select(dr.ctx, dr.db, &devices, query)
	if err != nil || devices == nil {
		return []*model.Device{}, nil
	}

	for _, d := range devices {
		if d.ScheduleIDs, err = dr.LoadScheduleIDs(d.ID); err != nil {
			return nil, err
		}
	}

	return devices, nil
}

// FindDeviceSchedule returns all existing entries from device schedule many-to-many table
func (dr *deviceRepo) FindDeviceSchedule() ([]*model.DeviceSchedule, error) {
	query := "SELECT device_id, schedule_id from device_schedule"
	var deviceSchedule []*model.DeviceSchedule

	err := sqlscan.Select(dr.ctx, dr.db, &deviceSchedule, query)
	if err != nil {
		return []*model.DeviceSchedule{}, nil
	}

	return deviceSchedule, nil
}
