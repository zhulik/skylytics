package updating

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"skylytics/internal/core"
	"skylytics/pkg/async"
	"skylytics/pkg/stormy"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/lo"

	"golang.org/x/sync/errgroup"

	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"
)

var (
	parseItems = apply.Map(func(_ context.Context, msg jetstream.Msg) (pipelineItem, error) {
		event := &core.BlueskyEvent{}
		err := json.Unmarshal(msg.Data(), event)
		if err != nil {
			return pipelineItem{}, err
		}

		return pipelineItem{
			msg:    msg,
			event:  event,
			exists: false,
		}, nil
	})

	filterOutExisting = apply.Filter[pipelineItem](func(_ context.Context, item pipelineItem) (bool, error) {
		if item.exists {
			return false, item.msg.Ack()
		}
		return true, nil
	})

	accountsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_updater_accounts_created_total",
		Help: "The total amount of accounts created but the updater.",
	}, []string{"test"})
)

type pipelineItem struct {
	msg     jetstream.Msg
	event   *core.BlueskyEvent
	exists  bool
	account core.AccountModel
}

func pipeline(updater *AccountUpdater) *pips.Pipeline[jetstream.Msg, any] {
	return pips.New[jetstream.Msg, any]().
		Then(parseItems).
		Then(apply.Batch[pipelineItem](1000)).
		Then(fetchExistingAccounts(updater)).
		Then(apply.Flatten[pipelineItem]()).
		Then(apply.Each(func(_ context.Context, item pipelineItem) error {
			return item.msg.Ack()
		})).
		Then(filterOutExisting).
		Then(apply.Batch[pipelineItem](100)).
		Then(fetchProfiles(updater)).
		Then(tryInsertInBatches(updater)).
		Then(apply.Flatten[pipelineItem]()).
		Then(insertOneByOne(updater))
}

func fetchAndSerializeProfiles(ctx context.Context, strmy *stormy.Client, dids []string) (map[string]core.AccountModel, error) {
	g, ctx := errgroup.WithContext(ctx)
	resultChan := make(chan []core.AccountModel, (len(dids)+24)/25)

	chunks := lo.Chunk(dids, 25)
	for _, chunk := range chunks {
		didChunk := chunk
		g.Go(func() error {
			profiles, err := strmy.GetProfiles(ctx, didChunk...)
			if err != nil {
				return err
			}

			models, err := async.AsyncMap(ctx, profiles, func(_ context.Context, profile *stormy.Profile) (core.AccountModel, error) {
				account, err := json.Marshal(profile)
				if err != nil {
					return core.AccountModel{}, err
				}
				var accountModel core.AccountModel

				err = json.Unmarshal(account, &accountModel)
				if err != nil {
					return core.AccountModel{}, err
				}

				return core.AccountModel{Account: account, DID: profile.DID}, nil
			})
			if err != nil {
				return err
			}

			resultChan <- models
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	close(resultChan)

	var result []core.AccountModel
	for models := range resultChan {
		result = append(result, models...)
	}

	return lo.Associate(result, func(item core.AccountModel) (string, core.AccountModel) {
		return item.DID, item
	}), nil
}

func fetchExistingAccounts(updater *AccountUpdater) pips.Stage {
	return apply.Map(func(ctx context.Context, items []pipelineItem) ([]pipelineItem, error) {
		dids := lo.Map(items, func(item pipelineItem, _ int) string {
			return item.event.Did
		})

		existing, err := updater.AccountRepo.ExistsByDID(ctx, dids...)
		if err != nil {
			return nil, err
		}

		return lo.Map(items, func(item pipelineItem, _ int) pipelineItem {
			_, exists := existing[item.event.Did]

			return pipelineItem{
				msg:    item.msg,
				event:  item.event,
				exists: exists,
			}
		}), nil
	})
}

func fetchProfiles(updater *AccountUpdater) pips.Stage {
	return apply.Map(func(ctx context.Context, items []pipelineItem) ([]pipelineItem, error) {
		dids := lo.Map(items, func(item pipelineItem, _ int) string {
			return item.event.Did
		})

		profiles, err := fetchAndSerializeProfiles(ctx, updater.stormy, dids)
		if err != nil {
			return nil, err
		}

		return lo.Map(items, func(item pipelineItem, _ int) pipelineItem {
			return pipelineItem{
				msg:     item.msg,
				event:   item.event,
				exists:  item.exists,
				account: profiles[item.event.Did],
			}
		}), nil
	})
}

func tryInsertInBatches(updater *AccountUpdater) pips.Stage {
	return apply.Map(func(ctx context.Context, items []pipelineItem) ([]pipelineItem, error) {
		serializedProfiles := lo.Map(items, func(item pipelineItem, _ int) core.AccountModel {
			return item.account
		})

		err := updater.AccountRepo.Insert(ctx, serializedProfiles...)
		if err == nil {
			lo.ForEach(items, func(item pipelineItem, _ int) {
				accountsCreated.WithLabelValues("test").Add(float64(len(serializedProfiles)))
				item.msg.Ack() //nolint:errcheck
			})
			return nil, nil
		}

		updater.Logger.Error("failed to insert accounts batch, processing them one by one", "error", err)

		return items, nil
	})
}

func insertOneByOne(updater *AccountUpdater) pips.Stage {
	return apply.Each(func(ctx context.Context, item pipelineItem) error {
		err := updater.AccountRepo.Insert(ctx, item.account)
		if err != nil {
			var pgError *pgconn.PgError
			if errors.As(err, &pgError) {
				// Ignore duplicate key errors.
				if pgError.Code != "23505" && pgError.Code != "40P01" {
					return err
				}
			} else {
				return err
			}
		}

		accountsCreated.WithLabelValues("test").Inc()
		return item.msg.Ack()
	})
}
