package repos

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/blkhole-sh/blkhole/internal/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// RuleRepo defines the interface for rule repository operations
type RuleRepo interface {
	CreateOrGet(r *model.Rule) (int, error)
	BatchCreateOrGet(rules []*model.Rule) error
	Update(id int, r *model.Rule) error
	Delete(id int) error
	FindByID(id int) (*model.Rule, error)
	FindByList(listID int) ([]*model.Rule, error)
	FindByDomain(domain string) ([]*model.Rule, error)
	FindAll() ([]*model.Rule, error)
	LinkToList(ruleID int, listID int) error
	BatchLinkToList(ruleIDs []int, listID int) error
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
	query := "SELECT id FROM rule WHERE domain_id = ? AND allowed = ?"
	err := rr.db.QueryRowContext(rr.ctx, query, r.DomainID, r.Allowed).Scan(&r.ID)
	if err == nil {
		// Rule already exists, return its ID
		return r.ID, nil
	}

	// Rule doesn't exist, create it
	if err == sql.ErrNoRows {
		query := "INSERT INTO rule (domain_id, allowed) VALUES (?, ?) RETURNING id"
		err = rr.db.QueryRowContext(rr.ctx, query, r.DomainID, r.Allowed).Scan(&r.ID)
		return r.ID, err
	}

	// Other error occurred
	return 0, err
}

