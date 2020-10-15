package mysql

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file" // For loading migration files
	"github.com/jmoiron/sqlx"
	"github.com/xo/dburl"
)

// Migrate performs all database migrations
func Migrate(db *sqlx.DB, dbURL dburl.URL, migrationDir string) {
	if dbURL.Query().Get("multiStatements") != "true" {
		log.Fatal("For MySQL, your DB URL must set the query param multiStatements=true")
	}
	driver, err := mysql.WithInstance(db.DB, &mysql.Config{})
	if err != nil {
		log.Fatal(err)
	}
	m, err := gomigrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationDir),
		dbURL.Driver,
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}
	m.Steps(2)
}
