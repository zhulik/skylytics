package forwarder

import (
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/samber/do"
	"skylytics/internal/core"
)

var (
	eventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_events_processed_total",
		Help: "The total number of processed events",
	}, []string{"kind", "operation", "status"})
)

type Forwarder struct {
	stop chan any
	sub  core.JetstreamSubscriber
}

func (f Forwarder) Shutdown() error {
	f.stop <- true
	return nil
}

func New(i *do.Injector) (core.Forwarder, error) {
	f := Forwarder{
		stop: make(chan any),
		sub:  do.MustInvoke[core.JetstreamSubscriber](i),
	}

	go func() {
		ch := f.sub.Chan()
		for {
			select {
			case <-f.stop:
				return
			case event := <-ch:
				forwardEvent(event)
			}
		}
	}()

	return f, nil
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
