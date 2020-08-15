package repository

import (
	"database/sql"
	"github.com/yashap/crius/internal/domain/model"
)

type ServiceRepository struct {
	db *sql.DB
}

func NewServiceRespository(db *sql.DB) ServiceRepository {
	return ServiceRepository{db}
}

// Create creates a service
func (serviceRepository *ServiceRepository) Create(service model.Service) (uint64, error) {
	id := uint64(0)
	err := serviceRepository.db.QueryRow(
		`INSERT INTO "service" ("name") VALUES ($1) RETURNING "id"`,
		service.Name,
	).Scan(&id)
	return id, err
}

// FindByID finds a service by id
func (serviceRepository *ServiceRepository) FindByID(id uint64) (model.Service, error) {
	var service model.Service
	err := serviceRepository.db.QueryRow(
		`SELECT "id", "name" FROM "service" WHERE id = $1`,
		id,
	).Scan(&service.ID, &service.Name)
	return service, err
}

// DeleteByID deletes a service by id
func (serviceRepository *ServiceRepository) DeleteByID(id uint64) (int64, error) {
	result, err := serviceRepository.db.Exec(`DELETE FROM "service" WHERE "id" = $1`, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
