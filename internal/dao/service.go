package dao

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"github.com/yashap/crius/internal/errors"
	"go.uber.org/zap"
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
	// FindByIDs finds Services by ids
	FindByIDs(ids []int64) ([]Service, error)
}

// NewServiceQueries creates a new ServiceQueries instance
func NewServiceQueries(
	dbURL *dburl.URL,
	db *sqlx.DB,
	logger *zap.SugaredLogger,
) (ServiceQueries, error) {
	if dbURL.Driver == "postgres" {
		return &postgresServiceQueries{logger, db}, nil
	}
	// TODO: mysql support
	msg := "Unsupported DB driver"
	logger.Errorw(msg, "driver", dbURL.Driver)
	return nil, errors.DatabaseError(msg, nil)
}

type postgresServiceQueries struct {
	logger *zap.SugaredLogger
	db     *sqlx.DB
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
		msg := "Failed to execute query when upserting Service"
		q.logger.Errorw(msg, "err", err.Error(), "service", s)
		return 0, errors.DatabaseError(msg, &err)
	}
	defer rows.Close()
	var id int64
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			msg := "Failed to scan for rows when upserting Service"
			q.logger.Errorw(msg, "err", err.Error(), "service", s)
			return 0, errors.DatabaseError(msg, &err)
		}
	}
	if err := rows.Err(); err != nil {
		msg := "Failure upserting Service"
		q.logger.Errorw(msg, "err", err.Error(), "service", s)
		return 0, errors.DatabaseError(msg, &err)
	}
	return id, nil
}

func (q *postgresServiceQueries) GetByCode(code string) (*Service, error) {
	var service Service
	err := q.db.Get(&service, "SELECT * FROM service WHERE code = $1", code)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		msg := "Failed to get Service from DB by code"
		q.logger.Errorw(msg, "err", err.Error(), "code", code)
		return nil, errors.DatabaseError(msg, &err)
	}
	return &service, nil
}

func (q *postgresServiceQueries) FindByIDs(ids []int64) ([]Service, error) {
	services := make([]Service, 0)
	if len(ids) == 0 {
		return services, nil
	}
	query, args, err := sqlx.In("SELECT * FROM service WHERE id IN (?)", ids)
	if err != nil {
		msg := "Failed to get generate IN clause for ServiceQueries.FindByIDs"
		q.logger.Errorw(msg, "err", err.Error(), "ids", ids)
		return nil, errors.DatabaseError(msg, &err)
	}
	query = q.db.Rebind(query)
	err = q.db.Select(&services, query, args...)
	if err != nil {
		msg := "Failed to find Services from DB by ids"
		q.logger.Errorw(msg, "ids", ids)
		return nil, errors.DatabaseError(msg, &err)
	}
	return services, nil
}
