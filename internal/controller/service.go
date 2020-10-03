package controller

import (
	"fmt"
	"net/http"

	"github.com/yashap/crius/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/yashap/crius/internal/domain/service"
	"github.com/yashap/crius/internal/dto"
)

// Service is a controller for /service endpoints
type Service struct {
	serviceRepository service.Repository
}

// NewService instantiates a Service controller
func NewService(serviceRepository service.Repository) Service {
	return Service{serviceRepository}
}

// Create creates a new service.Service
// POST /services { ... service DTO ... }
func (sc *Service) Create(c *gin.Context) {
	serviceDTO, err := dto.MakeServiceFromRequest(c)
	if err != nil {
		errors.SetResponse(err, c)
		return
	}
	svc := serviceDTO.ToEntity()
	err = sc.serviceRepository.Save(&svc)
	if err != nil {
		errors.SetResponse(err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": svc.ID})
}

// GetByCode gets a service.Service by the service's code
// GET /services/:code { ... service DTO ... }
func (sc *Service) GetByCode(c *gin.Context) {
	code := c.Param("code")
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
