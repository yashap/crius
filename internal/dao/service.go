package dao

import (
	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"github.com/yashap/crius/internal/errors"
)

// Service represents a Service database record
type Service struct {
	ID   int64  `db:"id"`
	Code string `db:"code"`
	Name string `db:"name"`
}

// ServiceQueries allows you to work with Service database records
type ServiceQueries interface {
	// Upsert upserts a Service
	Upsert(s Service) (int64, error)
	// GetByCode gets a Service by its code
	GetByCode(code string) (*Service, error)
}

// NewServiceQueries creates a new ServiceQueries instance
func NewServiceQueries(dbURL *dburl.URL, db *sqlx.DB) (ServiceQueries, error) {
	if dbURL.Driver == "postgres" {
		return &postgresServiceQueries{db}, nil
	}
	// TODO: mysql support
	return nil, errors.DatabaseError("Unsupported DB driver: "+dbURL.Driver, nil)
}

type postgresServiceQueries struct {
	db *sqlx.DB
}

func (q *postgresServiceQueries) Upsert(s Service) (int64, error) {
	rows, err := q.db.NamedQuery(
		`INSERT INTO service (code, name)
		VALUES (:code, :name)
		ON CONFLICT (code) DO UPDATE SET name = :name
		RETURNING id`,
		s,
	)
	if err != nil {
		return 0, err // TODO wrap
	}
	defer rows.Close()
	var id int64
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return 0, err // TODO wrap
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err // TODO wrap
	}
	return id, nil
}

func (q *postgresServiceQueries) GetByCode(code string) (*Service, error) {
	var service Service
	err := q.db.Get(&service, "SELECT * FROM service WHERE code = $1", code)
	// TODO not found?
	if err != nil {
		return nil, err // TODO wrap
	}
	return &service, nil
}
