package dao

import (
	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"github.com/yashap/crius/internal/errors"
	"go.uber.org/zap"
)

// ServiceEndpointDependency represents a ServiceEndpointDependency database record
type ServiceEndpointDependency struct {
	ID                          int64 `db:"id"`
	ServiceEndpointID           int64 `db:"service_endpoint_id"`
	DependencyServiceEndpointID int64 `db:"dependency_service_endpoint_id"`
}

// ServiceEndpointDependencyQueries allows you to work with ServiceEndpointDependency database records
type ServiceEndpointDependencyQueries interface {
	// Upsert upserts a Service
	Upsert(s ServiceEndpointDependency) (int64, error)
	// FindByServiceEndpointID finds ServiceEndpointDependencys by serviceEndpointID
	FindByServiceEndpointID(serviceEndpointID int64) ([]ServiceEndpointDependency, error)
}

// NewServiceEndpointDependencyQueries creates a new ServiceEndpointDependencyQueries instance
func NewServiceEndpointDependencyQueries(
	dbURL *dburl.URL,
	db *sqlx.DB,
	logger *zap.SugaredLogger,
) (ServiceEndpointDependencyQueries, error) {
	if dbURL.Driver == "postgres" {
		return &postgresServiceEndpointDependencyQueries{logger, db}, nil
	}
	// TODO: mysql support
	msg := "Unsupported DB driver"
	logger.Errorw(msg, "driver", dbURL.Driver)
	return nil, errors.DatabaseError(msg, nil)
}

type postgresServiceEndpointDependencyQueries struct {
	logger *zap.SugaredLogger
	db     *sqlx.DB
}

func (q *postgresServiceEndpointDependencyQueries) Upsert(s ServiceEndpointDependency) (int64, error) {
	rows, err := q.db.NamedQuery(
		`INSERT INTO service_endpoint_dependency (service_endpoint_id, dependency_service_endpoint_id)
		VALUES (:service_endpoint_id, :dependency_service_endpoint_id)
		ON CONFLICT (service_endpoint_id, dependency_service_endpoint_id) DO NOTHING
		RETURNING id`,
		s,
	)
	if err != nil {
		msg := "Failed to execute query when upserting Service Endpoint Dependency"
		q.logger.Errorw(msg, "err", err.Error(), "service", s)
		return 0, errors.DatabaseError(msg, &err)
	}
	defer rows.Close()
	var id int64
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			msg := "Failed to scan for rows when upserting Service Endpoint Dependency"
			q.logger.Errorw(msg, "err", err.Error(), "service", s)
			return 0, errors.DatabaseError(msg, &err)
		}
	}
	if err := rows.Err(); err != nil {
		msg := "Failure upserting Service Endpoint Dependency"
		q.logger.Errorw(msg, "err", err.Error(), "service", s)
		return 0, errors.DatabaseError(msg, &err)
	}
	return id, nil
}

func (q *postgresServiceEndpointDependencyQueries) FindByServiceEndpointID(
	serviceEndpointID int64,
) ([]ServiceEndpointDependency, error) {
	endpoints := make([]ServiceEndpointDependency, 0)
	err := q.db.Select(
		&endpoints,
		"SELECT * FROM service_endpoint_dependency WHERE service_endpoint_id = $1",
		serviceEndpointID,
	)
	if err != nil {
		msg := "Failed to get Service Endpoint Dependencies from DB by service endpoint id"
		q.logger.Errorw(msg, "err", err.Error(), "serviceEndpointID", serviceEndpointID)
		return nil, errors.DatabaseError(msg, &err)
	}
	return endpoints, nil
}
