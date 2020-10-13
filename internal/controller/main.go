package controller

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/yashap/crius/internal/db"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/yashap/crius/internal/domain/service"
	"go.uber.org/zap"
)

// SetupRouter sets up the Gin router
func SetupRouter(database db.Database, serviceRepository service.Repository, logger *zap.SugaredLogger) *gin.Engine {
	ctx := context.Background()
	serviceController := NewService(ctx, database, serviceRepository, logger)

	// Set up routes
	r := gin.New()
	r.Use(ginzap.Ginzap(logger.Desugar(), time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger.Desugar(), true))
	r.Use(cors.Default()) // TODO more restrictive
	r.POST("/services", serviceController.Create)
	r.GET("/services", serviceController.GetAll)
	r.GET("/services/:code", serviceController.GetByCode)
	// TODO r.GET
	// TODO r.DELETE

	return r
}
