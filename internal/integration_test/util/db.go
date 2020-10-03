package util

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/lib/pq" // Postgres driver
	"github.com/ory/dockertest/v3"
	"github.com/xo/dburl"
)

type TestDB struct {
	pool      *dockertest.Pool
	container *dockertest.Resource
	URL       *dburl.URL
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
	rawDBURL := fmt.Sprintf("postgres://%s:%s@localhost:%v/%s?sslmode=disable", user, password, port, dbName)
	dbURL, err := dburl.Parse(rawDBURL)
	if err != nil {
		log.Fatalf("Could not parse DB URL: %s %s", rawDBURL, err)
	}

	testDB := &TestDB{
		pool:      pool,
		container: resource,
		URL:       dbURL,
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
		testDB.Shutdown(false)
		log.Fatalf("Could not connect to Postgres Docker container: %s", err.Error())
	}

	return testDB
}

func (tdb *TestDB) Shutdown(fatal bool) {
	if err := tdb.pool.Purge(tdb.container); err != nil {
		if fatal {
			log.Fatalf("Could not purge Postgres Docker Container: %s\n", err.Error())
		}
		fmt.Printf("Could not purge Postgres Docker Container: %s\n", err.Error())
	}
}
