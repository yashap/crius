package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
	db     *sqlx.DB
	logger *zap.SugaredLogger
}

func NewRepository2(
	db *sqlx.DB,
	logger *zap.SugaredLogger,
) Repository {
	return &mysqlRepository{
		db:     db,
		logger: logger,
	}
}

func (r *mysqlRepository) Save(s *Service) error {
	// TODO: pull out private upsertEndpoint method?
	serviceDAO := &mysqldao.Service{}
	serviceDAO.Code = s.Code
	serviceDAO.Name = s.Name
	err := serviceDAO.Upsert(
		context.Background(),
		r.db.DB,
		boil.Whitelist("name"),
		boil.Infer(),
	)
	if err != nil {
		msg := "Failed to upsert service"
		r.logger.Errorw(msg, "err", err.Error(), "service", s)
		return errors.DatabaseError(msg, &err)
	}
	for _, endpoint := range s.Endpoints {
		err := r.upsertEndpoint(&endpoint, serviceDAO.ID)
		if err != nil {
			return err
		}
		// TODO: upsert deps
	}
	s.ID = &serviceDAO.ID
	return nil
}

func (r *mysqlRepository) FindByCode(code Code) (*Service, error) {
	serviceDAO, err := mysqldao.Services(qm.Where("code = ?", code)).One(context.Background(), r.db)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		msg := "Failed to find service by code"
		r.logger.Errorw(msg, "err", err.Error(), "code", code)
		return nil, errors.DatabaseError(msg, &err)
	}
	service := Service{
		ID:        &serviceDAO.ID,
		Code:      serviceDAO.Code,
		Name:      serviceDAO.Name,
		Endpoints: nil, // TODO: populate
	}
	return &service, nil
}

func (r *mysqlRepository) upsertEndpoint(endpoint *Endpoint, serviceID int64) error {
	// sqlboiler cannot upsert on compound unique key, for MySQL, thus the get/insert/update workaround
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
