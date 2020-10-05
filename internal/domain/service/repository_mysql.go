package service

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	mysqldao "github.com/yashap/crius/internal/db/mysql/dao"
	"github.com/yashap/crius/internal/errors"
	"go.uber.org/zap"
)

type mysqlRepository struct {
	db     *sqlx.DB
	logger *zap.SugaredLogger
}

func (r *mysqlRepository) Save(s *Service) error {
	// TODO: move transaction handling to top level, pass through in Context
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		msg := "Failed to begin transaction when saving service"
		r.logger.Errorw(msg, "err", err.Error(), "serviceCode", s.Code)
		return errors.DatabaseError(msg, &err)
	}
	serviceDAO := mysqldao.Service{Code: s.Code, Name: s.Name}
	err = r.upsertService(tx, &serviceDAO)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	s.ID = &serviceDAO.ID
	endpointIDs := make([]interface{}, len(s.Endpoints))
	for idx, endpoint := range s.Endpoints {
		endpointDAO := mysqldao.ServiceEndpoint{
			ServiceID: serviceDAO.ID,
			Code:      endpoint.Code,
			Name:      endpoint.Name,
		}
		err = r.upsertEndpoint(tx, &endpointDAO)
		endpoint.ID = &endpointDAO.ID
		endpointIDs[idx] = endpointDAO.ID
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		for depServiceCode, depEndpointCodes := range endpoint.Dependencies {
			depService, err := r.FindByCode(depServiceCode)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			for _, depEndpointCode := range depEndpointCodes {
				depEndpoint, err := r.findEndpointByServiceIDAndCode(tx, *depService.ID, depEndpointCode)
				if err != nil {
					_ = tx.Rollback()
					return err
				}
				dependencyDAO := mysqldao.ServiceEndpointDependency{
					ServiceEndpointID:           endpointDAO.ID,
					DependencyServiceEndpointID: depEndpoint.ID,
				}
				err = r.upsertDependency(tx, &dependencyDAO)
				if err != nil {
					_ = tx.Rollback()
					return err
				}
			}
		}
	}
	// We want to fully replace the service, so remove any endpoints that no longer exist
	_, err = mysqldao.ServiceEndpoints(
		qm.Where("service_id = ?", serviceDAO.ID),
		qm.AndNotIn("id not in ?", endpointIDs...),
	).DeleteAll(context.Background(), tx)
	if err != nil {
		msg := "Failed to delete endpoints by ids"
		r.logger.Errorw(msg, "err", err.Error(), "ids", endpointIDs)
		_ = tx.Rollback()
		return errors.DatabaseError(msg, &err)
	}
	_ = tx.Commit()
	// TODO: remove any dangling dependencies (that should no longer exist)
	return nil
}

