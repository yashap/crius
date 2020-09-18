package main

import (
	"log"
	"os"

	"github.com/xo/dburl"
	"github.com/yashap/crius/internal/app"
)

func main() {
	rawDBURL := os.Getenv("CRIUS_DB_URL")
	dbURL, err := dburl.Parse(rawDBURL)
	if err != nil {
		log.Fatalf("Failed to parse DB URL: %s", rawDBURL)
	}
	var migrationDir string
	if dbURL.Driver == "postgres" {
		migrationDir = os.Getenv("POSTGRES_MIGRATIONS_DIR")
	} else {
		// TODO mysql support
		log.Fatalf("Unsupported Failed to parse DB URL: %s", rawDBURL)
	}
	app.NewCrius(dbURL).MigrateDB(migrationDir).ListenAndServe()
}
