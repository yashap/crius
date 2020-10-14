package app

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // Postgres driver
	"github.com/yashap/crius/internal/controller"
	"github.com/yashap/crius/internal/db"
	"github.com/yashap/crius/internal/domain/service"
	"github.com/yashap/crius/internal/errors"
	"go.uber.org/zap"
)

// Crius is the Crius application
type Crius interface {
	// ListenAndServe starts the HTTP server
	ListenAndServe()
	// Router returns the app's Router
	Router() *gin.Engine
}

type crius struct {
	database db.Database
	router   *gin.Engine
}

// NewCrius creates a new Crius application
func NewCrius(database db.Database) Crius {
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	serviceRepository := service.NewRepository(database, logger)
	router := controller.SetupRouter(database, serviceRepository, logger)

	return &crius{
		database: database,
		router:   router,
	}
}

func (c *crius) ListenAndServe() {
	err := c.router.Run()
	if err != nil {
		panic(errors.InitializationError("failed to run server", errors.Details{}, &err))
	}
}

func (c *crius) Router() *gin.Engine {
	return c.router
}
