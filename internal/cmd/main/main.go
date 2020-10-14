package main

import (
	"github.com/yashap/crius/internal/db"
	"github.com/yashap/crius/internal/errors"
	"os"

	"github.com/yashap/crius/internal/app"
)

func main() {
	dbURL := getOrPanic("CRIUS_DB_URL")
	migrationDir := getOrPanic("CRIUS_MIGRATIONS_DIR")
	database := db.NewDatabase(dbURL, migrationDir)
	database.Migrate()
	app.NewCrius(database).ListenAndServe()
}

func getOrPanic(s string) string {
	envVar := os.Getenv(s)
	if envVar == "" {
		panic(errors.InitializationError("environment variable not set", errors.Details{"envVar": s}, nil))
	}
	return envVar
}
