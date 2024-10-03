package repos

import (
	"database/sql"
	"server/model"

	_ "github.com/mattn/go-sqlite3"
)

type DeviceRepo interface {
	Create(d *model.Device) error
	Update(d *model.Device) error
	Delete(d *model.Device) error
	FindByUserHash(userHash string) (*model.Device, error)
}

// Define DeviceRepoImpl struct
type DeviceRepoImpl struct {
	DB *sql.DB
}

// Create new DeviceRepoImpl
func NewDeviceRepo(db *sql.DB) *DeviceRepoImpl {
	return &DeviceRepoImpl{DB: db}
}

// Store a new device into the db
func (repo *DeviceRepoImpl) Create(d *model.Device) error {
	sql := "INSERT INTO Device VALUES (?, ?, ?)"
	_, err := repo.DB.Exec(sql, d.Hash, d.Name, d.UserHash)
	return err
}

// Update an existing device in the db
func (repo *DeviceRepoImpl) Update(d *model.Device) error {
	sql := "UPDATE Device SET name=?, user_hash=? WHERE hash=?"
	_, err := repo.DB.Exec(sql, d.Hash, d.UserHash, d.Hash)
	return err
}

// Delete an existing device from the db
func (repo *DeviceRepoImpl) Delete(d *model.Device) error {
	sql := "DELETE FROM Device WHERE hash=?"
	_, err := repo.DB.Exec(sql, d.Hash)
	return err
}

// Find all existing devices with given user hash in the db
func (repo *DeviceRepoImpl) FindByUserHash(userHash string) ([]*model.Device, error) {
	sql := "SELECT * FROM Device WHERE user_hash=?"

	rows, err := repo.DB.Query(sql, userHash)
	if err != nil {
		return nil, err
	}

	devices := []*model.Device{}

	for rows.Next() {
		d := model.Device{}
		err := rows.Scan(&d.Hash, &d.Name, &d.UserHash)
		if err != nil {
			return nil, err
		}

		devices = append(devices, &d)
	}

	return devices, nil
}
