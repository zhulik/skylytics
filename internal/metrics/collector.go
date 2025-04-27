package metrics

import (
	"context"
	"time"

	"skylytics/internal/core"
	"skylytics/pkg/async"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/samber/do"
	"gorm.io/gorm/schema"
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
				err := collectTableEstimatedCount(collector, core.AccountModel{})
				if err != nil {
					return nil, err
				}
			}
		}
	})

	return &collector, nil
}

func collectTableEstimatedCount(collector Collector, tabler schema.Tabler) error {
	var count int64
	count, err := collector.db.EstimatedCount(tabler.TableName())

	if err != nil {
		return err
	}
	tableCount.WithLabelValues(tabler.TableName()).Set(float64(count))
	return nil
}

func (c Collector) Shutdown() error {
	_, err := c.handle.StopWait()
	return err
}

func (c Collector) HealthCheck() error {
	return c.handle.Error()
}
