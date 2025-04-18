package forwarder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"skylytics/pkg/async"

	"skylytics/internal/core"

	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/samber/do"
)

var (
	eventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_events_processed_total",
		Help: "The total number of processed events",
	}, []string{"kind", "operation", "status"})
)

type Forwarder struct {
	sub       core.BlueskySubscriber
	jetstream jetstream.JetStream

	handle *async.JobHandle[any]
}

func New(i *do.Injector) (core.Forwarder, error) {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	f := Forwarder{
		sub:       do.MustInvoke[core.BlueskySubscriber](i),
		jetstream: js,
	}

	handle := async.Job(func(ctx context.Context) (any, error) {
		ch := f.sub.Chan()

		for {
			select {
			case <-ctx.Done():
				return nil, nil
			case result := <-ch:
				event, err := result.Unpack()
				if err != nil {
					return nil, err
				}
				err = f.Forward(ctx, event)
				if err != nil {
					return nil, err
				}
			}
		}
	})

	f.handle = handle

	return f, nil
}

func (f Forwarder) HealthCheck() error {
	return f.handle.Error()
}

func (f Forwarder) Shutdown() error {
	f.handle.Stop()
	_, err := f.handle.Wait()
	return err
}

func (f Forwarder) Forward(ctx context.Context, event core.BlueskyEvent) error {
	countEvent(event)

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = f.jetstream.Publish(
		ctx,
		fmt.Sprintf("skylytics.events.%s", event.Kind),
		payload,
	)
	return err
}

func countEvent(event core.BlueskyEvent) {
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
