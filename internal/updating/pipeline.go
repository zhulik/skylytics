package updating

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"skylytics/internal/core"
	"skylytics/pkg/async"
	"skylytics/pkg/stormy"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/lo"

	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"
)

var (
	accountsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_updater_accounts_created_total",
		Help: "The total amount of accounts created but the updater.",
	}, []string{"test"})

	eventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_updater_accounts_evets_processed_total",
		Help: "The total amount of events processed by the account updater.",
	}, []string{"acked"})

	eventProcessingLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "skylytics_updater_event_processing_latency_seconds",
			Help:    "Histogram of event processing latency in the account updater latency in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10, 20, 30, 60, 120},
		},
		[]string{"test"},
	)
)

type pipelineItem struct {
	msg     jetstream.Msg
	event   *core.BlueskyEvent
	exists  bool
	account core.AccountModel

	created time.Time
}

func (p pipelineItem) Ack() {
	eventsProcessed.WithLabelValues("true").Inc()

	eventProcessingLatency.WithLabelValues("test").Observe(time.Since(p.created).Seconds())

	p.msg.Ack() // nolint:errcheck
}

func (p pipelineItem) Nak() {
	eventsProcessed.WithLabelValues("false").Inc()
	p.msg.Nak() // nolint:errcheck
}

func pipeline(updater *AccountUpdater) *pips.Pipeline[jetstream.Msg, any] {
	return pips.New[jetstream.Msg, any]().
		Then( // Parse items
			apply.MapC(4, func(_ context.Context, msg jetstream.Msg) (pipelineItem, error) {
				event := &core.BlueskyEvent{}
				err := json.Unmarshal(msg.Data(), event)
				if err != nil {
					return pipelineItem{}, err
				}

				return pipelineItem{
					msg:     msg,
					event:   event,
					exists:  false,
					created: time.Now(),
				}, nil
			}),
		).
		Then(
			apply.EachC(4, func(_ context.Context, item pipelineItem) error {
				item.msg.InProgress() // nolint:errcheck
				return nil
			}),
		).
		Then(apply.Batch[pipelineItem](500)).
		Then( // Fetch and set existing records
			apply.MapC(4, func(ctx context.Context, items []pipelineItem) ([]pipelineItem, error) {
				updater.Logger.Info("Processing batch", "batch_size", len(items))

				dids := lo.Map(items, func(item pipelineItem, _ int) string {
					item.Ack()
					return item.event.Did
				})

				existing, err := updater.AccountRepo.ExistsByDID(ctx, dids...)
				if err != nil {
					return nil, err
				}

				return lo.Map(items, func(item pipelineItem, _ int) pipelineItem {
					_, exists := existing[item.event.Did]

					item.exists = exists

					return item
				}), nil
			}),
		).
		Then(apply.Flatten[pipelineItem]()).
		Then( // Filter out existing accounts.
			apply.Filter(func(_ context.Context, item pipelineItem) (bool, error) {
				if item.exists {
					item.Ack()
					return false, nil
				}
				return true, nil
			}),
		).
		Then(apply.Batch[pipelineItem](25)).
		Then( // Fetch profiles
			apply.MapC(4, func(ctx context.Context, items []pipelineItem) ([]pipelineItem, error) {
				dids := lo.Map(items, func(item pipelineItem, _ int) string {
					return item.event.Did
				})

				profiles, err := fetchAndSerializeProfiles(ctx, updater.stormy, dids)
				if err != nil {
					return nil, err
				}

				return lo.Map(items, func(item pipelineItem, _ int) pipelineItem {
					item.account = profiles[item.event.Did]
					return item
				}), nil
			}),
		).
		Then(apply.Flatten[pipelineItem]()).
		Then( // Insert records one by one
			apply.Each(func(ctx context.Context, item pipelineItem) error {
				err := updater.AccountRepo.Insert(ctx, item.account)
				if err != nil {
					if !errors.Is(err, jetstream.ErrKeyExists) {
						item.Nak()
						return err
					}
				}

				accountsCreated.WithLabelValues("test").Inc()
				item.Ack()
				return nil
			}),
		)
}

// TODO: turn into pipeline stages.
func fetchAndSerializeProfiles(ctx context.Context, strmy *stormy.Client, dids []string) (map[string]core.AccountModel, error) {
	profiles, err := strmy.GetProfiles(ctx, dids...)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return lo.Associate(models, func(item core.AccountModel) (string, core.AccountModel) {
		return item.DID, item
	}), nil
}
