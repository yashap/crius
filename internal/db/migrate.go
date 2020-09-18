package db

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // For loading migration files
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Postgres driver
	"github.com/xo/dburl"
)

// Migrate performs all database migrations
func Migrate(db *sqlx.DB, dbURL *dburl.URL, migrationDir string) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationDir),
		dbURL.Driver,
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}
	m.Steps(2)
}
