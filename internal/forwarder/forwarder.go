package forwarder

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/fx"
	"log"
	"skylytics/internal/core"
)

var (
	totalProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "skylytics_events_processed_total",
		Help: "The total number of processed events",
	})
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
					case <-ch:
						totalProcessed.Inc()
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
