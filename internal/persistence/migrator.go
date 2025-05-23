package persistence

import (
	"context"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file" // nolint:revive

	"skylytics/internal/core"
)

type Migrator struct {
	Logger *slog.Logger
	DB     core.DB

	migrator *migrate.Migrate
}

func (m *Migrator) Init(_ context.Context) error {
	m.Logger = m.Logger.With("component", "migrator")

	db, err := m.DB.DB()
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m.migrator, err = migrate.NewWithDatabaseInstance("file://internal/persistence/migrations", "postgres", driver)
	if err != nil {
		return err
	}

	return nil
}

func (m *Migrator) Up(ctx context.Context) error {
	err := m.Fix(ctx)
	if err != nil {
		return err
	}

	m.Logger.Info("Migrating database up")

	err = m.migrator.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	}

	m.Logger.Info("Database migration completed")
	return nil
}

func (m *Migrator) Down(ctx context.Context) error {
	err := m.Fix(ctx)
	if err != nil {
		return err
	}

	m.Logger.Info("Migrating database down")

	err = m.migrator.Steps(-1)
	if err != nil {
		return err
	}

	m.Logger.Info("Database migration completed")
	return nil
}

func (m *Migrator) Fix(_ context.Context) error {
	version, dirty, err := m.migrator.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return nil
		}
		return err
	}
	if !dirty {
		return nil
	}

	m.Logger.Info("Database is dirty, fixing", "version", version)

	return m.migrator.Force(int(version)) // nolint:gosec
}

func (m *Migrator) Migrate(_ context.Context, version uint) error {
	m.Logger.Info("Migrating database to version", "version", version)

	err := m.migrator.Migrate(version)
	if err != nil {
		return err
	}

	m.Logger.Info("Database migration completed")
	return nil
}
