package service

import (
	"github.com/yashap/crius/internal/dao"
	"gorm.io/gorm/clause"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ServiceCode = string
type ServiceName = string
type EndpointCode = string
type EndpointName = string

// Service represents a service
type Service struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
	// ID uniquely identifies this service
	ID *int64
	// Code is a unique code for the service. For example, "location_tracking" for a location tracking service
	Code ServiceCode
	// Name is a friendly name for the service. For example, "Location Tracking" for a location tracking service
	Name ServiceName
	// Endpoints is a list of Endpoints that the Service has
	Endpoints []Endpoint
}

// Endpoint represents an Endpoint of a Service
type Endpoint struct {
	// Code is a unique code for the Endpoint. Doesn't have to be globally unique, just unique per service. For example,
	// "POST /locations" or "GET /locations/:id"
	Code EndpointCode
	// Name is a friendly name for the Endpoint. For example, "Create location" or "Get location by id"
	Name EndpointName
	// Dependencies is a map of Dependencies for a given Endpoint. Keys are service codes, values are lists of endpoint codes
	Dependencies map[ServiceCode][]EndpointCode
}

func MakeService(
	db *gorm.DB,
	logger *zap.SugaredLogger,
	id *int64,
	code ServiceCode,
	name ServiceName,
	endpoints []Endpoint,
) Service {
	return Service{
		db:        db,
		logger:    logger,
		ID:        id,
		Code:      code,
		Name:      name,
		Endpoints: endpoints,
	}
}

// Save saves a Service
func (s *Service) Save() error {
	// TODO: log error, convert error, context
	serviceDAO, err := s.toDAO()
	if err != nil {
		return err
	}
	err = s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{"code", "name"}),
	}).Create(serviceDAO).Error
	if err != nil {
		return err
	}
	s.ID = &serviceDAO.ID
	return nil
}

func (s *Service) toDAO() (*dao.Service, error) {
	endpoints, err := newEndpointDAOs(s.db, s.Endpoints)
	if err != nil {
		return nil, err
	}
	service := dao.Service{
		Code:             s.Code,
		Name:             s.Name,
		ServiceEndpoints: endpoints,
	}
	return &service, nil
}

func newEndpointDAOs(db *gorm.DB, endpoints []Endpoint) ([]dao.ServiceEndpoint, error) {
	endpointDAOs := make([]dao.ServiceEndpoint, len(endpoints))
	for idx, e := range endpoints {
		deps, err := newDependencyDAOs(db, e.Dependencies)
		if err != nil {
			return endpointDAOs, err
		}
		endpointDAOs[idx] = dao.ServiceEndpoint{
			Code:                        e.Code,
			Name:                        e.Name,
			ServiceEndpointDependencies: deps,
		}
	}
	return endpointDAOs, nil
}

func newDependencyDAOs(
	db *gorm.DB,
	dependencies map[ServiceCode][]EndpointCode,
) ([]dao.ServiceEndpointDependency, error) {
	dependencyDAOs := make([]dao.ServiceEndpointDependency, len(dependencies)) // TODO: smarter?
	for depServiceCode, depEndpointCodes := range dependencies {
		for _, depEndpointCode := range depEndpointCodes {
			// TODO: join
			depService := dao.Service{}
			depEndpoint := dao.ServiceEndpoint{}
			if err := db.Where(dao.Service{Code: depServiceCode}).First(&depService).Error; err != nil {
				return dependencyDAOs, err
			}
			if err := db.Where(dao.ServiceEndpoint{ServiceID: depService.ID, Code: depEndpointCode}).First(&depEndpoint).Error; err != nil {
				return dependencyDAOs, err
			}
			dependencyDAOs = append(dependencyDAOs, dao.ServiceEndpointDependency{DependencyServiceEndpointID: depEndpoint.ID})
		}
	}
	return dependencyDAOs, nil
}
