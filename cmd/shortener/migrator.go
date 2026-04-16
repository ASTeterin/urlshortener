package main

import (
	"database/sql"
	"embed"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type MigrationRunner struct {
	db *sql.DB
	m  *migrate.Migrate
}

func NewMigrationRunner(db *sql.DB) (*MigrationRunner, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	sql, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("iofs", sql, "postgres", driver)
	if err != nil {
		return nil, err
	}

	return &MigrationRunner{
		db: db,
		m:  m,
	}, nil
}

func (r *MigrationRunner) Up() error {
	return r.m.Up()
}

func (r *MigrationRunner) Down() error {
	return r.m.Down()
}
