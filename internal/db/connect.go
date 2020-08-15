package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Postgres driver must be in scope
	"log"
)

func Connect(user, password, dbname string) *sql.DB {
	connString := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=disable",
		user, password, dbname,
	)

	var err error
	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
