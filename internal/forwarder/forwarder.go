package forwarder

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/samber/do"
	"github.com/samber/lo"
	"log"
	"os"
	"skylytics/internal/core"
)

var (
	eventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_events_processed_total",
		Help: "The total number of processed events",
	}, []string{"kind", "operation", "status"})
)

type Forwarder struct {
	stop      chan any
	sub       core.JetstreamSubscriber
	jetstream jetstream.JetStream
}

func (f Forwarder) Shutdown() error {
	f.stop <- true
	return nil
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
		stop:      make(chan any),
		sub:       do.MustInvoke[core.JetstreamSubscriber](i),
		jetstream: js,
	}

	go f.run()

	return f, nil
}

func (f Forwarder) run() {
	ch := f.sub.Chan()

	for {
		select {
		case <-f.stop:
			return

		case result := <-ch:
			event, err := result.Unpack()
			if err != nil {
				continue
			}
			countEvent(event)

			_, err = f.jetstream.Publish(context.Background(), fmt.Sprintf("skylytics.%s", event.Kind), lo.Must(json.Marshal(event)))
			if err != nil {
				log.Printf("error publishing event: %+v", err)
			}
		}
	}
}

func countEvent(event *core.JetstreamEvent) {
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
