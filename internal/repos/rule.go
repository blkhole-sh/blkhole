package repos

import (
	"context"
	"database/sql"

	"github.com/lemon3studio/leo/internal/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// RuleRepo defines the interface for rule repository operations
type RuleRepo interface {
	CreateOrGet(r *model.Rule) (int, error)
	Update(id int, r *model.Rule) error
	Delete(id int) error
	FindByID(id int) (*model.Rule, error)
	FindByList(listID int) ([]*model.Rule, error)
	FindByDomain(domain string) ([]*model.Rule, error)
	LinkToList(ruleID int, listID int) error
	UnlinkFromList(ruleID int, listID int) error
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

// CreateOrGet stores a new rule into the database or returns existing one
func (rr *ruleRepo) CreateOrGet(r *model.Rule) (int, error) {
	// First, try to find existing rule
	query := "SELECT id FROM rule WHERE domain = ? AND allowed = ?"
	err := rr.db.QueryRowContext(rr.ctx, query, r.Domain, r.Allowed).Scan(&r.ID)
	if err == nil {
		// Rule already exists, return its ID
		return r.ID, nil
	}

	// Rule doesn't exist, create it
	if err == sql.ErrNoRows {
		query := "INSERT INTO rule (domain, allowed) VALUES (?, ?) RETURNING id"
		err = rr.db.QueryRowContext(rr.ctx, query, r.Domain, r.Allowed).Scan(&r.ID)
		return r.ID, err
	}

	// Other error occurred
	return 0, err
}

// Update modifies an existing rule with given ID in the database if not referenced
func (rr *ruleRepo) Update(id int, r *model.Rule) error {
	query := `UPDATE rule SET domain=?, allowed=? WHERE id = ? AND id NOT IN (
		SELECT DISTINCT rule_id FROM list_rule 
		UNION 
		SELECT DISTINCT rule_id FROM schedule_rule
	)`

	_, err := rr.db.ExecContext(rr.ctx, query, r.Domain, r.Allowed, id)
	if err != nil {
		return err
	}

	return err
}

// Delete removes an existing rule with given ID from the database if not referenced
func (rr *ruleRepo) Delete(id int) error {
	query := `DELETE FROM rule WHERE id = ? AND id NOT IN (
		SELECT DISTINCT rule_id FROM list_rule 
		UNION 
		SELECT DISTINCT rule_id FROM schedule_rule
	)`

	_, err := rr.db.ExecContext(rr.ctx, query, id)
	if err != nil {
		return err
	}

	return err
}

// FindByID returns an existing rule with given id from the database
func (rr *ruleRepo) FindByID(id int) (*model.Rule, error) {
	query := "SELECT id, domain, allowed FROM rule WHERE id=?"
	var r model.Rule

	if err := sqlscan.Get(rr.ctx, rr.db, &r, query, id); err != nil {
		return nil, err
	}

	return &r, nil
}

// FindByList returns all existing rules for a given list ID
func (rr *ruleRepo) FindByList(listID int) ([]*model.Rule, error) {
	query := "SELECT r.id, r.domain, r.allowed FROM rule r JOIN list_rule lr ON r.id = lr.rule_id WHERE lr.list_id=?"
	var rules []*model.Rule

	if err := sqlscan.Select(rr.ctx, rr.db, &rules, query, listID); err != nil {
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
	query := "SELECT id, domain, allowed FROM rule WHERE domain=?"
	var rules []*model.Rule

	if err := sqlscan.Select(rr.ctx, rr.db, &rules, query, domain); err != nil {
		return []*model.Rule{}, nil
	}

	// Ensure we return empty slice instead of nil
	if rules == nil {
		return []*model.Rule{}, nil
	}

	return rules, nil
}

// LinkToList creates a link between a rule and a list
func (rr *ruleRepo) LinkToList(ruleID int, listID int) error {
	query := "INSERT INTO list_rule (list_id, rule_id) VALUES (?, ?)"
	_, err := rr.db.ExecContext(rr.ctx, query, listID, ruleID)
	return err
}

// UnlinkFromList removes a link between a rule and a list
func (rr *ruleRepo) UnlinkFromList(ruleID int, listID int) error {
	query := "DELETE FROM list_rule WHERE list_id=? AND rule_id=?"
	_, err := rr.db.ExecContext(rr.ctx, query, listID, ruleID)
	return err
}
