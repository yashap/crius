package db

import (
	"github.com/jmoiron/sqlx"
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
