package database

import (
	"context"
	"database/sql"
	"embed"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Database struct {
	db *bun.DB
}

func New(connection string) (*Database, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connection)))
	if err := sqldb.Ping(); err != nil {
		return nil, err
	}

	return &Database{
		db: bun.NewDB(sqldb, pgdialect.New()),
	}, nil
}

func (db *Database) Migrate() error {
	ctx := context.Background()

	migrations := migrate.NewMigrations()
	if err := migrations.Discover(migrationFiles); err != nil {
		return err
	}

	migrator := migrate.NewMigrator(db.db, migrations, migrate.WithMarkAppliedOnSuccess(true))
	if err := migrator.Init(ctx); err != nil {
		return err
	}

	_, err := migrator.Migrate(ctx)
	return err
}

func (db *Database) Close() error {
	return db.db.Close()
}
