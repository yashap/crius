package db

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/xo/dburl"
	"github.com/yashap/crius/internal/criusctx"
	"github.com/yashap/crius/internal/db/mysql"
	"github.com/yashap/crius/internal/db/postgresql"
	"github.com/yashap/crius/internal/errors"
)

type Database interface {
	// Migrate performs all database migrations
	Migrate()
	// GetExecutor returns a boil.ContextExecutor, which can be used to execute queries. If ctx contains an ongoing
	// sql.Tx, it will use that, otherwise it will return a plain sqlx.DB (no transaction)
	GetExecutor(ctx context.Context) boil.ContextExecutor
	// IsPostgres tells you if this Database is a Postgres database
	IsPostgres() bool
	// IsMySQL tells you if this Database is a MySQL database
	IsMySQL() bool
	// URL returns the database url
	URL() dburl.URL
	// BeginTransaction starts a new transaction
	BeginTransaction(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

type database struct {
	dbURL        dburl.URL
	migrationDir string
	underlying   *sqlx.DB
}

// NewDatabase creates a Database, which holds a connection to database at the provided url
func NewDatabase(url, migrationDir string) Database {
	dbURL, err := dburl.Parse(url)
	if err != nil {
		panic(errors.InitializationError("failed to parse DB URL", errors.Details{"url": url}, nil))
	}
	underlyingDB, err := sqlx.Connect(dbURL.Driver, dbURL.DSN)
	if err != nil {
		panic(errors.InitializationError("failed to connect to DB", errors.Details{"url": url}, nil))
	}
	return &database{
		dbURL:        *dbURL,
		migrationDir: migrationDir,
		underlying:   underlyingDB,
	}
}

func (d *database) Migrate() {
	if d.dbURL.Driver == "postgres" {
		postgresql.Migrate(d.underlying, d.dbURL, d.migrationDir)
	} else if d.dbURL.Driver == "mysql" {
		mysql.Migrate(d.underlying, d.dbURL, d.migrationDir)
	} else {
		panic(errors.InitializationError("unsupported database", errors.Details{"driver": d.dbURL.Driver}, nil))
	}
}

func (d *database) GetExecutor(ctx context.Context) boil.ContextExecutor {
	txn, ok := criusctx.GetTransaction(ctx)
	if ok {
		return txn
	}
	return d.underlying
}

func (d *database) IsPostgres() bool {
	return d.dbURL.Driver == "postgres"
}

func (d *database) IsMySQL() bool {
	return d.dbURL.Driver == "mysql"
}

func (d *database) URL() dburl.URL {
	return d.dbURL
}

func (d *database) BeginTransaction(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	txn, err := d.underlying.BeginTxx(ctx, opts)
	if err != nil {
		return nil, errors.DatabaseError("failed to begin transaction", errors.Details{}, &err)
	}
	return txn, nil
}
