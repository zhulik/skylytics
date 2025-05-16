package accounts

import (
	"context"
	"encoding/base64"
	"log/slog"

	"github.com/samber/lo"

	"skylytics/internal/core"
)

type Repository struct {
	Logger *slog.Logger
	JS     core.JetstreamClient
	KV     core.KeyValueClient
	Config *core.Config
}

func (r *Repository) Init(ctx context.Context) error {
	var err error

	r.Logger = r.Logger.With("component", "accounts.Repository")

	r.KV, err = r.JS.KV(ctx, r.Config.NatsAccountsCacheKVBucket)
	return err
}

func (r *Repository) ExistsByDID(ctx context.Context, dids ...string) (map[string]bool, error) {
	keys := lo.Map(dids, func(did string, _ int) string {
		return base64.StdEncoding.EncodeToString([]byte(did))
	})
	keys = lo.Uniq(keys)

	existing, err := r.KV.ExistingKeys(ctx, keys...)
	if err != nil {
		return nil, err
	}

	return lo.Associate(existing, func(item string) (string, bool) {
		return item, true
	}), nil
}

func (r *Repository) Insert(ctx context.Context, account core.AccountModel) error {
	return r.KV.Put(ctx, account.DID, account.Account)
}
