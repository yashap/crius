package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/xo/dburl"
	"github.com/yashap/crius/internal/dao"
	mysqldao "github.com/yashap/crius/internal/db/mysql/dao"
	"github.com/yashap/crius/internal/errors"
	"github.com/yashap/crius/internal/util"
	"go.uber.org/zap"
)

// Repository is a Service repository. It is a classic "Domain Driven Design" repository - the mental model is that
// it represents a collection of models.Service instances
type Repository interface {
	// Save saves a Service
	Save(s *Service) error
	// FindByCode finds a Service by its Code
	FindByCode(code Code) (*Service, error)
}

type mysqlRepository struct {
	dbURL  *dburl.URL
	db     *sqlx.DB
	logger *zap.SugaredLogger
}

func NewRepository2(
	dbURL *dburl.URL,
	db *sqlx.DB,
	logger *zap.SugaredLogger,
) Repository {
	// TODO: or return postgres repo
	return &mysqlRepository{
		db:     db,
		logger: logger,
	}
}

func (r *mysqlRepository) Save(s *Service) error {
	// TODO: move transaction handling to top level, pass through in Context
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		msg := "Failed to begin transaction when saving service"
		r.logger.Errorw(msg, "err", err.Error(), "serviceCode", s.Code)
		return errors.DatabaseError(msg, &err)
	}
	err = r.upsertService(tx, s)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	endpointIDs := make([]interface{}, len(s.Endpoints))
	//dependencyEndpointIDs := make([]interface{}, 0)
	for idx, endpoint := range s.Endpoints {
		err = r.upsertEndpoint(&endpoint, *s.ID)
		endpointIDs[idx] = *endpoint.ID
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		// TODO: remove any dangling endpoints, and also dangling deps
		for depServiceCode, depEndpointCodes := range endpoint.Dependencies {
			depService, err := r.FindByCode(depServiceCode)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			for _, depEndpointCode := range depEndpointCodes {
				depEndpoint, err := r.findEndpointByServiceIDAndCode(*depService.ID, depEndpointCode)
				if err != nil {
					_ = tx.Rollback()
					return err
				}
				err = r.upsertDependency(*endpoint.ID, depEndpoint.ID)
				if err != nil {
					_ = tx.Rollback()
					return err
				}
			}
		}
	}
	// We want to fully replace the service, so remove any endpoints that no longer exist
	_, err = mysqldao.ServiceEndpoints(
		qm.Where("service_id = ?", *s.ID),
		qm.AndNotIn("id not in ?", endpointIDs...),
	).DeleteAll(context.Background(), r.db)
	if err != nil {
		msg := "Failed to delete endpoints by ids"
		r.logger.Errorw(msg, "err", err.Error(), "ids", endpointIDs)
		_ = tx.Rollback()
		return errors.DatabaseError(msg, &err)
	}
	_ = tx.Commit()
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

