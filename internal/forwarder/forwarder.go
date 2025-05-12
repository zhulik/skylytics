package forwarder

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"

	"skylytics/internal/core"

	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	eventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_events_processed_total",
		Help: "The total number of processed events",
	}, []string{"kind", "operation", "status"})

	commitProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_commit_processed_total",
		Help: "The total number of processed commits",
	}, []string{"commit_type", "operation"})
)

type Forwarder struct {
	Sub core.BlueskySubscriber
	JS  core.JetstreamClient
}

func (f *Forwarder) Run(ctx context.Context) error {
	return pips.New[*core.BlueskyEvent, any]().
		Then(apply.Each(countEvent)).
		Then(
			apply.Map(func(ctx context.Context, event *core.BlueskyEvent) (any, error) {
				payload, err := json.Marshal(event)
				if err != nil {
					return nil, err
				}

				did64 := base64.StdEncoding.EncodeToString([]byte(event.Did))

				subject := fmt.Sprintf("event.%s.%s", event.Kind, did64)

				return f.JS.Publish(
					ctx,
					subject,
					payload,
				)
			}),
		).
		Run(ctx, f.Sub.C()).
		Wait(ctx)
}

func countEvent(_ context.Context, event *core.BlueskyEvent) error {
	operation := ""
	status := ""

	switch event.Kind {
	case models.EventKindCommit:
		operation = event.Commit.Operation
		commitProcessed.WithLabelValues(event.Commit.Collection, event.Commit.Operation).Inc()
	case models.EventKindAccount:
		if event.Account.Status != nil {
			status = *event.Account.Status
		}
	case models.EventKindIdentity:
	}

	eventsProcessed.WithLabelValues(event.Kind, operation, status).Inc()

	return nil
}
