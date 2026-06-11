package repos

import (
	"context"
	"database/sql"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/blkhole-sh/blkhole/internal/model"
)

// DomainRepo defines the interface for domain repository operations
type DomainRepo interface {
	CreateOrGet(d *model.Domain) (int, error)
	Update(id int, d *model.Domain) error
	Delete(id int) error
	FindByID(id int) (*model.Domain, error)
	FindAll() ([]*model.Domain, error)
}

// domainRepo implements the DomainRepo interface
type domainRepo struct {
	db  *sql.DB
	ctx context.Context
}

// NewDomainRepo creates a new DomainRepo instance
func NewDomainRepo(db *sql.DB) DomainRepo {
	return &domainRepo{db: db, ctx: context.Background()}
}

// CreateOrGet stores a new domain into the database or returns existing one
func (dr *domainRepo) CreateOrGet(d *model.Domain) (int, error) {
	// First, try to find existing domain
	query := "SELECT id FROM domain WHERE name = ?"
	err := dr.db.QueryRowContext(dr.ctx, query, d.Name).Scan(&d.ID)
	if err == nil {
		// Domain already exists, return its ID
		return d.ID, nil
	}

	// Domain doesn't exist, create it
	if err == sql.ErrNoRows {
		query := "INSERT INTO domain (name) VALUES (?) RETURNING id"
		err = dr.db.QueryRowContext(dr.ctx, query, d.Name).Scan(&d.ID)
		return d.ID, err
	}

	// Other error occurred
	return 0, err
}

// Update modifies an existing domain with given ID in the database if not referenced
func (dr *domainRepo) Update(id int, d *model.Domain) error {
	query := `UPDATE domain SET name=?  WHERE id = ? AND id NOT IN (
		SELECT DISTINCT domain_id FROM rule
	)`

	_, err := dr.db.ExecContext(dr.ctx, query, d.Name, id)
	if err != nil {
		return err
	}

	return err
}

// Delete removes an existing domain with given ID from the database if not referenced
func (dr *domainRepo) Delete(id int) error {
	query := `DELETE FROM domain WHERE id = ? AND id NOT IN (
		SELECT DISTINCT domain_id FROM rule 
	)`

	_, err := dr.db.ExecContext(dr.ctx, query, id)
	if err != nil {
		return err
	}

	return err
}

// FindAll returns all existing domain from the database
func (dr *domainRepo) FindAll() ([]*model.Domain, error) {
	query := "SELECT id, name FROM domain"
	var domains []*model.Domain

	err := sqlscan.Select(dr.ctx, dr.db, &domains, query)
	if err != nil {
		return nil, err
	}

	if domains == nil {
		return []*model.Domain{}, nil
	}

	return domains, nil
}

// FindByID returns an existing domain with given id from the database
func (dr *domainRepo) FindByID(id int) (*model.Domain, error) {
	query := "SELECT id, name FROM domain WHERE id=?"
	var d model.Domain

	if err := sqlscan.Get(dr.ctx, dr.db, &d, query, id); err != nil {
		return nil, err
	}

	return &d, nil
}
