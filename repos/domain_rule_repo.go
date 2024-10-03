package repos

import (
	"database/sql"
	"server/model"

	_ "github.com/mattn/go-sqlite3"
)

// Define DomainRuleRepo interface
type DomainRuleRepo interface {
	Create(dr *model.DomainRule) error
	Update(dr *model.DomainRule) error
	Delete(dr *model.DomainRule) error
}

// Define DomainRuleRepoImpl struct
type DomainRuleRepoImpl struct {
	DB *sql.DB
}

// Create new DomainRuleRepo
func NewDomainRuleRepo(db *sql.DB) *DomainRuleRepoImpl {
	return &DomainRuleRepoImpl{DB: db}
}

// Store a new domain rule into the db
func (repo *DomainRuleRepoImpl) Create(dr *model.DomainRule) error {
	sql := "INSERT INTO DomainRule VALUES (?, ?, ?, ?)"
	_, err := repo.DB.Exec(sql, dr.Id, dr.Domain.Id, dr.ScheduleId, dr.Blocked)
	return err
}

// Update an existng domain rule in the db
func (repo *DomainRuleRepoImpl) Update(dr *model.DomainRule) error {
	sql := "UPDATE DomainRule SET domain_id=?, schedule_id=?, blocked=? WHERE id=?"
	_, err := repo.DB.Exec(sql, dr.Domain.Id, dr.ScheduleId, dr.Blocked, dr.Id)
	return err
}

// Delete an existing domain rule from the db
func (repo *DomainRuleRepoImpl) Delete(dr *model.DomainRule) error {
	sql := "DELETE FROM DomainRule WHERE id=?"
	_, err := repo.DB.Exec(sql, dr.Id)
	return err
}
