package database

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Database struct {
	m  *migrate.Migrate
	db *sql.DB
}

func New(connection string) (*Database, error) {
	db, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	d, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return nil, err
	}
	return &Database{
		m:  m,
		db: db,
	}, nil
}

func (db *Database) Migrate() error {
	if err := db.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func (db *Database) Close() error {
	if _, err := db.m.Close(); err != nil {
		return err
	}
	return db.db.Close()
}
