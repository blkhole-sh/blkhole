package repos

import (
	"context"
	"database/sql"
	"server/model"

	"github.com/georgysavva/scany/v2/sqlscan"
	_ "github.com/mattn/go-sqlite3"
)

// DomainRepo defines the interface for domain repository operations
type DomainRepo interface {
	Create(d *model.Domain) (int, error)
	Update(id int, c *model.Domain) error
	Delete(id int) error
	LinkList(id int, listID int) error
	LinkSchedule(id int, scheduleID int) error
	LoadListIDs(id int) ([]int, error)
	LoadScheduleIDs(id int) ([]int, error)
	LoadRelations(d *model.Domain) error
	FindByID(id int) (*model.Domain, error)
	FindByList(listID int) ([]*model.Domain, error)
	FindBySchedule(scheduleID int) ([]*model.Domain, error)
}

// DomainRepoImpl implements the DomainRepo interface
type DomainRepoImpl struct {
	db  *sql.DB
	ctx context.Context
}

// NewDomainRepo creates a new DomainRepo instance
func NewDomainRepo(db *sql.DB) DomainRepo {
	return &DomainRepoImpl{db: db, ctx: context.Background()}
}

// Create stores a new domain into the database
func (repo *DomainRepoImpl) Create(d *model.Domain) (int, error) {
	sql := "INSERT INTO domain (name) VALUES (?) RETURNING id"

	err := repo.db.QueryRowContext(repo.ctx, sql, d.Name).Scan(&d.ID)
	return d.ID, err
}

// Update modifies an existing domain with given ID in the database
func (repo *DomainRepoImpl) Update(id int, d *model.Domain) error {
	sql := "UPDATE domain SET name=? WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, d.Name, id)
	return err
}

// Delete removes an existing domain with given ID from the database
func (repo *DomainRepoImpl) Delete(id int) error {
	sql := "DELETE FROM domain WHERE id=?"
	_, err := repo.db.ExecContext(repo.ctx, sql, id)
	return err
}

// LinkList links a list with given ID to a domain with given ID
func (repo *DomainRepoImpl) LinkList(id int, listID int) error {
	sql := "INSERT INTO domain_list (domain_id, list_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, id, listID)
	return err
}

// LinkSchedule links a schedule with given ID to a domain with given ID
func (repo *DomainRepoImpl) LinkSchedule(id int, scheduleID int) error {
	sql := "INSERT INTO domain_schedule (domain_id, schedule_id) VALUES (?, ?)"
	_, err := repo.db.ExecContext(repo.ctx, sql, id, scheduleID)
	return err
}

// LoadListIDs returns ids of all lists linked to domain with given id
func (repo *DomainRepoImpl) LoadListIDs(id int) ([]int, error) {
	sql := "SELECT DISTINCT dl.list_id FROM domain d JOIN domain_list dl ON d.id = dl.domain_id WHERE d.id = ?"
	var listIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &listIds, sql, id); err != nil {
		return nil, err
	}

	return listIds, nil
}

// LoadScheduleIDs returns ids of all schedules linked to domain with given id
func (repo *DomainRepoImpl) LoadScheduleIDs(id int) ([]int, error) {
	sql := "SELECT DISTINCT ds.schedule_id FROM domain d JOIN domain_schedule ds ON d.id = ds.domain_id WHERE d.id = ?"
	var scheduleIds []int

	if err := sqlscan.Select(repo.ctx, repo.db, &scheduleIds, sql, id); err != nil {
		return nil, err
	}

	return scheduleIds, nil
}

// LoadRelations loads all relations (hashes, ids) for domain with given id
func (repo *DomainRepoImpl) LoadRelations(d *model.Domain) error {
	var err error

	if d.ListIds, err = repo.LoadListIDs(d.ID); err != nil {
		return err
	}

	if d.ScheduleIds, err = repo.LoadScheduleIDs(d.ID); err != nil {
		return err
	}

	return nil
}

// FindByID returns an existing domain with given id from the database
func (repo *DomainRepoImpl) FindByID(id int) (*model.Domain, error) {
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

// FindByList returns all existing domains that are linked with list with given id
func (repo *DomainRepoImpl) FindByList(listID int) ([]*model.Domain, error) {
	sql := "SELECT DISTINCT d.id d.name FROM domain d JOIN domain_list dl ON d.id = dl.domain_id WHERE dl.list_id=?"
	var domains []*model.Domain

	if err := sqlscan.Select(repo.ctx, repo.db, &domains, sql, listID); err != nil {
		return nil, err
	}

	for _, d := range domains {
		if err := repo.LoadRelations(d); err != nil {
			return nil, err
		}
	}

	return domains, nil
}

// FindBySchedule returns all existing domains that are linked with schedule with given id
func (repo *DomainRepoImpl) FindBySchedule(scheduleID int) ([]*model.Domain, error) {
	sql := "SELECT DISTINCT d.id d.name FROM domain d JOIN domain_schedule ds ON d.id = ds.domain_id WHERE ds.schedule_id=?"
	var domains []*model.Domain

	if err := sqlscan.Select(repo.ctx, repo.db, &domains, sql, scheduleID); err != nil {
		return nil, err
	}

	for _, d := range domains {
		if err := repo.LoadRelations(d); err != nil {
			return nil, err
		}
	}

	return domains, nil
}
