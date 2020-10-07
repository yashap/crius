package service

import (
	"context"
	"database/sql"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/yashap/crius/internal/db"
	pgdao "github.com/yashap/crius/internal/db/postgresql/dao"
	"github.com/yashap/crius/internal/errors"
	"go.uber.org/zap"
)

type postgresRepository struct {
	database db.Database
	logger   *zap.SugaredLogger
}

func (r *postgresRepository) Save(ctx context.Context, s *Service) error {
	exec := r.database.GetExecutor(ctx)
	serviceDAO := pgdao.Service{Code: s.Code, Name: s.Name}
	err := r.upsertService(exec, &serviceDAO)
	if err != nil {
		return err
	}
	s.ID = &serviceDAO.ID
	endpointIDs := make([]interface{}, len(s.Endpoints))
	for idx, endpoint := range s.Endpoints {
		endpointDAO := pgdao.ServiceEndpoint{
			ServiceID: serviceDAO.ID,
			Code:      endpoint.Code,
			Name:      endpoint.Name,
		}
		err = r.upsertEndpoint(exec, &endpointDAO)
		endpoint.ID = &endpointDAO.ID
		endpointIDs[idx] = endpointDAO.ID
		if err != nil {
			return err
		}
		for depServiceCode, depEndpointCodes := range endpoint.Dependencies {
			depService, err := r.FindByCode(ctx, depServiceCode)
			if err != nil {
				return err
			}
			for _, depEndpointCode := range depEndpointCodes {
				depEndpoint, err := r.findEndpointByServiceIDAndCode(exec, *depService.ID, depEndpointCode)
				if err != nil {
					return err
				}
				dependencyDAO := pgdao.ServiceEndpointDependency{
					ServiceEndpointID:           endpointDAO.ID,
					DependencyServiceEndpointID: depEndpoint.ID,
				}
				err = r.upsertDependency(exec, &dependencyDAO)
				if err != nil {
					return err
				}
			}
		}
	}
	// We want to fully replace the service, so remove any endpoints that no longer exist
	_, err = pgdao.ServiceEndpoints(
		qm.Where("service_id = ?", serviceDAO.ID),
		qm.AndNotIn("id not in ?", endpointIDs...),
	).DeleteAll(ctx, exec)
	if err != nil {
		return errors.DatabaseError("Failed to delete endpoints by ids",
			errors.Details{"ids": endpointIDs}, &err,
		).Logged(r.logger)
	}
	return nil
}

func (r *postgresRepository) FindByCode(ctx context.Context, code Code) (*Service, error) {
	exec := r.database.GetExecutor(ctx)
	serviceDAO, err := pgdao.Services(
		qm.Load(qm.Rels(
			pgdao.ServiceRels.ServiceEndpoints,
			pgdao.ServiceEndpointRels.ServiceEndpointDependencies,
		)),
		qm.Where("code = ?", code),
	).One(ctx, exec)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.DatabaseError("Failed to find service by code",
			errors.Details{"code": code}, &err,
		).Logged(r.logger)
	}
	endpoints := make([]Endpoint, len(serviceDAO.R.ServiceEndpoints))
	for idx, endpointDAO := range serviceDAO.R.ServiceEndpoints {
		dependencies := make(map[Code][]EndpointCode)
		for _, dependencyDAO := range endpointDAO.R.ServiceEndpointDependencies {
			depEndpointDAO, err := pgdao.ServiceEndpoints(
				qm.Where("id = ?", dependencyDAO.DependencyServiceEndpointID),
			).One(ctx, exec)
			if err != nil {
				return nil, errors.DatabaseError("Failed to find service endpoint by id",
					errors.Details{"id": dependencyDAO.DependencyServiceEndpointID}, &err,
				).Logged(r.logger)
			}
			depServiceDAO, err := pgdao.Services(
				qm.Where("id = ?", depEndpointDAO.ServiceID),
			).One(ctx, exec)
			if err != nil {
				return nil, errors.DatabaseError("Failed to find service by id",
					errors.Details{"id": depEndpointDAO.ServiceID}, &err,
				).Logged(r.logger)
			}
			depEndpoints, ok := dependencies[depServiceDAO.Code]
			if ok {
				depEndpoints = append(depEndpoints, depEndpointDAO.Code)
			} else {
				depEndpoints = []EndpointCode{depEndpointDAO.Code}
			}
			dependencies[depServiceDAO.Code] = depEndpoints
		}
		endpoints[idx] = Endpoint{
			Code:         endpointDAO.Code,
			Name:         endpointDAO.Name,
			Dependencies: dependencies,
		}
	}
	service := Service{
		ID:        &serviceDAO.ID,
		Code:      serviceDAO.Code,
		Name:      serviceDAO.Name,
		Endpoints: endpoints,
	}
	return &service, nil
}

