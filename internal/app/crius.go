package app

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"

	_ "github.com/lib/pq" // Postgres driver
	"github.com/yashap/crius/internal/controller"
	"github.com/yashap/crius/internal/dao"
	"github.com/yashap/crius/internal/db"
	"github.com/yashap/crius/internal/domain/service"
	"go.uber.org/zap"
)

// Crius is the Crius application
type Crius interface {
	// MigrateDB runs the DB migrations
	MigrateDB(migrationDir string) Crius
	// ListenAndServe starts the HTTP server
	ListenAndServe() Crius

	// ServiceRepository returns the app's service.Repository
	ServiceRepository() *service.Repository
	// Router returns the app's Router
	Router() *gin.Engine
}

type crius struct {
	db                *sqlx.DB
	dbURL             *dburl.URL
	logger            *zap.SugaredLogger
	serviceRepository *service.Repository
	router            *gin.Engine
}

// NewCrius creates a new Crius application
func NewCrius(dbURL *dburl.URL) Crius {
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	database, err := sqlx.Connect(dbURL.Driver, dbURL.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database. URL: %s ; Error: %s", dbURL, err.Error())
	}
	serviceQueries, err := dao.NewServiceQueries(dbURL, database)
	if err != nil {
		log.Fatalf("Failed to create ServiceQueries. DB URL: %s, Error: %s", dbURL, err.Error())
	}
	serviceEndpointQueries, err := dao.NewServiceEndpointQueries(dbURL, database)
	if err != nil {
		log.Fatalf("Failed to create ServiceEndpointQueries. DB URL: %s, Error: %s", dbURL, err.Error())
	}
	serviceRepository := service.NewRepository(serviceQueries, serviceEndpointQueries, logger)
	router := controller.SetupRouter(serviceQueries, serviceEndpointQueries, &serviceRepository, logger)

	return &crius{
		db:                database,
		dbURL:             dbURL,
		logger:            logger,
		serviceRepository: &serviceRepository,
		router:            router,
	}
}

func (c *crius) MigrateDB(migrationDir string) Crius {
	db.Migrate(c.db, c.dbURL, migrationDir)
	return c
}

func (c *crius) ListenAndServe() Crius {
	err := c.router.Run()
	if err != nil {
		log.Fatal("Failed to run server: " + err.Error())
	}
	return c
}

func (c *crius) ServiceRepository() *service.Repository {
	return c.serviceRepository
}

func (c *crius) Router() *gin.Engine {
	return c.router
}
