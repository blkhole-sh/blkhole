package repos

import (
	"context"
	"database/sql"
	"server/internal/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// RuleRepo defines the interface for rule repository operations
type RuleRepo interface {
	Create(r *model.Rule) (int, error)
	Update(id int, r *model.Rule) error
	Delete(id int) error
	FindByID(id int) (*model.Rule, error)
	FindByList(listID int) ([]*model.Rule, error)
	FindByDomain(domain string) ([]*model.Rule, error)
}

// RuleRepoImpl implements the RuleRepo interface
type RuleRepoImpl struct {
	db  *sql.DB
	ctx context.Context
}

// NewRuleRepo creates a new RuleRepo instance
func NewRuleRepo(db *sql.DB) RuleRepo {
	return &RuleRepoImpl{db: db, ctx: context.Background()}
}

// Create stores a new rule into the database
func (repo *RuleRepoImpl) Create(r *model.Rule) (int, error) {
	sql := "INSERT INTO rule (domain, list_id, allowed) VALUES (?, ?, ?) RETURNING id"
	err := repo.db.QueryRowContext(repo.ctx, sql, r.Domain, r.ListID, r.Allowed).Scan(&r.ID)
	return r.ID, err
}

// Update modifies an existing rule with given ID in the database
func (repo *RuleRepoImpl) Update(id int, r *model.Rule) error {
	sql := "UPDATE rule SET domain=?, list_id=?, allowed=? WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, r.Domain, r.ListID, r.Allowed, id)
	return err
}

// Delete removes an existing rule with given ID from the database
func (repo *RuleRepoImpl) Delete(id int) error {
	sql := "DELETE FROM rule WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, id)
	return err
}

// FindByID returns an existing rule with given id from the database
func (repo *RuleRepoImpl) FindByID(id int) (*model.Rule, error) {
	sql := "SELECT id, domain, list_id, allowed FROM rule WHERE id=?"
	var r model.Rule

	if err := sqlscan.Get(repo.ctx, repo.db, &r, sql, id); err != nil {
		return nil, err
	}

	return &r, nil
}

// FindByList returns all existing rules for a given list ID
func (repo *RuleRepoImpl) FindByList(listID int) ([]*model.Rule, error) {
	sql := "SELECT id, domain, list_id, allowed FROM rule WHERE list_id=?"
	var rules []*model.Rule

	if err := sqlscan.Select(repo.ctx, repo.db, &rules, sql, listID); err != nil {
		return []*model.Rule{}, nil
	}

	// Ensure we return empty slice instead of nil
	if rules == nil {
		return []*model.Rule{}, nil
	}

	return rules, nil
}

// FindByDomain returns all existing rules for a given domain
func (repo *RuleRepoImpl) FindByDomain(domain string) ([]*model.Rule, error) {
	sql := "SELECT id, domain, list_id, allowed FROM rule WHERE domain=?"
	var rules []*model.Rule

	if err := sqlscan.Select(repo.ctx, repo.db, &rules, sql, domain); err != nil {
		return []*model.Rule{}, nil
	}

	// Ensure we return empty slice instead of nil
	if rules == nil {
		return []*model.Rule{}, nil
	}

	return rules, nil
}
