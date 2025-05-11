package updating

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"skylytics/internal/core"
	"skylytics/pkg/async"
	"skylytics/pkg/stormy"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/lo"

	"golang.org/x/sync/errgroup"

	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"
)

var (
	parseDIDs = apply.Zip(func(_ context.Context, msg jetstream.Msg) (string, error) {
		var event core.BlueskyEvent
		err := json.Unmarshal(msg.Data(), &event)
		return event.Did, err
	})

	accountsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_updater_accounts_created_total",
		Help: "The total amount of accounts created but the updater.",
	}, []string{"test"})
)

func pipeline(updater *AccountUpdater) *pips.Pipeline[jetstream.Msg, any] {
	return pips.New[jetstream.Msg, any]().
		Then(apply.Each(func(_ context.Context, msg jetstream.Msg) error {
			return msg.Ack()
		})).
		Then(parseDIDs).
		Then(apply.Batch[pips.P[jetstream.Msg, string]](1000)).
		Then(filterOutExistingAccounts(updater.AccountRepo)).
		Then(apply.Rebatch[pips.P[jetstream.Msg, string]](100)).
		Then(
			apply.Map(func(ctx context.Context, wraps []pips.P[jetstream.Msg, string]) (any, error) {
				dids := lo.Map(wraps, func(item pips.P[jetstream.Msg, string], _ int) string {
					return item.B()
				})

				serializedProfiles, err := fetchAndSerializeProfiles(ctx, updater.Stormy, dids)
				if err != nil {
					return nil, err
				}

				err = updater.AccountRepo.Insert(ctx, serializedProfiles...)
				if err != nil {
					var pgError *pgconn.PgError
					if errors.As(err, &pgError) {
						// Ignore duplicate key errors.
						if pgError.Code != "23505" &&
							pgError.Code != "40P01" {
							return nil, err
						}
					} else {
						return nil, err
					}
				}
				accountsCreated.WithLabelValues("test").Add(float64(len(serializedProfiles)))
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

				return core.AccountModel{Account: account}, nil
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

	return result, nil
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
