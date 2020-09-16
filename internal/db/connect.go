package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(user, password, dbname string, port int, ssl bool) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"user=%s password=%s dbname=%s port=%d host=localhost",
		user, password, dbname, port,
	)
	if !ssl {
		dsn += " sslmode=disable"
	}
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
