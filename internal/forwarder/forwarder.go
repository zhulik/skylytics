package forwarder

import (
	"context"
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
)

type Forwarder struct {
	Sub core.BlueskySubscriber
	JS  core.JetstreamClient
}

func (f *Forwarder) Run(ctx context.Context) error {
	return pips.New[core.BlueskyEvent, any]().
		Then(apply.Each(countEvent)).
		Then(
			apply.Map(func(ctx context.Context, event core.BlueskyEvent) (any, error) {
				payload, err := json.Marshal(event)
				if err != nil {
					return nil, err
				}

				return f.JS.Publish(
					ctx,
					fmt.Sprintf("skylytics.events.%s", event.Kind),
					payload,
				)
			}),
		).
		Run(ctx, f.Sub.C()).
		Wait(ctx)
}

func countEvent(_ context.Context, event core.BlueskyEvent) error {
	operation := ""
	status := ""

	switch event.Kind {
	case models.EventKindCommit:
		operation = event.Commit.Operation
	case models.EventKindAccount:
		if event.Account.Status != nil {
			status = *event.Account.Status
		}
	case models.EventKindIdentity:
	}

	eventsProcessed.WithLabelValues(event.Kind, operation, status).Inc()

	return nil
}