func (r *mysqlRepository) findEndpointByServiceIDAndCode(serviceID int64, code EndpointCode) (*mysqldao.ServiceEndpoint, error) {
	depEndpoint, err := mysqldao.ServiceEndpoints(
		qm.Where("service_id = ?", serviceID),
		qm.And("code = ?", code),
	).One(context.Background(), r.db)
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

func (r *mysqlRepository) upsertService(exec boil.ContextExecutor, service *Service) error {
	serviceDAO := mysqldao.Service{
		Code: service.Code,
		Name: service.Name,
	}
	err := serviceDAO.Upsert(
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
	service.ID = &serviceDAO.ID
	return nil
}

func (r *mysqlRepository) upsertEndpoint(endpoint *Endpoint, serviceID int64) error {
	// For MySQL, sqlboiler cannot upsert with a compound unique key, thus we do a get/insert-or-update workaround
	endpointDAO, err := mysqldao.ServiceEndpoints(
		qm.Where("code = ?", endpoint.Code),
		qm.And("service_id = ?", serviceID),
	).One(context.Background(), r.db)
	if err == sql.ErrNoRows {
		// If it doesn't exist, insert it
		endpointDAO = &mysqldao.ServiceEndpoint{
			ServiceID: serviceID,
			Code:      endpoint.Code,
			Name:      endpoint.Name,
		}
		err = endpointDAO.Insert(context.Background(), r.db, boil.Infer())
		if err != nil {
			msg := "Failed to insert service endpoint"
			r.logger.Errorw(msg, "err", err.Error(), "serviceID", serviceID, "serviceEndpoint", endpoint)
			return errors.DatabaseError(msg, &err)
		}
	} else if err != nil {
		// If the get failed, return a failure
		msg := "Failed to get service endpoint"
		r.logger.Errorw(msg, "err", err.Error(), "serviceID", serviceID, "serviceEndpoint", endpoint)
		return errors.DatabaseError(msg, &err)
	} else {
		// If found, update
		endpointDAO.Name = endpoint.Name
		_, err = endpointDAO.Update(context.Background(), r.db, boil.Infer())
		if err != nil {
			msg := "Failed to update service endpoint"
			r.logger.Errorw(msg, "err", err.Error(), "serviceID", serviceID, "serviceEndpoint", endpoint)
			return errors.DatabaseError(msg, &err)
		}
	}
	endpoint.ID = &endpointDAO.ID
	return nil
}

func (r *mysqlRepository) upsertDependency(serviceEndpointID int64, dependencyServiceEndpointID int64) error {
	// For MySQL, sqlboiler cannot upsert with a compound unique key, thus we do a get/insert-or-update workaround
	dependencyDAO, err := mysqldao.ServiceEndpointDependencies(
		qm.Where("service_endpoint_id = ?", serviceEndpointID),
		qm.And("dependency_service_endpoint_id = ?", dependencyServiceEndpointID),
	).One(context.Background(), r.db)
	if err == sql.ErrNoRows {
		// If it doesn't exist, insert it
		dependencyDAO = &mysqldao.ServiceEndpointDependency{
			ServiceEndpointID:           serviceEndpointID,
			DependencyServiceEndpointID: dependencyServiceEndpointID,
		}
		err = dependencyDAO.Insert(context.Background(), r.db, boil.Infer())
		if err != nil {
			msg := "Failed to insert service endpoint dependency"
			r.logger.Errorw(msg,
				"err", err.Error(),
				"serviceEndpointID", serviceEndpointID,
				"dependencyServiceEndpointID", dependencyServiceEndpointID,
			)
			return errors.DatabaseError(msg, &err)
		}
	} else if err != nil {
		// If the get failed, return a failure
		msg := "Failed to get service endpoint dependency"
		r.logger.Errorw(msg,
			"err", err.Error(),
			"serviceEndpointID", serviceEndpointID,
			"dependencyServiceEndpointID", dependencyServiceEndpointID,
		)
		return errors.DatabaseError(msg, &err)
	}
	// If found, nothing to do - there's no data we need to update, we just need to ensure it exists
	return nil
}

// TODO: possibly refactor to sqlboiler?
type postgresRepository struct {
	serviceQueries                   dao.ServiceQueries
	serviceEndpointQueries           dao.ServiceEndpointQueries
	serviceEndpointDependencyQueries dao.ServiceEndpointDependencyQueries
	logger                           *zap.SugaredLogger
}

// NewRepository constructs a new Service repository
func NewRepository(
	serviceQueries dao.ServiceQueries,
	serviceEndpointQueries dao.ServiceEndpointQueries,
	serviceEndpointDependencyQueries dao.ServiceEndpointDependencyQueries,
	logger *zap.SugaredLogger,
) Repository {
	return &postgresRepository{
		serviceQueries:                   serviceQueries,
		serviceEndpointQueries:           serviceEndpointQueries,
		serviceEndpointDependencyQueries: serviceEndpointDependencyQueries,
		logger:                           logger,
	}
}

// Save saves a Service
func (r *postgresRepository) Save(s *Service) error {
	serviceDAO := s.toDAO()
	serviceID, err := r.serviceQueries.Upsert(serviceDAO)
	if err != nil {
		return err
	}
	for _, endpoint := range s.Endpoints {
		endpointDAO := endpoint.toDAO(serviceID)
		serviceEndpointId, err := r.serviceEndpointQueries.Upsert(endpointDAO)
		if err != nil {
			return err
		}
		for depServiceCode, depEndpointCodes := range endpoint.Dependencies {
			depService, err := r.FindByCode(depServiceCode)
			if err != nil {
				return err
			}
			if depService == nil {
				msg := "Tried to depend on an service code that does not exist"
				r.logger.Errorw(msg,
					"dependencyServiceCode", depServiceCode,
					"serviceCode", s.Code,
				)
				return errors.ServiceNotFound(
					fmt.Sprintf(
						"%s. Dependency Service Code: %s, Service Code: %s",
						msg, depServiceCode, s.Code,
					),
					nil,
				)
			}
			endpointCodesMap := make(map[EndpointCode]Endpoint)
			for _, depEndpoint := range depService.Endpoints {
				endpointCodesMap[depEndpoint.Code] = depEndpoint
			}
			for _, depEndpointCode := range depEndpointCodes {
				depEndpoint, ok := endpointCodesMap[depEndpointCode]
				if !ok {
					msg := "Tried to depend on an endpoint code that does not exist"
					r.logger.Errorw(msg,
						"dependencyEndpointCode", depEndpointCode,
						"dependencyServiceCode", depService.Code,
						"serviceCode", s.Code,
					)
					return errors.EndpointNotFound(
						fmt.Sprintf(
							"%s. Dependency Endpoint Code: %s, Dependency Service Code: %s, Service Code: %s",
							msg, depEndpointCode, depService.Code, s.Code,
						),
						nil,
					)
				}
				dependencyDAO := dao.ServiceEndpointDependency{
					ServiceEndpointID:           serviceEndpointId,
					DependencyServiceEndpointID: *depEndpoint.ID,
				}
				_, err := r.serviceEndpointDependencyQueries.Upsert(dependencyDAO)
				if err != nil {
					return err
				}
			}
		}
	}
	// TODO: cleanup everything we DIDN'T insert

	s.ID = &serviceID
	return nil
}

// FindByCode finds a Service by its Code
func (r *postgresRepository) FindByCode(code Code) (*Service, error) {
	svcDAO, err := r.serviceQueries.GetByCode(code)
	if err != nil {
		return nil, err
	}
	if svcDAO == nil {
		return nil, nil
	}
	endpointDAOs, err := r.serviceEndpointQueries.FindByServiceID(svcDAO.ID)
	if err != nil {
		return nil, err
	}
	endpoints := make([]Endpoint, len(endpointDAOs))
	for idx, endpointDAO := range endpointDAOs {
		endpoint := Endpoint{
			ID:           &endpointDAO.ID,
			Code:         endpointDAO.Code,
			Name:         endpointDAO.Name,
			Dependencies: nil, // added right after
		}
		err = r.addDependenciesToEndpoint(&endpoint)
		if err != nil {
			return nil, err
		}
		endpoints[idx] = endpoint
	}
	service := MakeService(
		&svcDAO.ID,
		svcDAO.Code,
		svcDAO.Name,
		endpoints,
	)
	return &service, nil
}

func (r *postgresRepository) addDependenciesToEndpoint(endpoint *Endpoint) error {
	if endpoint.ID == nil {
		return errors.UnclassifiedError(
			fmt.Sprintf("Expected endpoint to have id, but it didn't. Endpoint code: %s", endpoint.Code),
			nil,
		)
	}

	// Find every endpoint id we depend on
	dependencyDAOs, err := r.serviceEndpointDependencyQueries.FindByServiceEndpointID(*endpoint.ID)
	if err != nil {
		return err
	}
	everyEndpointIdThisEndpointDependsOn := make([]int64, len(dependencyDAOs))
	for idx, endpointDepDAO := range dependencyDAOs {
		everyEndpointIdThisEndpointDependsOn[idx] = endpointDepDAO.DependencyServiceEndpointID
	}
	everyEndpointIdThisEndpointDependsOn = util.Unique(everyEndpointIdThisEndpointDependsOn)

	// For every endpoint id we depend on, find the endpoints themselves, to get endpoint codes
	everyEndpointThisEndpointDependsOn, err := r.serviceEndpointQueries.FindByIDs(everyEndpointIdThisEndpointDependsOn)
	if err != nil {
		return err
	}

	// For all the endpoints we depend on, find the unique services that those endpoints belong to
	everyServiceIdThisEndpointDependsOn := make([]int64, 0)
	for _, depEndpoint := range everyEndpointThisEndpointDependsOn {
		everyServiceIdThisEndpointDependsOn = append(everyServiceIdThisEndpointDependsOn, depEndpoint.ServiceID)
	}
	everyServiceIdThisEndpointDependsOn = util.Unique(everyServiceIdThisEndpointDependsOn)
	everyServiceThisEndpointDependsOn, err := r.serviceQueries.FindByIDs(everyServiceIdThisEndpointDependsOn)
	if err != nil {
		return err
	}

	// Using the collected data, construct this endpoint's dependencies
	dependencies := make(map[Code][]EndpointCode)
	for _, depSvc := range everyServiceThisEndpointDependsOn {
		depSvcEndpoints := make([]EndpointCode, 0)
		for _, endpointDependency := range everyEndpointThisEndpointDependsOn {
			depSvcEndpoints = append(depSvcEndpoints, endpointDependency.Code)
		}
		dependencies[depSvc.Code] = depSvcEndpoints
	}
	endpoint.Dependencies = dependencies
	return nil
}
