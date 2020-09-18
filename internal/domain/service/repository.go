package service

import (
	"github.com/yashap/crius/internal/dao"
	"go.uber.org/zap"
)

// Repository is a Service repository. It is a classic "Domain Driven Design" repository - the mental model is that
// it represets a collection of models.Service instances
type Repository struct {
	serviceQueries         dao.ServiceQueries
	serviceEndpointQueries dao.ServiceEndpointQueries
	logger                 *zap.SugaredLogger
}

// NewRepository constructs a new Service repository
func NewRepository(
	serviceQueries dao.ServiceQueries,
	serviceEndpointQueries dao.ServiceEndpointQueries,
	logger *zap.SugaredLogger,
) Repository {
	return Repository{
		serviceQueries: serviceQueries,
		logger:         logger,
	}
}

// FindByCode finds a Service by its Code
func (r *Repository) FindByCode(code Code) (*Service, error) {
	svcDAO, err := r.serviceQueries.GetByCode(code)
	if err != nil {
		return nil, err
	}
	service := MakeService(
		r.serviceQueries,
		r.serviceEndpointQueries,
		r.logger,
		&svcDAO.ID,
		svcDAO.Code,
		svcDAO.Name,
		[]Endpoint{}, // TODO more complete
	)
	return &service, nil
}
