package app

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/yashap/crius/internal/controller"
	"github.com/yashap/crius/internal/db"
	"github.com/yashap/crius/internal/domain/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Crius is the Crius application
type Crius interface {
	// MigrateDB runs the DB migrations
	MigrateDB() Crius
	// ListenAndServe starts the HTTP server
	ListenAndServe() Crius

	// ServiceRepository returns the app's service.Repository
	ServiceRepository() *service.Repository
	// ServiceFactory returns the app's service.Factory
	ServiceFactory() *service.Factory
	// Router returns the app's Router
	Router() *gin.Engine
}

type DBConfig struct {
	User     string
	Password string
	DBName   string
	Port     int
	SSL      bool
}

type crius struct {
	db                *gorm.DB
	logger            *zap.SugaredLogger
	serviceRepository *service.Repository
	serviceFactory    *service.Factory
	router            *gin.Engine
}

// NewCrius creates a new Crius application
func NewCrius(dbConfig DBConfig) Crius {
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	database, err := db.Connect(
		dbConfig.User,
		dbConfig.Password,
		dbConfig.DBName,
		dbConfig.Port,
		dbConfig.SSL,
	)
	if err != nil {
		log.Fatal(err)
	}
	serviceRepository := service.NewRepository(database, logger)
	serviceFactory := service.NewFactory(database, logger)
	router := controller.SetupRouter(&serviceRepository, &serviceFactory, logger)

	return &crius{
		db:                database,
		logger:            logger,
		serviceRepository: &serviceRepository,
		serviceFactory:    &serviceFactory,
		router:            router,
	}
}

func (c *crius) MigrateDB() Crius {
	db.AutoMigrate(c.db)
	return c
}

func (c *crius) ListenAndServe() Crius {
	err := c.router.Run()
	if err != nil {
		log.Fatal("Failed to run server" + err.Error())
	}
	return c
}

func (c *crius) ServiceRepository() *service.Repository {
	return c.serviceRepository
}

func (c *crius) ServiceFactory() *service.Factory {
	return c.serviceFactory
}

func (c *crius) Router() *gin.Engine {
	return c.router
}
