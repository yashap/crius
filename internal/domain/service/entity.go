package service

import (
	"fmt"

	"github.com/yashap/crius/internal/dao"

	"go.uber.org/zap"
)

// Code is a code that uniquely identifies a Service
type Code = string

// Name is the human-readable/friedly name of a Service
type Name = string

// EndpointCode is a code that uniquely identifies an Endpoint. It need not be globally unique, only unique within that one Service
type EndpointCode = string

// EndpointName is the human-readable/friedly name of an Endpoint
type EndpointName = string

// Service represents a service
type Service struct {
	serviceQueries         dao.ServiceQueries
	serviceEndpointQueries dao.ServiceEndpointQueries
	logger                 *zap.SugaredLogger
	// ID uniquely identifies this service
	ID *int64
	// Code is a unique code for the service. For example, "location_tracking" for a location tracking service
	Code Code
	// Name is a friendly name for the service. For example, "Location Tracking" for a location tracking service
	Name Name
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
	Dependencies map[Code][]EndpointCode
}

// MakeService constructs a Service
func MakeService(
	serviceQueries dao.ServiceQueries,
	serviceEndpointQueries dao.ServiceEndpointQueries,
	logger *zap.SugaredLogger,
	id *int64,
	code Code,
	name Name,
	endpoints []Endpoint,
) Service {
	return Service{
		serviceQueries:         serviceQueries,
		serviceEndpointQueries: serviceEndpointQueries,
		logger:                 logger,
		ID:                     id,
		Code:                   code,
		Name:                   name,
		Endpoints:              endpoints,
	}
}

// Save saves a Service
func (s *Service) Save() error {
	serviceDAO := s.toDAO()
	serviceID, err := s.serviceQueries.Upsert(serviceDAO)
	if err != nil {
		return err
	}
	fmt.Printf(">>>>>> 1 %v\n\n", s)
	for _, endpoint := range s.Endpoints {
		fmt.Printf(">>>>>> 2\n\n")
		endpointDAO := endpoint.toDAO(serviceID)
		fmt.Printf(">>>>>> 3\n\n")
		// TODO: for some reason, within the service, serviceEndpointQueries is nil ====> figure out why
		_, err := s.serviceEndpointQueries.Upsert(endpointDAO)
		if err != nil {
			return err
		}
		// TODO save endpoint deps
	}

	fmt.Printf(">>>>>> done %v\n\n", serviceID)
	s.ID = &serviceID
	return nil
}

func (s *Service) toDAO() dao.Service {
	return dao.Service{
		Code: s.Code,
		Name: s.Name,
	}
}

func (e *Endpoint) toDAO(serviceID int64) dao.ServiceEndpoint {
	return dao.ServiceEndpoint{
		ServiceID: serviceID,
		Code:      e.Code,
		Name:      e.Name,
	}
}

// func newEndpointDAOs(db *sqlx.DB, endpoints []Endpoint) ([]dao.ServiceEndpoint, error) {
// 	endpointDAOs := make([]dao.ServiceEndpoint, len(endpoints))
// 	for idx, e := range endpoints {
// 		deps, err := newDependencyDAOs(db, e.Dependencies)
// 		if err != nil {
// 			return endpointDAOs, err
// 		}
// 		endpointDAOs[idx] = dao.ServiceEndpoint{
// 			Code:                        e.Code,
// 			Name:                        e.Name,
// 			ServiceEndpointDependencies: deps,
// 		}
// 	}
// 	return endpointDAOs, nil
// }

// func newDependencyDAOs(
// 	db *sqlx.DB,
// 	dependencies map[Code][]EndpointCode,
// ) ([]dao.ServiceEndpointDependency, error) {
// 	dependencyDAOs := make([]dao.ServiceEndpointDependency, len(dependencies)) // TODO: smarter?
// 	for depServiceCode, depEndpointCodes := range dependencies {
// 		for _, depEndpointCode := range depEndpointCodes {
// 			// TODO: join
// 			depService := dao.Service{}
// 			depEndpoint := dao.ServiceEndpoint{}
// 			if err := db.Where(dao.Service{Code: depServiceCode}).First(&depService).Error; err != nil {
// 				return dependencyDAOs, err
// 			}
// 			if err := db.Where(dao.ServiceEndpoint{ServiceID: depService.ID, Code: depEndpointCode}).First(&depEndpoint).Error; err != nil {
// 				return dependencyDAOs, err
// 			}
// 			dependencyDAOs = append(dependencyDAOs, dao.ServiceEndpointDependency{DependencyServiceEndpointID: depEndpoint.ID})
// 		}
// 	}
// 	return dependencyDAOs, nil
// }
