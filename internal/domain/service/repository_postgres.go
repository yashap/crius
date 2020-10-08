package service

import (
	"context"
	"database/sql"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/yashap/crius/internal/db"
	dao "github.com/yashap/crius/internal/db/postgresql/dao"
	"github.com/yashap/crius/internal/errors"
	"go.uber.org/zap"
)

type postgresRepository struct {
	database db.Database
	logger   *zap.SugaredLogger
}

func (r *postgresRepository) Save(ctx context.Context, s *Service) error {
	exec := r.database.GetExecutor(ctx)
	serviceDAO := dao.Service{Code: s.Code, Name: s.Name}
	err := r.upsertService(ctx, &serviceDAO)
	if err != nil {
		return err
	}
	s.ID = &serviceDAO.ID
	endpointIDs := make([]interface{}, len(s.Endpoints))
	for idx, endpoint := range s.Endpoints {
		endpointDAO := dao.ServiceEndpoint{
			ServiceID: serviceDAO.ID,
			Code:      endpoint.Code,
			Name:      endpoint.Name,
		}
		err = r.upsertEndpoint(ctx, &endpointDAO)
		endpoint.ID = &endpointDAO.ID
		endpointIDs[idx] = endpointDAO.ID
		if err != nil {
			return err
		}
		for depServiceCode, depEndpointCodes := range endpoint.ServiceEndpointDependencies {
			depService, err := r.FindByCode(ctx, depServiceCode)
			if err != nil {
				return err
			}
			for _, depEndpointCode := range depEndpointCodes {
				depEndpoint, err := r.findEndpointByServiceIDAndCode(ctx, *depService.ID, depEndpointCode)
				if err != nil {
					return err
				}
				dependencyDAO := dao.ServiceEndpointDependency{
					ServiceEndpointID:           endpointDAO.ID,
					DependencyServiceEndpointID: depEndpoint.ID,
				}
				err = r.upsertDependency(ctx, &dependencyDAO)
				if err != nil {
					return err
				}
			}
		}
	}
	// We want to fully replace the service, so remove any endpoints that no longer exist
	_, err = dao.ServiceEndpoints(
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
	serviceDAO, err := dao.Services(
		qm.Load(qm.Rels(
			dao.ServiceRels.ServiceEndpoints,
			dao.ServiceEndpointRels.ServiceEndpointDependencies,
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
	service, err := r.serviceDAOToEntity(ctx, *serviceDAO)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *postgresRepository) FindAll(ctx context.Context) ([]Service, error) {
	exec := r.database.GetExecutor(ctx)
	serviceDAOs, err := dao.Services(
		qm.Load(qm.Rels(
			dao.ServiceRels.ServiceEndpoints,
			dao.ServiceEndpointRels.ServiceEndpointDependencies,
		)),
	).All(ctx, exec)
	if err != nil {
		return nil, errors.DatabaseError("Failed to find all services", errors.Details{}, &err).Logged(r.logger)
	}
	services := make([]Service, len(serviceDAOs))
	// TODO make this more efficient (bulk fetch, provide data?)
	for idx, serviceDAO := range serviceDAOs {
		service, err := r.serviceDAOToEntity(ctx, *serviceDAO)
		if err != nil {
			return nil, err
		}
		services[idx] = service
	}
	return services, nil
}

func (r *postgresRepository) findEndpointByServiceIDAndCode(
	ctx context.Context,
	serviceID int64,
	code EndpointCode,
) (*dao.ServiceEndpoint, error) {
	exec := r.database.GetExecutor(ctx)
	depEndpoint, err := dao.ServiceEndpoints(
		qm.Where("service_id = ?", serviceID),
		qm.And("code = ?", code),
	).One(ctx, exec)
	if err != nil {
		return nil, errors.DatabaseError("Failed to find endpoint by service id and code",
			errors.Details{"serviceId": serviceID, "code": code}, &err,
		).Logged(r.logger)
	}
	return depEndpoint, nil
}

func (r *postgresRepository) upsertService(ctx context.Context, service *dao.Service) error {
	exec := r.database.GetExecutor(ctx)
	err := service.Upsert(
		ctx,
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

func (r *postgresRepository) upsertEndpoint(ctx context.Context, endpoint *dao.ServiceEndpoint) error {
	exec := r.database.GetExecutor(ctx)
	err := endpoint.Upsert(
		ctx,
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

func (r *postgresRepository) upsertDependency(ctx context.Context, dependency *dao.ServiceEndpointDependency) error {
	// We have nothing to update, so upsert won't work. Instead we do a get, and maybe insert
	exec := r.database.GetExecutor(ctx)
	previousDependency, err := dao.ServiceEndpointDependencies(
		qm.Where("service_endpoint_id = ?", dependency.ServiceEndpointID),
		qm.And("dependency_service_endpoint_id = ?", dependency.DependencyServiceEndpointID),
	).One(ctx, exec)
	if err == sql.ErrNoRows {
		// If it doesn't exist, insert it
		err = dependency.Insert(ctx, exec, boil.Infer())
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

func (r *postgresRepository) serviceDAOToEntity(ctx context.Context, serviceDAO dao.Service) (Service, error) {
	exec := r.database.GetExecutor(ctx)
	endpoints := make([]Endpoint, len(serviceDAO.R.ServiceEndpoints))
	for idx, endpointDAO := range serviceDAO.R.ServiceEndpoints {
		dependencies := make(map[Code][]EndpointCode)
		for _, dependencyDAO := range endpointDAO.R.ServiceEndpointDependencies {
			depEndpointDAO, err := dao.ServiceEndpoints(
				qm.Where("id = ?", dependencyDAO.DependencyServiceEndpointID),
			).One(ctx, exec)
			if err != nil {
				return Service{}, errors.DatabaseError("Failed to find service endpoint by id",
					errors.Details{"id": dependencyDAO.DependencyServiceEndpointID}, &err,
				).Logged(r.logger)
			}
			depServiceDAO, err := dao.Services(
				qm.Where("id = ?", depEndpointDAO.ServiceID),
			).One(ctx, exec)
			if err != nil {
				return Service{}, errors.DatabaseError("Failed to find service by id",
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
			Code:                        endpointDAO.Code,
			Name:                        endpointDAO.Name,
			ServiceEndpointDependencies: dependencies,
		}
	}
	service := Service{
		ID:        &serviceDAO.ID,
		Code:      serviceDAO.Code,
		Name:      serviceDAO.Name,
		Endpoints: endpoints,
	}
	return service, nil
}
