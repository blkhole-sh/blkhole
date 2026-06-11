// Package repos provides database repository implementations for the blkhole DNS blocker.
package repos

import (
	"context"
	"database/sql"

	"github.com/blkhole-sh/blkhole/internal/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// UserRepo defines the interface for user repository operations
type UserRepo interface {
	Create(u *model.User) error
	Update(id int, u *model.User) error
	Delete(id int) error
	LoadDeviceIDs(id int) ([]int, error)
	LoadListIDs(id int) ([]int, error)
	LoadScheduleIDs(id int) ([]int, error)
	LoadRelations(u *model.User) error
	FindByEmail(email string) (*model.User, error)
	FindByID(id int) (*model.User, error)
	FindAllIDs() ([]int, error)
}

// userRepo implements the Userur interface
type userRepo struct {
	db  *sql.DB
	ctx context.Context
}

// NewUserRepo creates a new Userur instance
func NewUserRepo(db *sql.DB) UserRepo {
	return &userRepo{db: db, ctx: context.Background()}
}

// Create stores a new user into the database
func (ur *userRepo) Create(u *model.User) error {
	query := "INSERT INTO user (name, email, password_hash) VALUES (?, ?, ?)"

	result, err := ur.db.ExecContext(ur.ctx, query, u.Name, u.Email, u.PasswordHash)
	if err != nil {
		return err
	}

	// Get the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = int(id)

	return nil
}

// Update modifies an existing user with given id in the database
func (ur *userRepo) Update(id int, u *model.User) error {
	query := "UPDATE user SET name=?, email=?, password_hash=? WHERE id=?"
	_, err := ur.db.ExecContext(ur.ctx, query, u.Name, u.Email, u.PasswordHash, id)
	return err
}

// Delete removes an existing user with given id from the database
func (ur *userRepo) Delete(id int) error {
	query := "DELETE FROM user WHERE id=?"
	_, err := ur.db.ExecContext(ur.ctx, query, id)
	return err
}

// LoadDeviceIDs returns ids of all devices linked to user with given id
func (ur *userRepo) LoadDeviceIDs(id int) ([]int, error) {
	query := "SELECT id FROM device WHERE user_id = ?"
	var deviceIDs []int

	if err := sqlscan.Select(ur.ctx, ur.db, &deviceIDs, query, id); err != nil {
		return nil, err
	}

	// Ensure we return empty slice instead of nil
	if deviceIDs == nil {
		return []int{}, nil
	}

	return deviceIDs, nil
}

// LoadListIDs returns ids of all lists linked to user with given id
func (ur *userRepo) LoadListIDs(id int) ([]int, error) {
	query := "SELECT id FROM list WHERE user_id = ?"
	var listIDs []int

	if err := sqlscan.Select(ur.ctx, ur.db, &listIDs, query, id); err != nil {
		return nil, err
	}

	// Ensure we return empty slice instead of nil
	if listIDs == nil {
		return []int{}, nil
	}

	return listIDs, nil
}

// LoadScheduleIDs returns ids of all schedules linked to user with given id
func (ur *userRepo) LoadScheduleIDs(id int) ([]int, error) {
	query := "SELECT id FROM schedule WHERE user_id = ?"
	var scheduleIDs []int

	if err := sqlscan.Select(ur.ctx, ur.db, &scheduleIDs, query, id); err != nil {
		return nil, err
	}

	// Ensure we return empty slice instead of nil
	if scheduleIDs == nil {
		return []int{}, nil
	}

	return scheduleIDs, nil
}

// LoadRelations loads all relations (ids) for user with given id
func (ur *userRepo) LoadRelations(u *model.User) error {
	var err error

	if u.DeviceIDs, err = ur.LoadDeviceIDs(u.ID); err != nil {
		return err
	}

	if u.ListIDs, err = ur.LoadListIDs(u.ID); err != nil {
		return err
	}

	if u.ScheduleIDs, err = ur.LoadScheduleIDs(u.ID); err != nil {
		return err
	}

	return nil
}

// FindByEmail returns an existing user with given email from the database
func (ur *userRepo) FindByEmail(email string) (*model.User, error) {
	query := "SELECT id, name, email, password_hash FROM user WHERE email=?"
	var u model.User

	if err := sqlscan.Get(ur.ctx, ur.db, &u, query, email); err != nil {
		return nil, err
	}

	if err := ur.LoadRelations(&u); err != nil {
		return nil, err
	}

	return &u, nil
}

// FindAllIDs returns the IDs of all users in the database.
func (ur *userRepo) FindAllIDs() ([]int, error) {
	var ids []int
	if err := sqlscan.Select(ur.ctx, ur.db, &ids, "SELECT id FROM user"); err != nil {
		return nil, err
	}
	if ids == nil {
		return []int{}, nil
	}
	return ids, nil
}

// FindByID returns an existing user with given id from the database
func (ur *userRepo) FindByID(id int) (*model.User, error) {
	query := "SELECT id, name, email, password_hash FROM user WHERE id=?"
	var u model.User

	if err := sqlscan.Get(ur.ctx, ur.db, &u, query, id); err != nil {
		return nil, err
	}

	if err := ur.LoadRelations(&u); err != nil {
		return nil, err
	}

	return &u, nil
}
