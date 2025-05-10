package metrics

import (
	"context"
	"log"
	"time"

	"skylytics/internal/core"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm/schema"
)

var (
	tableCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "skylytics_table_estimated_count",
		Help: "Estimated record count for a table.",
	}, []string{"table"})
)

type Collector struct {
	DB core.DB
}

func (c *Collector) Run(ctx context.Context) error {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			log.Println("Collecting metrics")
			err := c.collectTableEstimatedCount(core.AccountModel{})
			if err != nil {
				return err
			}
		}
	}
}

func (c *Collector) collectTableEstimatedCount(tabler schema.Tabler) error {
	var count int64
	count, err := c.DB.EstimatedCount(tabler.TableName())

	if err != nil {
		return err
	}
	tableCount.WithLabelValues(tabler.TableName()).Set(float64(count))
	return nil
}
