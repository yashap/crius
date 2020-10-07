package service

import (
	"context"
	"github.com/yashap/crius/internal/db"
	"github.com/yashap/crius/internal/errors"
	"go.uber.org/zap"
)

// Repository is a Service repository. It is a classic "Domain Driven Design" repository - the mental model is that
// it represents a collection of models.Service instances
type Repository interface {
	// Save saves a Service
	Save(ctx context.Context, s *Service) error
	// FindByCode finds a Service by its Code
	FindByCode(ctx context.Context, code Code) (*Service, error)
	// FindAll finds every Service
	FindAll(ctx context.Context) ([]Service, error)
}

func NewRepository(database db.Database, logger *zap.SugaredLogger) Repository {
	if database.IsPostgres() {
		return &postgresRepository{
			database: database,
			logger:   logger,
		}
	} else if database.IsMySQL() {
		return &mysqlRepository{
			database: database,
			logger:   logger,
		}
	}
	panic(errors.InitializationError("Unsupported database", errors.Details{}, nil))
}
