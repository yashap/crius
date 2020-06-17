package app

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // Postgres driver must be in scope
)

// App represents the application
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// Initialize the app (connect to DB, etc.)
func (a *App) Initialize(user, password, dbname string) {
	connString := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=disable",
		user, password, dbname,
	)

	var err error
	a.DB, err = sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
	}
	runMigrations(user, password, dbname)

	a.Router = mux.NewRouter()
}

// Run the app (start HTTP server)
func (a *App) Run(addr string) {
	fmt.Println("running app")
}
