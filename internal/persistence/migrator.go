package persistence

import (
	"context"
	"log/slog"
	"skylytics/internal/core"
)

type Migrator struct {
	Logger *slog.Logger
	DB     core.DB
}

func (m *Migrator) Init(ctx context.Context) error {
	m.Logger = m.Logger.With("component", "migrator")
	return nil
}

func (m *Migrator) Run(context.Context) error {
	m.Logger.Info("Starting database migration")
	err := m.DB.Migrate()
	if err != nil {
		return err
	}
	m.Logger.Info("Database migration completed")
	return nil
}