func (r *mysqlRepository) FindByCode(code Code) (*Service, error) {
	serviceDAO, err := mysqldao.Services(
		qm.Load(qm.Rels(
			mysqldao.ServiceRels.ServiceEndpoints,
			mysqldao.ServiceEndpointRels.ServiceEndpointDependencies,
		)),
		qm.Where("code = ?", code),
	).One(context.Background(), r.db)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		msg := "Failed to find service by code"
		r.logger.Errorw(msg, "err", err.Error(), "code", code)
		return nil, errors.DatabaseError(msg, &err)
	}
	endpoints := make([]Endpoint, len(serviceDAO.R.ServiceEndpoints))
	for idx, endpointDAO := range serviceDAO.R.ServiceEndpoints {
		dependencies := make(map[Code][]EndpointCode)
		for _, dependencyDAO := range endpointDAO.R.ServiceEndpointDependencies {
			depEndpointDAO, err := mysqldao.ServiceEndpoints(
				qm.Where("id = ?", dependencyDAO.DependencyServiceEndpointID),
			).One(context.Background(), r.db)
			if err != nil {
				msg := "Failed to find service endpoint by id"
				r.logger.Errorw(msg, "err", err.Error(), "id", dependencyDAO.DependencyServiceEndpointID)
				return nil, errors.DatabaseError(msg, &err)
			}
			depServiceDAO, err := mysqldao.Services(
				qm.Where("id = ?", depEndpointDAO.ServiceID),
			).One(context.Background(), r.db)
			if err != nil {
				msg := "Failed to find service by id"
				r.logger.Errorw(msg, "err", err.Error(), "id", depEndpointDAO.ServiceID)
				return nil, errors.DatabaseError(msg, &err)
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

func (r *mysqlRepository) findEndpointByServiceIDAndCode(
	exec boil.ContextExecutor,
	serviceID int64,
	code EndpointCode,
) (*mysqldao.ServiceEndpoint, error) {
	depEndpoint, err := mysqldao.ServiceEndpoints(
		qm.Where("service_id = ?", serviceID),
		qm.And("code = ?", code),
	).One(context.Background(), exec)
	if err != nil {
		msg := "Failed to find endpoint by service id and code"
		r.logger.Errorw(msg,
			"err", err.Error(),
			"serviceId", serviceID,
			"code", code,
		)
		return nil, errors.DatabaseError(msg, &err)
	}
	return depEndpoint, nil
}

func (r *mysqlRepository) upsertService(exec boil.ContextExecutor, service *mysqldao.Service) error {
	err := service.Upsert(
		context.Background(),
		exec,
		boil.Whitelist("name"),
		boil.Infer(),
	)
	if err != nil {
		msg := "Failed to upsert service"
		r.logger.Errorw(msg, "err", err.Error(), "serviceCode", service.Code)
		return errors.DatabaseError(msg, &err)
	}
	return nil
}

func (r *mysqlRepository) upsertEndpoint(exec boil.ContextExecutor, endpoint *mysqldao.ServiceEndpoint) error {
	// For MySQL, sqlboiler cannot upsert with a compound unique key, thus we do a get/insert-or-update workaround
	previousEndpoint, err := mysqldao.ServiceEndpoints(
		qm.Where("code = ?", endpoint.Code),
		qm.And("service_id = ?", endpoint.ServiceID),
	).One(context.Background(), exec)
	if err == sql.ErrNoRows {
		// If it doesn't exist, insert it
		err = endpoint.Insert(context.Background(), exec, boil.Infer())
		if err != nil {
			msg := "Failed to insert service endpoint"
			r.logger.Errorw(msg, "err", err.Error(), "serviceID", endpoint.ServiceID, "code", endpoint.Code)
			return errors.DatabaseError(msg, &err)
		}
		return nil
	} else if err != nil {
		// If the get failed, return a failure
		msg := "Failed to get service endpoint"
		r.logger.Errorw(msg, "err", err.Error(), "serviceID", endpoint.ServiceID, "code", endpoint.Code)
		return errors.DatabaseError(msg, &err)
	}
	// If found, update to the new endpoint
	endpoint.ID = previousEndpoint.ID
	_, err = endpoint.Update(context.Background(), exec, boil.Infer())
	if err != nil {
		msg := "Failed to update service endpoint"
		r.logger.Errorw(msg, "err", err.Error(), "serviceID", endpoint.ServiceID, "code", endpoint.Code)
		return errors.DatabaseError(msg, &err)
	}
	return nil
}

func (r *mysqlRepository) upsertDependency(exec boil.ContextExecutor, dependency *mysqldao.ServiceEndpointDependency) error {
	// For MySQL, sqlboiler cannot upsert with a compound unique key, thus we do a get/insert-or-update workaround
	previousDependency, err := mysqldao.ServiceEndpointDependencies(
		qm.Where("service_endpoint_id = ?", dependency.ServiceEndpointID),
		qm.And("dependency_service_endpoint_id = ?", dependency.DependencyServiceEndpointID),
	).One(context.Background(), exec)
	if err == sql.ErrNoRows {
		// If it doesn't exist, insert it
		err = dependency.Insert(context.Background(), exec, boil.Infer())
		if err != nil {
			msg := "Failed to insert service endpoint dependency"
			r.logger.Errorw(msg,
				"err", err.Error(),
				"serviceEndpointID", dependency.ServiceEndpointID,
				"dependencyServiceEndpointID", dependency.DependencyServiceEndpointID,
			)
			return errors.DatabaseError(msg, &err)
		}
		return nil
	} else if err != nil {
		// If the get failed, return a failure
		msg := "Failed to get service endpoint dependency"
		r.logger.Errorw(msg,
			"err", err.Error(),
			"serviceEndpointID", dependency.ServiceEndpointID,
			"dependencyServiceEndpointID", dependency.DependencyServiceEndpointID,
		)
		return errors.DatabaseError(msg, &err)
	}
	// If found, nothing to do but update the passed in model
	dependency.ID = previousDependency.ID
	return nil
}
