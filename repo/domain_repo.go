package repo

import (
	"database/sql"
	"server/model"

	_ "github.com/mattn/go-sqlite3"
)

// Define DomainRepo interface
type DomainRepo interface {
	Create(d *model.Domain) error
	Update(d *model.Domain) error
	Delete(d *model.Domain) error
	FindById(id string) (*model.Domain, error)
}

// Define DomainRepo struct
type DomainRepoImpl struct {
	DB *sql.DB
}

// Create new DomainRepoImpl
func NewDomainRepo(db *sql.DB) *DomainRepoImpl {
	return &DomainRepoImpl{DB: db}
}

// Store a new domain into the db
func (repo *DomainRepoImpl) Create(d *model.Domain) error {
	sql := "INSERT INTO Domain VALUES (?, ?)"
	_, err := repo.DB.Exec(sql, d.Id, d.Name)
	return err
}

// Update an existing domain in the db
func (repo *DomainRepoImpl) Update(d *model.Domain) error {
	sql := "UPDATE Domain SET name=? WHERE id=?"
	_, err := repo.DB.Exec(sql, d.Id, d.Id)
	return err
}

// Delete an existing domain from the db
func (repo *DomainRepoImpl) Delete(d *model.Domain) error {
	sql := "DELETE FROM Domain WHERE id=?"
	_, err := repo.DB.Exec(sql, d.Id)
	return err
}

// Find an existing domain with given id in db
func (repo *DomainRepoImpl) FindById(id int) (*model.Domain, error) {
	sql := "SELECT * FROM Domain WHERE id=?"

	domain := model.Domain{}
	err := repo.DB.QueryRow(sql, id).Scan(&domain.Id, &domain.Name)
	if err != nil {
		return nil, err
	}

	return &domain, nil
}