func (r *postgresRepository) findEndpointByServiceIDAndCode(
	exec boil.ContextExecutor,
	serviceID int64,
	code EndpointCode,
) (*pgdao.ServiceEndpoint, error) {
	depEndpoint, err := pgdao.ServiceEndpoints(
		qm.Where("service_id = ?", serviceID),
		qm.And("code = ?", code),
	).One(context.Background(), exec)
	if err != nil {
		return nil, errors.DatabaseError("Failed to find endpoint by service id and code",
			errors.Details{"serviceId": serviceID, "code": code}, &err,
		).Logged(r.logger)
	}
	return depEndpoint, nil
}

func (r *postgresRepository) upsertService(exec boil.ContextExecutor, service *pgdao.Service) error {
	err := service.Upsert(
		context.Background(),
		exec,
		true,
		[]string{"code"},
		boil.Whitelist("name"),
		boil.Infer(),
	)
	if err != nil {
		return errors.DatabaseError("Failed to upsert service",
			errors.Details{"serviceCode": service.Code}, &err,
		).Logged(r.logger)
	}
	return nil
}

func (r *postgresRepository) upsertEndpoint(exec boil.ContextExecutor, endpoint *pgdao.ServiceEndpoint) error {
	err := endpoint.Upsert(
		context.Background(),
		exec,
		true,
		[]string{"service_id", "code"},
		boil.Whitelist("name"),
		boil.Infer(),
	)
	if err != nil {
		return errors.DatabaseError("Failed to upsert service endpoint",
			errors.Details{"serviceID": endpoint.ServiceID, "code": endpoint.Code}, &err,
		).Logged(r.logger)
	}
	return nil
}

func (r *postgresRepository) upsertDependency(exec boil.ContextExecutor, dependency *pgdao.ServiceEndpointDependency) error {
	// We have nothing to update, so upsert won't work. Instead we do a get, and maybe insert
	previousDependency, err := pgdao.ServiceEndpointDependencies(
		qm.Where("service_endpoint_id = ?", dependency.ServiceEndpointID),
		qm.And("dependency_service_endpoint_id = ?", dependency.DependencyServiceEndpointID),
	).One(context.Background(), exec)
	if err == sql.ErrNoRows {
		// If it doesn't exist, insert it
		err = dependency.Insert(context.Background(), exec, boil.Infer())
		if err != nil {
			return errors.DatabaseError("Failed to insert service endpoint dependency",
				errors.Details{
					"serviceEndpointID":           dependency.ServiceEndpointID,
					"dependencyServiceEndpointID": dependency.DependencyServiceEndpointID,
				},
				&err,
			).Logged(r.logger)
		}
		return nil
	} else if err != nil {
		// If the get failed, return a failure
		return errors.DatabaseError("Failed to get service endpoint dependency",
			errors.Details{
				"serviceEndpointID":           dependency.ServiceEndpointID,
				"dependencyServiceEndpointID": dependency.DependencyServiceEndpointID,
			},
			&err,
		).Logged(r.logger)
	}
	// If found, nothing to do but update the passed in model
	dependency.ID = previousDependency.ID
	return nil
}
