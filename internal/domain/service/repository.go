package service

import (
	"gorm.io/gorm"

	"go.uber.org/zap"
)

// Repository is a service repository. It is a classic "Domain Driven Design" repository, representing a
// collection of models.Service instances
type Repository struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewRepository(db *gorm.DB, logger *zap.SugaredLogger) Repository {
	return Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) FindByCode(code ServiceCode) (*Service, error) {
	var service Service
	err := r.db.Where(&Service{Code: code}).First(&service).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}
