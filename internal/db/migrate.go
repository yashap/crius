package db

import (
	_ "github.com/golang-migrate/migrate/v4/source/file" // For loading migration files
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Postgres driver
	"github.com/xo/dburl"
	"github.com/yashap/crius/internal/db/mysql"
	"github.com/yashap/crius/internal/db/postgresql"
	"log"
)

// Migrate performs all database migrations
func Migrate(db *sqlx.DB, dbURL *dburl.URL, migrationDir string) {
	if dbURL.Driver == "postgres" {
		postgresql.Migrate(db, dbURL, migrationDir)
	} else if dbURL.Driver == "mysql" {
		mysql.Migrate(db, dbURL, migrationDir)
	} else {
		log.Fatalf("Unsupported database: %s", dbURL.Driver)
	}
}
