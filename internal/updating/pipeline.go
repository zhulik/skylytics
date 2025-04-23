package updating

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"skylytics/internal/core"
	"skylytics/pkg/async"
	"skylytics/pkg/stormy"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/lo"

	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"
)

var (
	parseDIDs = apply.Zip(func(_ context.Context, msg jetstream.Msg) (string, error) {
		var event core.BlueskyEvent
		err := json.Unmarshal(msg.Data(), &event)
		return event.Did, err
	})
)

func pipeline(updater *AccountUpdater) *pips.Pipeline[jetstream.Msg, any] {
	return pips.New[jetstream.Msg, any]().
		Then(parseDIDs).
		Then(apply.Batch[pips.P[jetstream.Msg, string]](100)).
		Then(filterOutExistingAccounts(updater.accountRepo)).
		Then(apply.Rebatch[pips.P[jetstream.Msg, string]](25)).
		Then(
			apply.Map(func(ctx context.Context, wraps []pips.P[jetstream.Msg, string]) (any, error) {
				log.Printf("processing %d", len(wraps))

				dids := lo.Map(wraps, func(item pips.P[jetstream.Msg, string], _ int) string {
					return item.B()
				})

				serializedProfiles, err := fetchAndSerializeProfiles(ctx, updater.stormy, dids)
				if err != nil {
					return nil, err
				}

				err = updater.accountRepo.Insert(ctx, serializedProfiles...)
				if err != nil {
					// Ignore duplicate key errors.
					var pgError *pgconn.PgError
					if !errors.As(err, &pgError) || pgError.Code != "23505" {
						return nil, err
					}
				}
				lo.ForEach(wraps, func(item pips.P[jetstream.Msg, string], _ int) {
					if item.A() != nil {
						item.A().Ack() //nolint:errcheck
					}
				})

				return nil, nil
			}),
		)
}

func fetchAndSerializeProfiles(ctx context.Context, strmy *stormy.Client, dids []string) ([]core.AccountModel, error) {
	profiles, err := strmy.GetProfiles(ctx, dids...)
	if err != nil {
		return nil, err
	}

	return async.AsyncMap(ctx, profiles, func(_ context.Context, profile *stormy.Profile) (core.AccountModel, error) {
		account, err := json.Marshal(profile)
		if err != nil {
			return core.AccountModel{}, err
		}
		var accountModel core.AccountModel

		err = json.Unmarshal(account, &accountModel)
		if err != nil {
			return core.AccountModel{}, err
		}

		return core.AccountModel{Account: account}, nil
	})
}

func filterOutExistingAccounts(repo core.AccountRepository) pips.Stage {
	return apply.Map(func(ctx context.Context, wraps []pips.P[jetstream.Msg, string]) ([]pips.P[jetstream.Msg, string], error) {
		dids := lo.Map(wraps, func(item pips.P[jetstream.Msg, string], _ int) string {
			return item.B()
		})

		existing, err := repo.ExistsByDID(ctx, dids...)
		if err != nil {
			return nil, err
		}

		log.Printf("existing: %d", len(existing))

		return lo.Reject(wraps, func(item pips.P[jetstream.Msg, string], _ int) bool {
			if lo.Contains(existing, item.B()) {
				item.A().Ack() //nolint:errcheck
				return true
			}
			return false
		}), nil
	})
}
