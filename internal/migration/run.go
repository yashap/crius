package migration

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Postgres driver must be in scope
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Postgres driver must be in scope
)

// Run the DB migrations
func Run(user, password, dbname string) {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", os.Getenv("MIGRATIONS_DIR")),
		fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable", user, password, dbname),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err.Error() != "no change" {
		log.Fatal(err)
	}
}
