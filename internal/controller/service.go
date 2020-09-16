package controller

import (
	"fmt"
	"net/http"

	"github.com/yashap/crius/internal/errors"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/yashap/crius/internal/domain/service"
	"github.com/yashap/crius/internal/dto"
)

// Service is a controller for /service endpoints
type Service struct {
	serviceRepository *service.Repository
	serviceFactory    *service.Factory
	logger            *zap.SugaredLogger
}

// NewService instantiates a Service controller
func NewService(
	serviceRepository *service.Repository,
	serviceFactory *service.Factory,
	logger *zap.SugaredLogger,
) Service {
	return Service{
		serviceRepository: serviceRepository,
		serviceFactory:    serviceFactory,
		logger:            logger,
	}
}

// Create creates a new service.Service
// POST /services { ... service DTO ... }
func (sc *Service) Create(c *gin.Context) {
	svcDTO, err := dto.MakeServiceFromRequest(c)
	if err != nil {
		errors.SetResponse(err, c)
		return
	}
	s := sc.serviceFactory.NewService(svcDTO)
	err = s.Save()
	if err != nil {
		errors.SetResponse(err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": s.ID})
}

// GetByCode gets a service.Service by the service's code
// GET /services/:code { ... service DTO ... }
func (sc *Service) GetByCode(c *gin.Context) {
	code := c.Param("name")
	svc, err := sc.serviceRepository.FindByCode(code)
	if err != nil {
		errors.SetResponse(err, c)
		return
	}
	if svc == nil {
		errors.SetResponse(
			errors.ServiceNotFound(fmt.Sprintf("Service with code %s not found", code), nil),
			c,
		)
		return
	}
	c.JSON(http.StatusOK, dto.MakeServiceFromEntity(*svc))
}
