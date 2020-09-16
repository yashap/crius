package db

import (
	"log"

	"github.com/yashap/crius/internal/dao"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&dao.Service{},
		&dao.ServiceEndpoint{},
		&dao.ServiceEndpointDependency{},
	)
	if err != nil {
		log.Fatal(err)
	}
}
