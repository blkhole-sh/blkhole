package repos

import (
	"database/sql"
	"server/model"

	_ "github.com/mattn/go-sqlite3"
)

// Define CategoryRepo interface
type CategoryRepo interface {
	Create(c *model.Category) error
	Update(c *model.Category) error
	Delete(c *model.Category) error
}

// Define CategoryRepoImpl
type CategoryRepoImpl struct {
	DB *sql.DB
}

// Create new CategoryRepoImpl
func NewCategoryRepo(db *sql.DB) *CategoryRepoImpl {
	return &CategoryRepoImpl{DB: db}
}

// Store a new category into the db
func (repo *CategoryRepoImpl) Create(c *model.Category) error {
	sql := "INSERT INTO Category VALUES (?)"
	_, err := repo.DB.Exec(sql, c.Name)
	return err
}

// Update an existing category in the db
func (repo *CategoryRepoImpl) Update(c *model.Category) error {
	sql := "UPDATE Category SET name=? WHERE id=?"
	_, err := repo.DB.Exec(sql, c.Name, c.Id)
	return err
}

// Delete an existing category from the db
func (repo *CategoryRepoImpl) Delete(c *model.Category) error {
	sql := "DELETE FROM Category WHERE id=?"
	_, err := repo.DB.Exec(sql, c.Id)
	return err
}
