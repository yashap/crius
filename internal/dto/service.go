package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/yashap/crius/internal/domain/service"
	"github.com/yashap/crius/internal/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ServiceCode = string
type ServiceName = string
type EndpointCode = string
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

func (s *Service) ToEntity(db *gorm.DB, logger *zap.SugaredLogger) service.Service {
	var endpoints []service.Endpoint
	if s.Endpoints == nil {
		endpoints = make([]service.Endpoint, 0)
	} else {
		endpoints = endpointsToEntities(*s.Endpoints)
	}
	return service.MakeService(db, logger, nil, *s.Code, *s.Name, endpoints)
}

func MakeServiceFromRequest(c *gin.Context) (Service, error) {
	var s Service
	err := c.ShouldBindJSON(&s)
	if err != nil {
		return s, errors.InvalidInput("failed to unmarshall json to Service", &err)
	}
	err = s.validate()
	return s, err
}

func MakeServiceFromEntity(s service.Service) Service {
	endpointDTOs := makeEndpointsFromEntities(s.Endpoints)
	return Service{
		Code:      &s.Code,
		Name:      &s.Name,
		Endpoints: &endpointDTOs,
	}
}

// TODO: finish this, then get rid of service factory
func endpointsToEntities(endpoints []Endpoint) []service.Endpoint {
	var dependencies map[ServiceCode][]EndpointCode
	if d.Dependencies != nil {
		dtoDependencies := *d.Dependencies
		dependencies = make(map[ServiceCode][]EndpointCode, len(dtoDependencies))
		for serviceCode, dtoEndpointCodes := range dtoDependencies {
			endpointCodes := make([]EndpointCode, len(dtoEndpointCodes))
			for idx, dtoEndpointCode := range dtoEndpointCodes {
				endpointCodes[idx] = dtoEndpointCode
			}
			if len(endpointCodes) > 0 {
				dependencies[serviceCode] = endpointCodes
			}
		}
	}

	e := Endpoint{
		Code:         *d.Code,
		Name:         *d.Name,
		Dependencies: dependencies,
	}
	return e
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
