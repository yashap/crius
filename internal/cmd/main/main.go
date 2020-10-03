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
	migrationDir := os.Getenv("CRIUS_MIGRATIONS_DIR")
	app.NewCrius(dbURL).MigrateDB(migrationDir).ListenAndServe()
}
