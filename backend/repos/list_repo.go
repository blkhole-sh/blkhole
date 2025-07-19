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
	Create(c *model.List) (int, error)
	Update(id int, c *model.List) error
	Delete(id int) error
	LinkDomain(id int, domainID int) error
	LinkSchedule(id int, scheduleID int) error
	LoadDomainIDs(id int) ([]int, error)
	LoadScheduleIDs(id int) ([]int, error)
	LoadRelations(l *model.List) error
	FindByID(id int) (*model.List, error)
	FindByUser(userHash string) ([]*model.List, error)
	FindByDomain(domainID int) ([]*model.List, error)
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

// LinkDomain links a domain with given ID to a list with given ID
func (repo *ListRepoImpl) LinkDomain(id int, domainID int) error {
	sql := "INSERT INTO domain_list (domain_id, list_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, domainID, id)
	return err
}

// LinkSchedule links a schedule with given ID to a list with given ID
func (repo *ListRepoImpl) LinkSchedule(id int, scheduleID int) error {
	sql := "INSERT INTO list_schedule (list_id, schedule_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, id, scheduleID)
	return err
}

// LoadDomainIDs returns ids of all domains linked to list with given id
func (repo *ListRepoImpl) LoadDomainIDs(id int) ([]int, error) {
	sql := "SELECT DISTINCT dl.domain_id from domain_list dl JOIN list l ON dl.id = l.id WHERE l.id = ?"
	var domainIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &domainIds, sql, id); err != nil {
		return nil, err
	}

	return domainIds, nil
}

// LoadScheduleIDs returns ids of all schedules linked to list with given id
func (repo *ListRepoImpl) LoadScheduleIDs(id int) ([]int, error) {
	sql := "SELECT DISTINCT ls.schedule_od from list l JOIN list_schedule ls ON l.id = ls.list_id WHERE l.id = ?"
	var scheduleIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &scheduleIds, sql, id); err != nil {
		return nil, err
	}

	return scheduleIds, nil
}

// LoadRelations loads all relations (hashes, ids) for list with given id
func (repo *ListRepoImpl) LoadRelations(l *model.List) error {
	var err error

	if l.DomainIds, err = repo.LoadDomainIDs(l.ID); err != nil {
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

// FindByUser returns all existing lists with given user hash from the database
func (repo *ListRepoImpl) FindByUser(userHash string) ([]*model.List, error) {
	sql := "SELECT id, name, description, source, user_hash FROM list WHERE user_hash=?"
	var lists []*model.List

	if err := sqlscan.Select(repo.ctx, repo.db, &lists, sql, userHash); err != nil {
		return nil, err
	}

	for _, l := range lists {
		if err := repo.LoadRelations(l); err != nil {
			return nil, err
		}
	}

	return lists, nil
}

// FindByDomain returns all existing lists linked with domain with given id
func (repo *ListRepoImpl) FindByDomain(domainID int) ([]*model.List, error) {
	sql := "SELECT l.id, l.name, l.description, l.source, l.user_hash FROM domain_list dl JOIN list l on dl.list_id = l.id WHERE dl.domain_id = ?"
	var lists []*model.List

	if err := sqlscan.Select(repo.ctx, repo.db, &lists, sql, domainID); err != nil {
		return nil, err
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
		return nil, err
	}

	for _, l := range lists {
		if err := repo.LoadRelations(l); err != nil {
			return nil, err
		}
	}

	return lists, nil
}
