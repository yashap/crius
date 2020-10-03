package dao

import (
	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"github.com/yashap/crius/internal/errors"
	"go.uber.org/zap"
)

// ServiceEndpoint represents a ServiceEndpoint database record
type ServiceEndpoint struct {
	ID        int64  `db:"id"`
	ServiceID int64  `db:"service_id"`
	Code      string `db:"code"`
	Name      string `db:"name"`
}

// ServiceEndpointQueries allows you to work with ServiceEndpoint database records
type ServiceEndpointQueries interface {
	// Upsert upserts a Service
	Upsert(s ServiceEndpoint) (int64, error)
	// FindByServiceID find ServiceEndpoints by Service id
	FindByServiceID(serviceID int64) ([]ServiceEndpoint, error)
	// FindByIDs finds ServiceEndpoints by ids
	FindByIDs(ids []int64) ([]ServiceEndpoint, error)
}

// NewServiceEndpointQueries creates a new ServiceEndpointQueries instance
func NewServiceEndpointQueries(
	dbURL *dburl.URL,
	db *sqlx.DB,
	logger *zap.SugaredLogger,
) (ServiceEndpointQueries, error) {
	if dbURL.Driver == "postgres" {
		return &postgresServiceEndpointQueries{logger, db}, nil
	}
	// TODO: mysql support
	msg := "Unsupported DB driver"
	logger.Errorw(msg, "driver", dbURL.Driver)
	return nil, errors.DatabaseError(msg, nil)
}

type postgresServiceEndpointQueries struct {
	logger *zap.SugaredLogger
	db     *sqlx.DB
}

func (q *postgresServiceEndpointQueries) Upsert(s ServiceEndpoint) (int64, error) {
	rows, err := q.db.NamedQuery(
		`INSERT INTO service_endpoint (service_id, code, name)
		VALUES (:service_id, :code, :name)
		ON CONFLICT (service_id, code) DO UPDATE SET name = :name
		RETURNING id`,
		s,
	)
	if err != nil {
		msg := "Failed to execute query when upserting Service Endpoint"
		q.logger.Errorw(msg, "err", err.Error(), "service", s)
		return 0, errors.DatabaseError(msg, &err)
	}
	defer rows.Close()
	var id int64
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			msg := "Failed to scan for rows when upserting Service Endpoint"
			q.logger.Errorw(msg, "err", err.Error(), "service", s)
			return 0, errors.DatabaseError(msg, &err)
		}
	}
	if err := rows.Err(); err != nil {
		msg := "Failure upserting Service Endpoint"
		q.logger.Errorw(msg, "err", err.Error(), "service", s)
		return 0, errors.DatabaseError(msg, &err)
	}
	return id, nil
}

func (q *postgresServiceEndpointQueries) FindByServiceID(serviceID int64) ([]ServiceEndpoint, error) {
	endpoints := make([]ServiceEndpoint, 0)
	err := q.db.Select(&endpoints, "SELECT * FROM service_endpoint WHERE service_id = $1", serviceID)
	if err != nil {
		msg := "Failed to get Service Endpoints from DB by service id"
		q.logger.Errorw(msg, "err", err.Error(), "serviceID", serviceID)
		return nil, errors.DatabaseError(msg, &err)
	}
	return endpoints, nil
}

func (q *postgresServiceEndpointQueries) FindByIDs(ids []int64) ([]ServiceEndpoint, error) {
	endpoints := make([]ServiceEndpoint, 0)
	if len(ids) == 0 {
		return endpoints, nil
	}
	query, args, err := sqlx.In("SELECT * FROM service_endpoint WHERE id IN (?)", ids)
	if err != nil {
		msg := "Failed to get generate IN clause for ServiceEndpointQueries.FindByIDs"
		q.logger.Errorw(msg, "err", err.Error(), "ids", ids)
		return nil, errors.DatabaseError(msg, &err)
	}
	query = q.db.Rebind(query)
	err = q.db.Select(&endpoints, query, args...)
	if err != nil {
		msg := "Failed to find Service Endpoints from DB by ids"
		q.logger.Errorw(msg, "err", err.Error(), "ids", ids)
		return nil, errors.DatabaseError(msg, &err)
	}
	return endpoints, nil
}
