package updating

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/lo"
	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"
	"skylytics/internal/core"
	"skylytics/pkg/async"
	"skylytics/pkg/stormy"
)

var (
	parseDIDs = apply.Map(func(ctx context.Context, msg jetstream.Msg) (msgWrap[string], error) {
		var event core.BlueskyEvent
		err := json.Unmarshal(msg.Data(), &event)
		return msgWrap[string]{msg, event.Did}, err
	})
)

type msgWrap[T any] struct {
	msg  jetstream.Msg
	data T
}

type msgWrapSlice[T any] []msgWrap[T]

func (s msgWrapSlice[T]) Data() []T {
	return lo.Map(s, func(item msgWrap[T], _ int) T {
		return item.data
	})
}

func pipeline(updater *AccountUpdater) *pips.Pipeline[jetstream.Msg, any] {
	return pips.New[jetstream.Msg, any]().
		Then(parseDIDs).
		Then(apply.Batch[msgWrap[string]](100)).
		Then(filterOutExistingAccounts(updater.accountRepo)).
		Then(apply.Rebatch[msgWrap[string]](25)).
		Then(
			apply.Map(func(ctx context.Context, wraps msgWrapSlice[string]) (any, error) {
				dids := wraps.Data()

				serializedProfiles, err := fetchAndSerializeProfiles(ctx, updater.stormy, dids)
				if err != nil {
					return nil, err
				}

				_, err = updater.accountRepo.InsertRaw(ctx, serializedProfiles...)
				if err != nil {
					return nil, err
				}
				lo.ForEach(wraps, func(item msgWrap[string], _ int) {
					item.msg.Ack()
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
	return apply.Map(func(ctx context.Context, wraps msgWrapSlice[string]) (msgWrapSlice[string], error) {
		dids := lo.Uniq(wraps.Data())

		existing, err := repo.ExistsByDID(ctx, dids...)
		if err != nil {
			return nil, err
		}

		return lo.Reject(wraps, func(item msgWrap[string], _ int) bool {
			if lo.Contains(existing, item.data) {
				item.msg.Ack()
				return true
			}
			return false
		}), nil
	})
}
