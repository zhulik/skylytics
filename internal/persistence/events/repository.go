package events

import (
	"context"

	"skylytics/internal/core"

	"github.com/samber/do"
)

type Repository struct {
}

func NewRepository(_ *do.Injector) (core.EventRepository, error) {
	return Repository{}, nil
}

func (r Repository) Insert(_ context.Context, _ ...core.EventModel) error {
	return nil
}

func (r Repository) HealthCheck() error {
	return nil
}

func (r Repository) Shutdown() error {
	return nil
}
