package service

import (
	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
	"go.uber.org/zap"
	"log"
)

// Repository is a Service repository. It is a classic "Domain Driven Design" repository - the mental model is that
// it represents a collection of models.Service instances
type Repository interface {
	// Save saves a Service
	Save(s *Service) error
	// FindByCode finds a Service by its Code
	FindByCode(code Code) (*Service, error)
}

func NewRepository(
	dbURL *dburl.URL,
	db *sqlx.DB,
	logger *zap.SugaredLogger,
) Repository {
	if dbURL.Driver == "postgres" {
		return &postgresRepository{
			db:     db,
			logger: logger,
		}
	} else if dbURL.Driver == "mysql" {
		return &mysqlRepository{
			db:     db,
			logger: logger,
		}
	}
	log.Fatalf("Unsupported database: %s", dbURL.Driver)
	return nil
}
