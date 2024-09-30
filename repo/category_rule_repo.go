package repo

import (
	"database/sql"
	"server/model"

	_ "github.com/mattn/go-sqlite3"
)

// Define CategoryRuleRepo interface
type CategoryRuleRepo interface {
	Create(cr *model.CategoryRule) error
	Update(cr *model.CategoryRule) error
	Delete(cr *model.CategoryRule) error
}

// Define CategoryRuleRepoImpl struct
type CategoryRuleRepoImpl struct {
	DB *sql.DB
}

// Create new CategoryRuleRepo
func NewCategoryRuleRepo(db *sql.DB) *CategoryRuleRepoImpl {
	return &CategoryRuleRepoImpl{DB: db}
}

// Store a new category rule into the db
func (repo *CategoryRuleRepoImpl) Create(cr *model.CategoryRule) error {
	sql := "INSERT INTO CategoryRule VALUES (?, ?, ?, ?)"
	_, err := repo.DB.Exec(sql, cr.Id, cr.Category.Id, cr.ScheduleId, cr.Blocked)
	return err
}

// Update an existing category rule in the db
func (repo *CategoryRuleRepoImpl) Update(cr *model.CategoryRule) error {
	sql := "UPDATE CategoryRule SET category_id=?, schedule_id=?, blocked=? WHERE id=?"
	_, err := repo.DB.Exec(sql, cr.Category.Id, cr.ScheduleId, cr.Blocked, cr.Id)
	return err
}

// Delete an existing category rule from the db
func (repo *CategoryRuleRepoImpl) Delete(cr *model.CategoryRule) error {
	sql := "DELETE FROM CategoryRule WHERE id=?"
	_, err := repo.DB.Exec(sql, cr.Id)
	return err
}
