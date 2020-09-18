package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/yashap/crius/internal/dao"
	"github.com/yashap/crius/internal/domain/service"
	"github.com/yashap/crius/internal/errors"
	"go.uber.org/zap"
)

// ServiceCode is a code that uniquely identifies a Service
type ServiceCode = string

// ServiceName is the human-readable/friedly name of a Service
type ServiceName = string

// EndpointCode is a code that uniquely identifies an Endpoint. It need not be globally unique, only unique within that one Service
type EndpointCode = string

// EndpointName is the human-readable/friedly name of an Endpoint
type EndpointName = string

// Service represents a service
type Service struct {
	// Code is a unique code for the service. For example, "location_tracking" for a location tracking service
	Code *ServiceCode `json:"code"`
	// Name is a friendly name for the service. For example, "Location Tracking" for a location tracking service
	Name *ServiceName `json:"name"`
	// Endpoints is a list of Endpoints that the Service has
	Endpoints *[]Endpoint `json:"endpoints"`
}

// Endpoint represents an Endpoint of a Service
type Endpoint struct {
	// Code is a unique code for the Endpoint. Doesn't have to be globally unique, just unique per service. For example,
	// "POST /locations" or "GET /locations/:id"
	Code *EndpointCode `json:"code"`
	// Name is a friendly name for the Endpoint. For example, "Create location" or "Get location by id"
	Name *EndpointName `json:"name"`
	// Dependencies is a map of Dependencies for a given Endpoint. Keys are service codes, values are lists of endpoint codes
	Dependencies *map[ServiceCode][]EndpointCode `json:"dependencies"`
}

// ToEntity converts a Service DTO into a Service Entity
func (s *Service) ToEntity(
	serviceQueries dao.ServiceQueries,
	serviceEndpointQueries dao.ServiceEndpointQueries,
	logger *zap.SugaredLogger,
) service.Service {
	var endpoints []service.Endpoint
	if s.Endpoints == nil {
		endpoints = make([]service.Endpoint, 0)
	} else {
		endpoints = endpointsToEntities(*s.Endpoints)
	}
	return service.MakeService(serviceQueries, serviceEndpointQueries, logger, nil, *s.Code, *s.Name, endpoints)
}

// MakeServiceFromRequest constructs a Service DTO from an HTTP request
func MakeServiceFromRequest(c *gin.Context) (Service, error) {
	var s Service
	err := c.ShouldBindJSON(&s)
	if err != nil {
		return s, errors.InvalidInput("failed to unmarshall json to Service", &err)
	}
	err = s.validate()
	return s, err
}

// MakeServiceFromEntity constructs a Service DTO from a Service Entity
func MakeServiceFromEntity(s service.Service) Service {
	endpointDTOs := makeEndpointsFromEntities(s.Endpoints)
	return Service{
		Code:      &s.Code,
		Name:      &s.Name,
		Endpoints: &endpointDTOs,
	}
}

func endpointsToEntities(endpoints []Endpoint) []service.Endpoint {
	endpointEntities := make([]service.Endpoint, len(endpoints))
	for idx, endpoint := range endpoints {
		var dependencies map[ServiceCode][]EndpointCode
		if endpoint.Dependencies == nil {
			dependencies = make(map[ServiceCode][]EndpointCode)
		} else {
			dependencies = *endpoint.Dependencies
		}
		endpointEntities[idx] = service.Endpoint{
			Code:         *endpoint.Code,
			Name:         *endpoint.Name,
			Dependencies: dependencies,
		}
	}
	return endpointEntities
}

func makeEndpointsFromEntities(endpoints []service.Endpoint) []Endpoint {
	endpointDTOs := make([]Endpoint, len(endpoints))
	for idx, e := range endpoints {
		endpointDTOs[idx] = Endpoint{
			Code:         &e.Code,
			Name:         &e.Name,
			Dependencies: &e.Dependencies,
		}
	}
	return endpointDTOs
}

func (s Service) validate() error {
	if s.Code == nil {
		return errors.InvalidInput("field 'code' on object Service is required", nil)
	}
	if s.Name == nil {
		return errors.InvalidInput("field 'name' on object Service is required", nil)
	}
	for _, endpoint := range *s.Endpoints {
		err := endpoint.validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (e Endpoint) validate() error {
	if e.Code == nil {
		return errors.InvalidInput("field 'code' on object Endpoint is required", nil)
	}
	if e.Name == nil {
		return errors.InvalidInput("field 'name' on object Endpoint is required", nil)
	}
	return nil
}
