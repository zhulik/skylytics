package persistence

import (
	"context"

	"skylytics/internal/core"
)

type MigrationUpRunner struct {
	Migrator core.Migrator
}

func (m *MigrationUpRunner) Run(ctx context.Context) error {
	return m.Migrator.Up(ctx)
}

type MigrationDownRunner struct {
	Migrator core.Migrator
}

func (m *MigrationDownRunner) Run(ctx context.Context) error {
	return m.Migrator.Down(ctx)
}
