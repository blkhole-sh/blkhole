package repo

import (
	"database/sql"
	"server/model"

	_ "github.com/mattn/go-sqlite3"
)

type UserRepo interface {
	Create(u *model.User) error
	Update(u *model.User) error
	Delete(u *model.User) error
	FindByHash(hash string) (*model.User, error)
	FindByLogin(email string, passwordHash string) (*model.User, error)
}

// Define UserRepoImpl struct
type UserRepoImpl struct {
	DB *sql.DB
}

// Create new UserRepoImpl
func NewUserRepo(db *sql.DB) *UserRepoImpl {
	return &UserRepoImpl{DB: db}
}

// Store a new user into the db
func (repo *UserRepoImpl) Create(u *model.User) error {
	sql := "INSERT INTO TABLE User VALUES (?, ?, ?, ?)"
	_, err := repo.DB.Exec(sql, u.Hash, u.Name, u.Email, u.PasswordHash)
	return err
}

// Update an existing user in the db
func (repo *UserRepoImpl) Update(u *model.User) error {
	sql := "UPDATE User SET name=?, email=?, password_hash=? WHERE hash=?"
	_, err := repo.DB.Exec(sql, u.Name, u.Email, u.PasswordHash, u.Hash)
	return err
}

// Delete an existing user from the db
func (repo *UserRepoImpl) Delete(u *model.User) error {
	sql := "DELETE FROM User WHERE hash=?"
	_, err := repo.DB.Exec(sql, u.Hash)
	return err
}

// Find an existing user with given hash in the db
func (repo *UserRepoImpl) FindByHash(hash string) (*model.User, error) {
	sql := "SELECT * FROM User WHERE hash=?"
	user := model.User{}
	err := repo.DB.QueryRow(sql, hash).Scan(&user.Hash, &user.Name, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Find an existing user with given email and password hash in the db
func (repo *UserRepoImpl) FindByLogin(email string, passwordHash string) (*model.User, error) {
	sql := "SELECT * FROM User WHERE email=? AND password_hash=?"
	user := model.User{}
	err := repo.DB.QueryRow(sql, email, passwordHash).Scan(&user.Hash, &user.Name, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
