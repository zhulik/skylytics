package accounts

import (
	"context"

	"skylytics/internal/core"
)

type Repository struct {
	DB core.DB
}

func (r *Repository) ExistsByDID(_ context.Context, dids ...string) ([]string, error) {
	var existing []string
	err := r.DB.Model(&core.AccountModel{}).
		Select("account->>'did' as did").
		Where("account->>'did' in (?)", dids).
		Find(&existing).Error

	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (r *Repository) Insert(_ context.Context, accounts ...core.AccountModel) error {
	return r.DB.Model(&core.AccountModel{}).Create(&accounts).Error
}
