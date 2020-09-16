package main

import (
	"os"

	"github.com/yashap/crius/internal/app"
)

func main() {
	app.NewCrius(
		app.DBConfig{
			User:     os.Getenv("APP_DB_USERNAME"),
			Password: os.Getenv("APP_DB_PASSWORD"),
			DBName:   os.Getenv("APP_DB_NAME"),
			Port:     5432,
			SSL:      false,
		},
	).MigrateDB().ListenAndServe()
}
