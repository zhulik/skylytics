package accounts

import (
	"context"

	"skylytics/internal/core"

	"github.com/samber/do"
)

type Repository struct {
}

func NewRepository(_ *do.Injector) (core.AccountRepository, error) {
	return Repository{}, nil
}

func (r Repository) ExistsByDID(_ context.Context, _ ...string) ([]string, error) {
	return nil, nil
}

func (r Repository) InsertRaw(_ context.Context, _ ...[]byte) ([]any, error) {
	return nil, nil
}

func (r Repository) HealthCheck() error {
	return nil
}

func (r Repository) Shutdown() error {
	return nil
}
