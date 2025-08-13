package repos

import (
	"context"
	"database/sql"

	"github.com/lemon3studio/leo/internal/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// ListRepo defines the interface for list repository operations
type ListRepo interface {
	Create(l *model.List) (int, error)
	Update(id int, l *model.List) error
	Delete(id int) error
	LinkSchedule(id int, scheduleID int) error
	LoadRules(id int) ([]model.Rule, error)
	LoadScheduleIDs(id int) ([]int, error)
	LoadRelations(l *model.List) error
	FindByID(id int) (*model.List, error)
	FindAll() ([]*model.List, error)
	FindByUser(userID int) ([]*model.List, error)
	FindBySchedule(scheduleID int) ([]*model.List, error)
}

// listRepo implements the ListRepo interface
type listRepo struct {
	db  *sql.DB
	ctx context.Context
}

// NewListRepo creates a new ListRepo instance
func NewListRepo(db *sql.DB) ListRepo {
	return &listRepo{db: db, ctx: context.Background()}
}

// Create stores a new list into the database
func (lr *listRepo) Create(l *model.List) (int, error) {
	query := "INSERT INTO list (name, description, source, user_id) VALUES (?, ?, ?, ?) RETURNING id"
	err := lr.db.QueryRowContext(lr.ctx, query, l.Name, l.Description, l.Source, l.UserID).Scan(&l.ID)
	return l.ID, err
}

// Update modifies an existing list with given ID in the database
func (lr *listRepo) Update(id int, l *model.List) error {
	query := "UPDATE list SET name=?, description=?, source=?, user_id=? WHERE id=?"
	_, err := lr.db.ExecContext(lr.ctx, query, l.Name, l.Description, l.Source, l.UserID, id)
	return err
}

// Delete removes an existing list with given ID from the database
func (lr *listRepo) Delete(id int) error {
	query := "DELETE FROM list WHERE id=?"
	_, err := lr.db.ExecContext(lr.ctx, query, id)
	return err
}

// LinkSchedule links a schedule with given ID to a list with given ID
func (lr *listRepo) LinkSchedule(id int, scheduleID int) error {
	query := "INSERT INTO list_schedule (list_id, schedule_id) VALUES (?, ?)"
	_, err := lr.db.ExecContext(lr.ctx, query, id, scheduleID)
	return err
}

// LoadRules returns all rules for a list with given id
func (lr *listRepo) LoadRules(id int) ([]model.Rule, error) {
	query := "SELECT r.id, r.domain_id, r.allowed FROM rule r JOIN list_rule lr ON r.id = lr.rule_id WHERE lr.list_id = ?"
	var rules []model.Rule

	err := sqlscan.Select(lr.ctx, lr.db, &rules, query, id)

	if err != nil || rules == nil {
		return []model.Rule{}, nil
	}

	return rules, nil
}

// LoadScheduleIDs returns ids of all schedules linked to list with given id
func (lr *listRepo) LoadScheduleIDs(id int) ([]int, error) {
	query := "SELECT DISTINCT ls.schedule_id from list l JOIN list_schedule ls ON l.id = ls.list_id WHERE l.id = ?"
	var scheduleIDs []int

	err := sqlscan.Select(lr.ctx, lr.db, &scheduleIDs, query, id)

	if err != nil || scheduleIDs == nil {
		return []int{}, nil
	}

	return scheduleIDs, nil
}

// LoadRelations loads all relations (rules, schedule ids) for list with given id
func (lr *listRepo) LoadRelations(l *model.List) error {
	var err error

	if l.Rules, err = lr.LoadRules(l.ID); err != nil {
		return err
	}

	if l.ScheduleIDs, err = lr.LoadScheduleIDs(l.ID); err != nil {
		return err
	}

	return nil
}

// FindByID returns an existing list with given id from the database
func (lr *listRepo) FindByID(id int) (*model.List, error) {
	query := "SELECT id, name, description, source, user_id FROM list WHERE id=?"
	var l model.List

	if err := sqlscan.Get(lr.ctx, lr.db, &l, query, id); err != nil {
		return nil, err
	}

	if err := lr.LoadRelations(&l); err != nil {
		return nil, err
	}

	return &l, nil
}

// FindAll returns all existing lists from the database
func (lr *listRepo) FindAll() ([]*model.List, error) {
	query := "SELECT id, name, description, source, user_id FROM list"
	var lists []*model.List

	err := sqlscan.Select(lr.ctx, lr.db, &lists, query)

	if err != nil || lists == nil {
		return []*model.List{}, nil
	}

	for _, l := range lists {
		if err := lr.LoadRelations(l); err != nil {
			return nil, err
		}
	}

	return lists, nil
}

// FindByUser returns all existing lists with given user ID from the database
func (lr *listRepo) FindByUser(userID int) ([]*model.List, error) {
	query := "SELECT id, name, description, source, user_id FROM list WHERE user_id=?"
	var lists []*model.List

	err := sqlscan.Select(lr.ctx, lr.db, &lists, query, userID)

	if err != nil || lists == nil {
		return []*model.List{}, nil
	}

	for _, l := range lists {
		if err := lr.LoadRelations(l); err != nil {
			return nil, err
		}
	}

	return lists, nil
}

// FindBySchedule returns all existing lists linked to schedule with given id
func (lr *listRepo) FindBySchedule(scheduleID int) ([]*model.List, error) {
	query := "SELECT l.id, l.name, l.description, l.source, l.user_id FROM list l JOIN list_schedule ls ON l.id = ls.list_id WHERE ls.schedule_id = ?"
	var lists []*model.List

	err := sqlscan.Select(lr.ctx, lr.db, &lists, query, scheduleID)

	if err != nil || lists == nil {
		return []*model.List{}, nil
	}

	for _, l := range lists {
		if err := lr.LoadRelations(l); err != nil {
			return nil, err
		}
	}

	return lists, nil
}
