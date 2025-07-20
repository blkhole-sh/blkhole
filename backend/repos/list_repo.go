package repos

import (
	"context"
	"database/sql"
	"server/model"

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
	FindByUser(userHash string) ([]*model.List, error)
	FindBySchedule(scheduleID int) ([]*model.List, error)
}

// ListRepoImpl implements the ListRepo interface
type ListRepoImpl struct {
	db  *sql.DB
	ctx context.Context
}

// NewListRepo creates a new ListRepo instance
func NewListRepo(db *sql.DB) ListRepo {
	return &ListRepoImpl{db: db, ctx: context.Background()}
}

// Create stores a new list into the database
func (repo *ListRepoImpl) Create(l *model.List) (int, error) {
	sql := "INSERT INTO list (name, description, source, user_hash) VALUES (?, ?, ?, ?) RETURNING id"
	err := repo.db.QueryRowContext(repo.ctx, sql, l.Name, l.Description, l.Source, l.UserHash).Scan(&l.ID)
	return l.ID, err
}

// Update modifies an existing list with given ID in the database
func (repo *ListRepoImpl) Update(id int, l *model.List) error {
	sql := "UPDATE list SET name=?, description=?, source=?, user_hash=? WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, l.Name, l.Description, l.Source, l.UserHash, id)
	return err
}

// Delete removes an existing list with given ID from the database
func (repo *ListRepoImpl) Delete(id int) error {
	sql := "DELETE FROM list WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, id)
	return err
}

// LinkSchedule links a schedule with given ID to a list with given ID
func (repo *ListRepoImpl) LinkSchedule(id int, scheduleID int) error {
	sql := "INSERT INTO list_schedule (list_id, schedule_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, id, scheduleID)
	return err
}

// LoadRules returns all rules for a list with given id
func (repo *ListRepoImpl) LoadRules(id int) ([]model.Rule, error) {
	sql := "SELECT id, domain, list_id, allowed FROM rule WHERE list_id = ?"
	var rules []model.Rule

	if err := sqlscan.Select(repo.ctx, repo.db, &rules, sql, id); err != nil {
		return []model.Rule{}, nil
	}

	// Ensure we return empty slice instead of nil
	if rules == nil {
		return []model.Rule{}, nil
	}

	return rules, nil
}

// LoadScheduleIDs returns ids of all schedules linked to list with given id
func (repo *ListRepoImpl) LoadScheduleIDs(id int) ([]int, error) {
	sql := "SELECT DISTINCT ls.schedule_id from list l JOIN list_schedule ls ON l.id = ls.list_id WHERE l.id = ?"
	var scheduleIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &scheduleIds, sql, id); err != nil {
		return []int{}, nil
	}

	// Ensure we return empty slice instead of nil
	if scheduleIds == nil {
		return []int{}, nil
	}

	return scheduleIds, nil
}

// LoadRelations loads all relations (rules, schedule ids) for list with given id
func (repo *ListRepoImpl) LoadRelations(l *model.List) error {
	var err error

	if l.Rules, err = repo.LoadRules(l.ID); err != nil {
		return err
	}

	if l.ScheduleIds, err = repo.LoadScheduleIDs(l.ID); err != nil {
		return err
	}

	return nil
}

// FindByID returns an existing list with given id from the database
func (repo *ListRepoImpl) FindByID(id int) (*model.List, error) {
	sql := "SELECT id, name, description, source, user_hash FROM list WHERE id=?"
	var l model.List

	if err := sqlscan.Get(repo.ctx, repo.db, &l, sql, id); err != nil {
		return nil, err
	}

	if err := repo.LoadRelations(&l); err != nil {
		return nil, err
	}

	return &l, nil
}

// FindAll returns all existing lists from the database
func (repo *ListRepoImpl) FindAll() ([]*model.List, error) {
	sql := "SELECT id, name, description, source, user_hash FROM list"
	var lists []*model.List

	if err := sqlscan.Select(repo.ctx, repo.db, &lists, sql); err != nil {
		return []*model.List{}, nil
	}

	// Ensure we return empty slice instead of nil
	if lists == nil {
		return []*model.List{}, nil
	}

	for _, l := range lists {
		if err := repo.LoadRelations(l); err != nil {
			return nil, err
		}
	}

	return lists, nil
}

// FindByUser returns all existing lists with given user hash from the database
func (repo *ListRepoImpl) FindByUser(userHash string) ([]*model.List, error) {
	sql := "SELECT id, name, description, source, user_hash FROM list WHERE user_hash=?"
	var lists []*model.List

	if err := sqlscan.Select(repo.ctx, repo.db, &lists, sql, userHash); err != nil {
		return []*model.List{}, nil
	}

	// Ensure we return empty slice instead of nil
	if lists == nil {
		return []*model.List{}, nil
	}

	for _, l := range lists {
		if err := repo.LoadRelations(l); err != nil {
			return nil, err
		}
	}

	return lists, nil
}

// FindBySchedule returns all existing lists linked to schedule with given id
func (repo *ListRepoImpl) FindBySchedule(scheduleID int) ([]*model.List, error) {
	sql := "SELECT l.id, l.name, l.description, l.source, l.user_hash FROM list l JOIN list_schedule ls ON l.id = ls.list_id WHERE ls.schedule_id = ?"
	var lists []*model.List

	if err := sqlscan.Select(repo.ctx, repo.db, &lists, sql, scheduleID); err != nil {
		return []*model.List{}, nil
	}

	// Ensure we return empty slice instead of nil
	if lists == nil {
		return []*model.List{}, nil
	}

	for _, l := range lists {
		if err := repo.LoadRelations(l); err != nil {
			return nil, err
		}
	}

	return lists, nil
}
