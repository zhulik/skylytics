package accounts

import (
	"context"

	"github.com/samber/lo"

	"skylytics/internal/core"
)

type Repository struct {
	DB core.DB
}

func (r *Repository) ExistsByDID(ctx context.Context, dids ...string) (map[string]bool, error) {
	var existing []string
	err := r.DB.
		Model(&core.AccountModel{}).
		WithContext(ctx).
		Select("account->>'did' as did").
		Where("account->>'did' in (?)", dids).
		Find(&existing).Error

	if err != nil {
		return nil, err
	}
	return lo.Associate(existing, func(item string) (string, bool) {
		return item, true
	}), nil
}

func (r *Repository) Insert(_ context.Context, accounts ...core.AccountModel) error {
	return r.DB.Model(&core.AccountModel{}).Create(&accounts).Error
}
