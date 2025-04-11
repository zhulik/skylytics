package forwarder

import (
	"context"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/fx"
	"log"
	"skylytics/internal/core"
)

var (
	eventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_events_processed_total",
		Help: "The total number of processed events",
	}, []string{"kind", "operation", "status"})
)

func New(lc fx.Lifecycle, sub core.JetstreamSubscriber) core.Forwarder {
	subCtx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				ch := sub.Chan()
				for {
					select {
					case <-subCtx.Done():
						log.Println("Forwarder stopped")
						return
					case event := <-ch:
						forwardEvent(event)
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Stopping forwarder")
			cancel()
			return nil
		},
	})

	return nil
}

func forwardEvent(event core.JetstreamEvent) {
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
}
