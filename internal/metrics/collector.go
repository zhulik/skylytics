package metrics

import (
	"context"
	"time"

	"skylytics/internal/core"
	"skylytics/pkg/async"

	"github.com/samber/do"
)

type Collector struct {
	handle *async.JobHandle[any]
}

func NewCollector(_ *do.Injector) (core.MetricsCollector, error) {
	collector := Collector{}

	collector.handle = async.Job(func(ctx context.Context) (any, error) {
		ticker := time.NewTicker(15 * time.Second)

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()

				return nil, nil
			case <-ticker.C:
			}
		}
	})

	return &collector, nil
}

func (c Collector) Shutdown() error {
	_, err := c.handle.StopWait()
	return err
}

func (c Collector) HealthCheck() error {
	return c.handle.Error()
}
