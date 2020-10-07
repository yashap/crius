package util

import (
	"database/sql"
	"fmt"
	"github.com/xo/dburl"
	"github.com/yashap/crius/internal/db"
	"github.com/yashap/crius/internal/errors"
	"log"
	"path/filepath"
	"strconv"

	_ "github.com/lib/pq" // Postgres driver
	"github.com/ory/dockertest/v3"
)

type TestDB struct {
	pool      *dockertest.Pool
	container *dockertest.Resource
	Database  db.Database
}

func NewTestDB() *TestDB {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	user, password, dbName := "testuser", "testpass", "crius"
	resource, err := pool.Run("postgres", "13", []string{
		"POSTGRES_USER=" + user,
		"POSTGRES_PASSWORD=" + password,
		"POSTGRES_DB=" + dbName,
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	port, _ := strconv.Atoi(resource.GetPort("5432/tcp"))

	relativeMigrationsDir := "../../script/postgresql/migrations"
	migrationsDir, err := filepath.Abs(relativeMigrationsDir)
	if err != nil {
		panic(errors.InitializationError("Could not convert to absolute path",
			errors.Details{"relativePath": relativeMigrationsDir},
			nil,
		))
	}
	rawDBURL := fmt.Sprintf("postgres://%s:%s@localhost:%v/%s?sslmode=disable", user, password, port, dbName)
	dbURL, err := dburl.Parse(rawDBURL)
	if err != nil {
		panic(errors.InitializationError("failed to parse DB URL", errors.Details{"url": rawDBURL}, nil))
	}
	// Ensure container is up and ready to accept connections
	if err = pool.Retry(func() error {
		var err error
		dbase, err := sql.Open(
			dbURL.Driver,
			dbURL.DSN,
		)
		if err != nil {
			return err
		}
		return dbase.Ping()
	}); err != nil {
		shutdown(pool, resource, false)
		log.Fatalf("Could not connect to Postgres Docker container: %s", err.Error())
	}

	database := db.NewDatabase(rawDBURL, migrationsDir)
	database.Migrate()
	testDB := &TestDB{
		pool:      pool,
		container: resource,
		Database:  database,
	}

	return testDB
}

func (tdb *TestDB) Shutdown(fatal bool) {
	shutdown(tdb.pool, tdb.container, fatal)
}

func shutdown(pool *dockertest.Pool, container *dockertest.Resource, fatal bool) {
	if err := pool.Purge(container); err != nil {
		if fatal {
			log.Fatalf("Could not purge Postgres Docker Container: %s\n", err.Error())
		}
		fmt.Printf("Could not purge Postgres Docker Container: %s\n", err.Error())
	}
}
