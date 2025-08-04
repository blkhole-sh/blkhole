package repos

import (
	"context"
	"database/sql"
	"server/internal/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// RuleRepo defines the interface for rule rrsitory operations
type RuleRepo interface {
	Create(r *model.Rule) (int, error)
	Update(id int, r *model.Rule) error
	Delete(id int) error
	FindByID(id int) (*model.Rule, error)
	FindByList(listID int) ([]*model.Rule, error)
	FindByDomain(domain string) ([]*model.Rule, error)
}

// ruleRepo implements the RuleRepo interface
type ruleRepo struct {
	db  *sql.DB
	ctx context.Context
}

// NewRuleRepo creates a new RuleRepo instance
func NewRuleRepo(db *sql.DB) RuleRepo {
	return &ruleRepo{db: db, ctx: context.Background()}
}

// Create stores a new rule into the database
func (rr *ruleRepo) Create(r *model.Rule) (int, error) {
	sql := "INSERT INTO rule (domain, list_id, allowed) VALUES (?, ?, ?) RETURNING id"
	err := rr.db.QueryRowContext(rr.ctx, sql, r.Domain, r.ListID, r.Allowed).Scan(&r.ID)
	return r.ID, err
}

// Update modifies an existing rule with given ID in the database
func (rr *ruleRepo) Update(id int, r *model.Rule) error {
	sql := "UPDATE rule SET domain=?, list_id=?, allowed=? WHERE id=?"
	_, err := rr.db.ExecContext(rr.ctx, sql, r.Domain, r.ListID, r.Allowed, id)
	return err
}

// Delete removes an existing rule with given ID from the database
func (rr *ruleRepo) Delete(id int) error {
	sql := "DELETE FROM rule WHERE id=?"
	_, err := rr.db.ExecContext(rr.ctx, sql, id)
	return err
}

// FindByID returns an existing rule with given id from the database
func (rr *ruleRepo) FindByID(id int) (*model.Rule, error) {
	sql := "SELECT id, domain, list_id, allowed FROM rule WHERE id=?"
	var r model.Rule

	if err := sqlscan.Get(rr.ctx, rr.db, &r, sql, id); err != nil {
		return nil, err
	}

	return &r, nil
}

// FindByList returns all existing rules for a given list ID
func (rr *ruleRepo) FindByList(listID int) ([]*model.Rule, error) {
	sql := "SELECT id, domain, list_id, allowed FROM rule WHERE list_id=?"
	var rules []*model.Rule

	if err := sqlscan.Select(rr.ctx, rr.db, &rules, sql, listID); err != nil {
		return []*model.Rule{}, nil
	}

	// Ensure we return empty slice instead of nil
	if rules == nil {
		return []*model.Rule{}, nil
	}

	return rules, nil
}

// FindByDomain returns all existing rules for a given domain
func (rr *ruleRepo) FindByDomain(domain string) ([]*model.Rule, error) {
	sql := "SELECT id, domain, list_id, allowed FROM rule WHERE domain=?"
	var rules []*model.Rule

	if err := sqlscan.Select(rr.ctx, rr.db, &rules, sql, domain); err != nil {
		return []*model.Rule{}, nil
	}

	// Ensure we return empty slice instead of nil
	if rules == nil {
		return []*model.Rule{}, nil
	}

	return rules, nil
}
