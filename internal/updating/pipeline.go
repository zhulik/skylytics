package updating

import (
	"context"
	"encoding/json"

	"skylytics/internal/core"
	"skylytics/pkg/async"
	"skylytics/pkg/stormy"

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
		Then(apply.Each(func(_ context.Context, msg jetstream.Msg) error {
			return msg.Ack()
		})).
		Then(parseDIDs).
		Then(apply.Batch[pips.P[jetstream.Msg, string]](100)).
		Then(filterOutExistingAccounts(updater.accountRepo)).
		Then(apply.Rebatch[pips.P[jetstream.Msg, string]](25)).
		Then(
			apply.Map(func(ctx context.Context, wraps []pips.P[jetstream.Msg, string]) (any, error) {
				dids := lo.Map(wraps, func(item pips.P[jetstream.Msg, string], _ int) string {
					return item.B()
				})

				serializedProfiles, err := fetchAndSerializeProfiles(ctx, updater.stormy, dids)
				if err != nil {
					return nil, err
				}

				err = updater.accountRepo.Insert(ctx, serializedProfiles...)
				if err != nil {
					// TODO: handle duplicated records here
					return nil, err
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

func fetchAndSerializeProfiles(ctx context.Context, strmy *stormy.Client, dids []string) ([][]byte, error) {
	profiles, err := strmy.GetProfiles(ctx, dids...)
	if err != nil {
		return nil, err
	}

	return async.AsyncMap(ctx, profiles, func(_ context.Context, profile *stormy.Profile) ([]byte, error) {
		return json.Marshal(profile)
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

		return lo.Reject(wraps, func(item pips.P[jetstream.Msg, string], _ int) bool {
			if lo.Contains(existing, item.B()) {
				item.A().Ack() //nolint:errcheck
				return true
			}
			return false
		}), nil
	})
}
