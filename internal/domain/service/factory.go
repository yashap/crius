package service

import (
	"gorm.io/gorm"

	"github.com/yashap/crius/internal/dto"
	"go.uber.org/zap"
)

type Factory struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewFactory(db *gorm.DB, logger *zap.SugaredLogger) Factory {
	return Factory{
		db:     db,
		logger: logger,
	}
}

func (sf *Factory) NewService(d dto.Service) Service {
	endpoints := make([]Endpoint, len(*d.Endpoints))
	for idx, endpoint := range *d.Endpoints {
		endpoints[idx] = newEndpoint(endpoint)
	}
	s := Service{
		db:        sf.db,
		logger:    sf.logger,
		Code:      *d.Code,
		Name:      *d.Name,
		Endpoints: endpoints,
	}
	return s
}

func newEndpoint(d dto.Endpoint) Endpoint {
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
