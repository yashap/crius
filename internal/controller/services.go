package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/yashap/crius/internal/domain/model"
	"github.com/yashap/crius/internal/domain/repository"
	"net/http"
)

type ServiceController struct {
	serviceRepository *repository.ServiceRepository
}

// NewServiceController creates a new ServiceController
func NewServiceController(serviceRepository *repository.ServiceRepository) ServiceController {
	return ServiceController{serviceRepository}
}

// Create creates a new model.Service
func (serviceController *ServiceController) Create(c *gin.Context) {
	var service model.Service
	err := c.ShouldBindJSON(&service)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": "Invalid input",
				"error":   err.Error(),
			},
		)
		return
	}

	id, err := serviceController.serviceRepository.Create(service)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Failed to save service",
				"error":   err.Error(),
			},
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// TODO: get

// TODO: delete
