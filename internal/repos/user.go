// Package repos provides database repository implementations for the Leo DNS blocker.
package repos

import (
	"context"
	"database/sql"
	"github.com/lemon3studio/leo/internal/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

type UserRepo interface {
	Create(u *model.User) error
	Update(hash string, u *model.User) error
	Delete(hash string) error
	LoadDeviceHashes(hash string) ([]string, error)
	LoadListIDs(hash string) ([]int, error)
	LoadScheduleIDs(hash string) ([]int, error)
	LoadRelations(u *model.User) error
	FindByEmail(email string) (*model.User, error)
	FindByHash(hash string) (*model.User, error)
}

// userRepo implements the UserRepo interface
type userRepo struct {
	db  *sql.DB
	ctx context.Context
}

// NewUserRepo creates a new UserRepo instance
func NewUserRepo(db *sql.DB) UserRepo {
	return &userRepo{db: db, ctx: context.Background()}
}

// Create stores a new user into the database
func (repo *userRepo) Create(u *model.User) error {
	sql := "INSERT INTO user (hash, name, email, password_hash) VALUES (?, ?, ?, ?)"

	_, err := repo.db.ExecContext(repo.ctx, sql, u.Hash, u.Name, u.Email, u.PasswordHash)
	return err
}

// Update modifies an existing user with given hash in the database
func (repo *userRepo) Update(hash string, u *model.User) error {
	sql := "UPDATE user SET name=?, email=?, password_hash=? WHERE hash=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, u.Name, u.Email, u.PasswordHash, hash)
	return err
}

// Delete removes an existing user with given hash from the database
func (repo *userRepo) Delete(hash string) error {
	sql := "DELETE FROM user WHERE hash=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, hash)
	return err
}

// LoadDeviceHashes returns hashes of all devices linked to user with given hash
func (repo *userRepo) LoadDeviceHashes(hash string) ([]string, error) {
	sql := "SELECT hash FROM device WHERE user_hash = ?"
	var deviceHashes []string

	if err := sqlscan.Select(repo.ctx, repo.db, &deviceHashes, sql, hash); err != nil {
		return []string{}, nil
	}

	// Ensure we return empty slice instead of nil
	if deviceHashes == nil {
		return []string{}, nil
	}

	return deviceHashes, nil
}

// LoadListIDs returns ids of all lists linked to user with given hash
func (repo *userRepo) LoadListIDs(hash string) ([]int, error) {
	sql := "SELECT id FROM list WHERE user_hash = ?"
	var listIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &listIds, sql, hash); err != nil {
		return []int{}, nil
	}

	// Ensure we return empty slice instead of nil
	if listIds == nil {
		return []int{}, nil
	}

	return listIds, nil
}

// LoadScheduleIDs returns ids of all schedules linked to user with given hash
func (repo *userRepo) LoadScheduleIDs(hash string) ([]int, error) {
	sql := "SELECT id FROM schedule WHERE user_hash = ?"
	var scheduleIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &scheduleIds, sql, hash); err != nil {
		return []int{}, nil
	}

	// Ensure we return empty slice instead of nil
	if scheduleIds == nil {
		return []int{}, nil
	}

	return scheduleIds, nil
}

// LoadRelations loads all relations (hashes, ids) for user with given hash
func (repo *userRepo) LoadRelations(u *model.User) error {
	var err error

	if u.DeviceHashes, err = repo.LoadDeviceHashes(u.Hash); err != nil {
		return err
	}

	if u.ListIds, err = repo.LoadListIDs(u.Hash); err != nil {
		return err
	}

	if u.ScheduleIds, err = repo.LoadScheduleIDs(u.Hash); err != nil {
		return err
	}

	return nil
}

// FindByEmail returns an existing user with given email from the database
func (repo *userRepo) FindByEmail(email string) (*model.User, error) {
	sql := "SELECT hash, name, email, password_hash FROM user WHERE email=?"
	var u model.User

	if err := sqlscan.Get(repo.ctx, repo.db, &u, sql, email); err != nil {
		return nil, err
	}

	if err := repo.LoadRelations(&u); err != nil {
		return nil, err
	}

	return &u, nil
}

// FindByHash returns an existing user with given hash from the database
func (repo *userRepo) FindByHash(hash string) (*model.User, error) {
	sql := "SELECT hash, name, email, password_hash FROM user WHERE hash=?"
	var u model.User

	if err := sqlscan.Get(repo.ctx, repo.db, &u, sql, hash); err != nil {
		return nil, err
	}

	if err := repo.LoadRelations(&u); err != nil {
		return nil, err
	}

	return &u, nil
}
