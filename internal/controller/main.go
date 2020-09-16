package controller

import (
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/yashap/crius/internal/domain/service"
	"go.uber.org/zap"
)

func SetupRouter(
	serviceRepository *service.Repository,
	serviceFactory *service.Factory,
	logger *zap.SugaredLogger,
) *gin.Engine {
	serviceController := NewService(serviceRepository, serviceFactory, logger)

	// Run the server
	r := gin.New()
	r.Use(ginzap.Ginzap(logger.Desugar(), time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger.Desugar(), true))
	r.POST("/services", serviceController.Create)
	r.GET("/service/:code", serviceController.GetByCode)
	// TODO r.GET
	// TODO r.DELETE

	return r
}
