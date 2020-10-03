package controller

import (
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/yashap/crius/internal/domain/service"
	"go.uber.org/zap"
)

// SetupRouter sets up the Gin router
func SetupRouter(serviceRepository service.Repository, logger *zap.SugaredLogger) *gin.Engine {
	serviceController := NewService(serviceRepository)

	// Run the server
	r := gin.New()
	r.Use(ginzap.Ginzap(logger.Desugar(), time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger.Desugar(), true))
	r.POST("/services", serviceController.Create)
	r.GET("/services/:code", serviceController.GetByCode)
	// TODO r.GET
	// TODO r.DELETE

	return r
}
