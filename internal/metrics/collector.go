package metrics

import (
	"context"
	"time"

	"skylytics/internal/core"
	"skylytics/pkg/async"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/samber/do"
)

var (
	tableCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "skylytics_table_estimated_count",
		Help: "Estimated record count for a table.",
	}, []string{"table"})
)

type Collector struct {
	db     core.DB
	handle *async.JobHandle[any]
}

func NewCollector(i *do.Injector) (core.MetricsCollector, error) {
	collector := Collector{
		db: do.MustInvoke[core.DB](i),
	}

	collector.handle = async.Job(func(ctx context.Context) (any, error) {
		ticker := time.NewTicker(15 * time.Second)

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()

				return nil, nil
			case <-ticker.C:
				var count int64
				err := collector.db.Model(&core.EventModel{}).
					Raw(
						`SELECT reltuples::bigint AS count 
						FROM pg_class 
						WHERE relname = ?`, core.EventModel{}.TableName(),
					).
					Scan(&count).Error

				if err != nil {
					return nil, err
				}
				tableCount.WithLabelValues("events").Set(float64(count))
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
