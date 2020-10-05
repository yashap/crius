package app

import (
	"fmt"
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
	serviceQueries, err := dao.NewServiceQueries(dbURL, database, logger)
	if err != nil {
		fmt.Printf("Failed to create ServiceQueries. DB URL: %s, Error: %s", dbURL, err.Error()) // TODO log.Fatalf
	}
	serviceEndpointQueries, err := dao.NewServiceEndpointQueries(dbURL, database, logger)
	if err != nil {
		fmt.Printf("Failed to create ServiceEndpointQueries. DB URL: %s, Error: %s", dbURL, err.Error()) // TODO log.Fatalf
	}
	serviceEndpointDependencyQueries, err := dao.NewServiceEndpointDependencyQueries(dbURL, database, logger)
	if err != nil {
		fmt.Printf("Failed to create ServiceEndpointDependencyQueries. DB URL: %s, Error: %s", dbURL, err.Error()) // TODO log.Fatalf
	}
	// TODO: clean up the below
	var serviceRepository service.Repository
	if dbURL.Driver == "postgresql" {
		serviceRepository = service.NewRepository(
			serviceQueries, serviceEndpointQueries, serviceEndpointDependencyQueries, logger,
		)
	} else if dbURL.Driver == "mysql" {
		serviceRepository = service.NewRepository2(dbURL, database, logger)
	} else {
		log.Fatal("fooooo!!!!")
	}

	router := controller.SetupRouter(serviceRepository, logger)

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
