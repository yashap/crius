package dao

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"github.com/yashap/crius/internal/errors"
)

// ServiceEndpoint represents a ServiceEndpoint database record
type ServiceEndpoint struct {
	ID        int64  `db:"id"`
	ServiceID int64  `db:"service_id"`
	Code      string `db:"code"`
	Name      string `db:"name"`
}

// ServiceEndpointQueries allows you to work with Service database records
type ServiceEndpointQueries interface {
	// Upsert upserts a Service
	Upsert(s ServiceEndpoint) (int64, error)
}

// NewServiceEndpointQueries creates a new ServiceEndpointQueries instance
func NewServiceEndpointQueries(dbURL *dburl.URL, db *sqlx.DB) (ServiceEndpointQueries, error) {
	if dbURL.Driver == "postgres" {
		return &postgresServiceEndpointQueries{db}, nil
	}
	// TODO: mysql support
	return nil, errors.DatabaseError("Unsupported DB driver: "+dbURL.Driver, nil)
}

type postgresServiceEndpointQueries struct {
	db *sqlx.DB
}

func (q *postgresServiceEndpointQueries) Upsert(s ServiceEndpoint) (int64, error) {
	fmt.Printf(">>>>>> 4 %v\n\n", s)
	rows, err := q.db.NamedQuery(
		`INSERT INTO service_endpoint (service_id, code, name)
		VALUES (:service_id, :code, :name)
		ON CONFLICT (service_id, code) DO UPDATE SET name = :name
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
