package repos

import (
	"context"
	"database/sql"
	"server/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// Define ListRepo interface
type ListRepo interface {
	Create(c *model.List) (int, error)
	Update(id int, c *model.List) error
	Delete(id int) error
	LinkDomain(id int, domainId int) error
	LinkSchedule(id int, scheduleId int) error
	LoadDomainIds(id int) ([]int, error)
	LoadScheduleIds(id int) ([]int, error)
	LoadRelations(l *model.List) error
	FindById(id int) (*model.List, error)
	FindByUser(userHash string) ([]*model.List, error)
	FindByDomain(domainId int) ([]*model.List, error)
	FindBySchedule(scheduleId int) ([]*model.List, error)
}

// Define ListRepoImpl
type ListRepoImpl struct {
	db  *sql.DB
	ctx context.Context
}

// Create new ListRepoImpl
func NewListRepo(db *sql.DB) ListRepo {
	return &ListRepoImpl{db: db, ctx: context.Background()}
}

// Store a new list into the db
func (repo *ListRepoImpl) Create(l *model.List) (int, error) {
	sql := "INSERT INTO list (name, user_hash) VALUES (?, ?) RETURNING id"
	err := repo.db.QueryRowContext(repo.ctx, sql, l.Name, l.UserHash).Scan(&l.Id)
	return l.Id, err
}

// Update an existing list with given id in the db
func (repo *ListRepoImpl) Update(id int, l *model.List) error {
	sql := "UPDATE list SET name=?, user_hash=? WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, l.Name, l.UserHash, id)
	return err
}

// Delete an existing list with given id from the db
func (repo *ListRepoImpl) Delete(id int) error {
	sql := "DELETE FROM list WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, id)
	return err
}

// Link a domain with given id to a list with given id
func (repo *ListRepoImpl) LinkDomain(id int, domainId int) error {
	sql := "INSERT INTO domain_list (domain_id, list_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, domainId, id)
	return err
}

// Link a schedule with given id to a list with given id
func (repo *ListRepoImpl) LinkSchedule(id int, scheduleId int) error {
	sql := "INSERT INTO list_schedule (list_id, schedule_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, id, scheduleId)
	return err
}

// Load ids of all domains linked to list with given id
func (repo *ListRepoImpl) LoadDomainIds(id int) ([]int, error) {
	sql := "SELECT DISTINCT dl.domain_id from domain_list dl JOIN list l ON dl.id = l.id WHERE l.id = ?"
	var domainIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &domainIds, sql, id); err != nil {
		return nil, err
	}

	return domainIds, nil
}

// Load ids of all schedules linked to list with given id
func (repo *ListRepoImpl) LoadScheduleIds(id int) ([]int, error) {
	sql := "SELECT DISTINCT ls.schedule_od from list l JOIN list_schedule ls ON l.id = ls.list_id WHERE l.id = ?"
	var scheduleIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &scheduleIds, sql, id); err != nil {
		return nil, err
	}

	return scheduleIds, nil
}

// Load all relations (hashes, ids) for list with given id
func (repo *ListRepoImpl) LoadRelations(l *model.List) error {
	var err error

	if l.DomainIds, err = repo.LoadDomainIds(l.Id); err != nil {
		return err
	}

	if l.ScheduleIds, err = repo.LoadScheduleIds(l.Id); err != nil {
		return err
	}

	return nil
}

// Find an existing list with given id in the db
func (repo *ListRepoImpl) FindById(id int) (*model.List, error) {
	sql := "SELECT id, name, user_hash FROM list WHERE id=?"
	var l model.List

	if err := sqlscan.Get(repo.ctx, repo.db, &l, sql, id); err != nil {
		return nil, err
	}

	if err := repo.LoadRelations(&l); err != nil {
		return nil, err
	}

	return &l, nil
}

// Find all existing lists with given user hash in the db
func (repo *ListRepoImpl) FindByUser(userHash string) ([]*model.List, error) {
	sql := "SELECT id, name, user_hash FROM list WHERE userHash=?"
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

// Find all existing lists linked with domain with given id in the db
func (repo *ListRepoImpl) FindByDomain(domainId int) ([]*model.List, error) {
	sql := "SELECT l.id, l.name, l.user_hash FROM domain_list dl JOIN list l on dl.list_id = l.id WHERE dl.domain_id = ?"
	var lists []*model.List

	if err := sqlscan.Select(repo.ctx, repo.db, &lists, sql, domainId); err != nil {
		return nil, err
	}

	for _, l := range lists {
		if err := repo.LoadRelations(l); err != nil {
			return nil, err
		}
	}

	return lists, nil
}

// Find all existing lists linked to schedule with given id in the db
func (repo *ListRepoImpl) FindBySchedule(scheduleId int) ([]*model.List, error) {
	sql := "SELECT l.id, l.name, l.user_hash FROM list l JOIN list_schedule ls ON l.id = ls.list_id WHERE ls.schedule_id = `"
	var lists []*model.List

	if err := sqlscan.Select(repo.ctx, repo.db, &lists, sql, scheduleId); err != nil {
		return nil, err
	}

	for _, l := range lists {
		if err := repo.LoadRelations(l); err != nil {
			return nil, err
		}
	}

	return lists, nil
}
