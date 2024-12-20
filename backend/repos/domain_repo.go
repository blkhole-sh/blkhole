package repos

import (
	"context"
	"database/sql"
	"server/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// Define DomainRepo interface
type DomainRepo interface {
	Create(d *model.Domain) (int, error)
	Update(id int, c *model.Domain) error
	Delete(id int) error
	LinkList(id int, listId int) error
	LinkSchedule(id int, scheduleId int) error
	LoadListIds(id int) ([]int, error)
	LoadScheduleIds(id int) ([]int, error)
	LoadRelations(d *model.Domain) error
	FindById(id int) (*model.Domain, error)
	FindByList(listId int) ([]*model.Domain, error)
	FindBySchedule(scheduleId int) ([]*model.Domain, error)
}

// Define DomainRepoImpl
type DomainRepoImpl struct {
	db  *sql.DB
	ctx context.Context
}

// Create new DomainRepoImpl
func NewDomainRepo(db *sql.DB) DomainRepo {
	return &DomainRepoImpl{db: db, ctx: context.Background()}
}

// Store a new domain into the db
func (repo *DomainRepoImpl) Create(d *model.Domain) (int, error) {
	sql := "INSERT INTO domain (name) VALUES (?) RETURNING id"

	err := repo.db.QueryRowContext(repo.ctx, sql, d.Name).Scan(&d.Id)
	return d.Id, err
}

// Update an existing domain with given id in the db
func (repo *DomainRepoImpl) Update(id int, d *model.Domain) error {
	sql := "UPDATE domain SET name=? WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, d.Name, id)
	return err
}

// Delete an existing domain with given id from the db
func (repo *DomainRepoImpl) Delete(id int) error {
	sql := "DELETE FROM domain WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, id)
	return err
}

// Link a list with given id to a domain with given id
func (repo *DomainRepoImpl) LinkList(id int, listId int) error {
	sql := "INSERT INTO domain_list (domain_id, list_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, id, listId)
	return err
}

// Link a schedule with given id to a domain with given id
func (repo *DomainRepoImpl) LinkSchedule(id int, scheduleId int) error {
	sql := "INSERT INTO domain_schedule (domain_id, schedule_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, id, scheduleId)
	return err
}

// Load ids of all lists linked to domain with given id
func (repo *DomainRepoImpl) LoadListIds(id int) ([]int, error) {
	sql := "SELECT DISTINCT dl.list_id FROM domain d JOIN domain_list dl ON d.id = dl.domain_id WHERE d.id = ?"
	var listIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &listIds, sql, id); err != nil {
		return nil, err
	}

	return listIds, nil
}

// Load ids of all schedules linked to domain with given id
func (repo *DomainRepoImpl) LoadScheduleIds(id int) ([]int, error) {
	sql := "SELECT DISTINCT ds.schedule_id FROM domain d JOIN domain_schedule ds ON d.id = ds.domain_id WHERE d.id = ?"
	var scheduleIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &scheduleIds, sql, id); err != nil {
		return nil, err
	}

	return scheduleIds, nil
}

// Load all relations (hashes, ids) for domain with given id
func (repo *DomainRepoImpl) LoadRelations(d *model.Domain) error {
	var err error

	if d.ListIds, err = repo.LoadListIds(d.Id); err != nil {
		return err
	}

	if d.ScheduleIds, err = repo.LoadScheduleIds(d.Id); err != nil {
		return err
	}

	return nil
}

// Find an existing domain with given id in the db
func (repo *DomainRepoImpl) FindById(id int) (*model.Domain, error) {
	sql := "SELECT id, name FROM domain WHERE id = ?"
	var d model.Domain

	err := sqlscan.Get(repo.ctx, repo.db, &d, sql, id)
	if err != nil {
		return nil, err
	}

	if err = repo.LoadRelations(&d); err != nil {
		return nil, err
	}

	return &d, nil
}

// Find all existing domains that are linked with list with given id in db
func (repo *DomainRepoImpl) FindByList(listId int) ([]*model.Domain, error) {
	sql := "SELECT DISTINCT d.id d.name FROM domain d JOIN domain_list dl ON d.id = dl.domain_id WHERE dl.list_id=?"
	var domains []*model.Domain

	if err := sqlscan.Select(repo.ctx, repo.db, &domains, sql, listId); err != nil {
		return nil, err
	}

	for _, d := range domains {
		if err := repo.LoadRelations(d); err != nil {
			return nil, err
		}
	}

	return domains, nil
}

// Find all existing domains that are linked with schedule with given id in db
func (repo *DomainRepoImpl) FindBySchedule(scheduleId int) ([]*model.Domain, error) {
	sql := "SELECT DISTINCT d.id d.name FROM domain d JOIN domain_schedule ds ON d.id = ds.domain_id WHERE ds.schedule_id=?"
	var domains []*model.Domain

	if err := sqlscan.Select(repo.ctx, repo.db, &domains, sql, scheduleId); err != nil {
		return nil, err
	}

	for _, d := range domains {
		if err := repo.LoadRelations(d); err != nil {
			return nil, err
		}
	}

	return domains, nil
}
