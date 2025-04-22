package events

import (
	"context"

	"skylytics/internal/core"

	"github.com/samber/do"
)

type Repository struct {
	db core.DB
}

func NewRepository(i *do.Injector) (core.EventRepository, error) {
	return Repository{
		db: do.MustInvoke[core.DB](i),
	}, nil
}

func (r Repository) Insert(_ context.Context, events ...core.EventModel) error {
	return r.db.Model(&core.EventModel{}).Create(&events).Error
}

func (r Repository) HealthCheck() error {
	return nil
}

func (r Repository) Shutdown() error {
	return nil
}