// BatchCreateOrGet stores new rules into the database or returns existing ones
func (rr *ruleRepo) BatchCreateOrGet(rules []*model.Rule) error {
	if len(rules) == 0 {
		return nil
	}

	// Use a transaction for the entire batch operation
	tx, err := rr.db.BeginTx(rr.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create temporary table for batch data
	_, err = tx.ExecContext(rr.ctx, "CREATE TEMPORARY TABLE IF NOT EXISTS batch_rules (domain_id INTEGER, allowed INTEGER)")
	if err != nil {
		return fmt.Errorf("failed to create temporary table: %w", err)
	}
	// Ensure we start empty
	_, err = tx.ExecContext(rr.ctx, "DELETE FROM batch_rules")
	if err != nil {
		return fmt.Errorf("failed to clear temporary table: %w", err)
	}

	// Prepare statement for inserting into temp table
	stmt, err := tx.PrepareContext(rr.ctx, "INSERT INTO batch_rules (domain_id, allowed) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, r := range rules {
		_, err = stmt.ExecContext(rr.ctx, r.DomainID, r.Allowed)
		if err != nil {
			return fmt.Errorf("failed to insert into temp table: %w", err)
		}
	}

	// Insert new rules from temp table into main table, ignoring duplicates
	_, err = tx.ExecContext(rr.ctx, "INSERT OR IGNORE INTO rule (domain_id, allowed) SELECT domain_id, allowed FROM batch_rules")
	if err != nil {
		return fmt.Errorf("failed to insert rules: %w", err)
	}

	// Select IDs back
	// We join with the temp table to get only the requested rules
	rows, err := tx.QueryContext(rr.ctx, `
		SELECT r.id, r.domain_id, r.allowed
		FROM rule r
		JOIN batch_rules br ON r.domain_id = br.domain_id AND r.allowed = br.allowed
	`)
	if err != nil {
		return fmt.Errorf("failed to select rule IDs: %w", err)
	}
	defer rows.Close()

	// Map results back to input rules
	type ruleKey struct {
		domainID int
		allowed  bool
	}
	idMap := make(map[ruleKey]int)

	for rows.Next() {
		var id, domainID int
		var allowed bool
		if err := rows.Scan(&id, &domainID, &allowed); err != nil {
			return fmt.Errorf("failed to scan rule: %w", err)
		}
		idMap[ruleKey{domainID, allowed}] = id
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}

	// Update input rules with IDs
	for _, r := range rules {
		if id, ok := idMap[ruleKey{r.DomainID, r.Allowed}]; ok {
			r.ID = id
		} else {
			// This should not happen if everything went well
			return fmt.Errorf("failed to retrieve ID for rule domain=%d allowed=%v", r.DomainID, r.Allowed)
		}
	}

	// Drop temp table to be clean
	_, err = tx.ExecContext(rr.ctx, "DROP TABLE batch_rules")
	if err != nil {
		return fmt.Errorf("failed to drop temp table: %w", err)
	}

	return tx.Commit()
}

// Update modifies an existing rule with given ID in the database if not referenced
func (rr *ruleRepo) Update(id int, r *model.Rule) error {
	query := `UPDATE rule SET domain_id=?, allowed=? WHERE id = ? AND id NOT IN (
		SELECT DISTINCT rule_id FROM list_rule
		UNION
		SELECT DISTINCT rule_id FROM schedule_rule
	)`

	result, err := rr.db.ExecContext(rr.ctx, query, r.DomainID, r.Allowed, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("rule %d not updated: missing or referenced by a list or schedule", id)
	}

	return nil
}

// Delete removes an existing rule with given ID from the database if not referenced
func (rr *ruleRepo) Delete(id int) error {
	query := `DELETE FROM rule WHERE id = ? AND id NOT IN (
		SELECT DISTINCT rule_id FROM list_rule
		UNION
		SELECT DISTINCT rule_id FROM schedule_rule
	)`

	result, err := rr.db.ExecContext(rr.ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("rule %d not deleted: missing or referenced by a list or schedule", id)
	}

	return nil
}

// FindAll returns all existing rules from the database
func (rr *ruleRepo) FindAll() ([]*model.Rule, error) {
	query := "SELECT id, domain_id, allowed FROM rule"
	var rules []*model.Rule

	err := sqlscan.Select(rr.ctx, rr.db, &rules, query)
	if err != nil {
		return nil, err
	}

	if rules == nil {
		return []*model.Rule{}, nil
	}

	return rules, nil
}

// FindByID returns an existing rule with given id from the database
func (rr *ruleRepo) FindByID(id int) (*model.Rule, error) {
	query := "SELECT id, domain_id, allowed FROM rule WHERE id=?"
	var r model.Rule

	if err := sqlscan.Get(rr.ctx, rr.db, &r, query, id); err != nil {
		return nil, err
	}

	return &r, nil
}

// FindByList returns all existing rules for a given list ID
func (rr *ruleRepo) FindByList(listID int) ([]*model.Rule, error) {
	query := "SELECT r.id, r.domain_id, r.allowed FROM rule r JOIN list_rule lr ON r.id = lr.rule_id WHERE lr.list_id=?"
	var rules []*model.Rule

	err := sqlscan.Select(rr.ctx, rr.db, &rules, query, listID)
	if err != nil {
		return nil, err
	}

	if rules == nil {
		return []*model.Rule{}, nil
	}

	return rules, nil
}

// FindByDomain returns all existing rules for a given domain
func (rr *ruleRepo) FindByDomain(domain string) ([]*model.Rule, error) {
	query := "SELECT r.id, r.domain_id, r.allowed FROM rule r JOIN domain d ON r.domain_id = d.id WHERE d.name=?"
	var rules []*model.Rule

	err := sqlscan.Select(rr.ctx, rr.db, &rules, query, domain)
	if err != nil {
		return nil, err
	}

	if rules == nil {
		return []*model.Rule{}, nil
	}

	return rules, nil
}

// LinkToList creates a link between a rule and a list
func (rr *ruleRepo) LinkToList(ruleID int, listID int) error {
	query := "INSERT OR IGNORE INTO list_rule (list_id, rule_id) VALUES (?, ?)"
	result, err := rr.db.ExecContext(rr.ctx, query, listID, ruleID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows > 0 {
		updateQuery := "UPDATE list SET count = count + 1 WHERE id = ?"
		_, err := rr.db.ExecContext(rr.ctx, updateQuery, listID)
		return err
	}

	return nil
}

// BatchLinkToList links multiple rules to a list
func (rr *ruleRepo) BatchLinkToList(ruleIDs []int, listID int) error {
	if len(ruleIDs) == 0 {
		return nil
	}

	// SQLite limit for variables is usually 999 or higher.
	// We'll process in chunks of 400 rules (800 params) to be safe.
	chunkSize := 400

	for i := 0; i < len(ruleIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(ruleIDs) {
			end = len(ruleIDs)
		}

		chunk := ruleIDs[i:end]
		if len(chunk) == 0 {
			continue
		}

		valueStrings := make([]string, 0, len(chunk))
		valueArgs := make([]interface{}, 0, len(chunk)*2)

		for _, ruleID := range chunk {
			valueStrings = append(valueStrings, "(?, ?)")
			valueArgs = append(valueArgs, listID, ruleID)
		}

		stmt := fmt.Sprintf("INSERT OR IGNORE INTO list_rule (list_id, rule_id) VALUES %s",
			strings.Join(valueStrings, ","))

		_, err := rr.db.ExecContext(rr.ctx, stmt, valueArgs...)
		if err != nil {
			return fmt.Errorf("failed to batch link rules: %w", err)
		}
	}

	return nil
}

// UnlinkFromList removes a link between a rule and a list
func (rr *ruleRepo) UnlinkFromList(ruleID int, listID int) error {
	query := "DELETE FROM list_rule WHERE list_id=? AND rule_id=?"
	result, err := rr.db.ExecContext(rr.ctx, query, listID, ruleID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows > 0 {
		updateQuery := "UPDATE list SET count = count - 1 WHERE id = ?"
		_, err := rr.db.ExecContext(rr.ctx, updateQuery, listID)
		return err
	}

	return nil
}
