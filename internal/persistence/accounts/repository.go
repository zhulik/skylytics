package accounts

import (
	"context"

	"skylytics/internal/core"

	"github.com/samber/do"
)

type Repository struct {
	db core.DB
}

func NewRepository(i *do.Injector) (core.AccountRepository, error) {
	return Repository{
		db: do.MustInvoke[core.DB](i),
	}, nil
}

func (r Repository) ExistsByDID(_ context.Context, dids ...string) ([]string, error) {
	var existing []string
	err := r.db.Model(&core.AccountModel{}).
		Select("account->>'did' as did").
		Where("account->>'did' in (?)", dids).
		Find(&existing).Error

	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (r Repository) Insert(_ context.Context, accounts ...core.AccountModel) error {
	return r.db.Model(&core.AccountModel{}).Create(&accounts).Error
}

func (r Repository) HealthCheck() error {
	return nil
}

func (r Repository) Shutdown() error {
	return nil
}
