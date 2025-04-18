package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/samber/do"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"os"
	"skylytics/internal/core"
	"skylytics/internal/persistence"
	"skylytics/pkg/async"
	"time"
)

var (
	collectionCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mongodb_collection_estimated_count",
		Help: "Estimated document count in MongoDB collection",
	}, []string{"collection"})
)

type Collector struct {
	handle *async.JobHandle[any]
}

func NewCollector(i *do.Injector) (core.MetricsCollector, error) {
	collector := Collector{}

	collector.handle = async.Job(func(ctx context.Context) (any, error) {
		ticker := time.NewTicker(15 * time.Second)

		uri := os.Getenv("MONGODB_URI")
		if uri == "" {
			return nil, persistence.ErrNoMongodbURI
		}

		client, err := mongo.Connect(options.Client().ApplyURI(uri))
		if err != nil {
			return nil, err
		}

		defer client.Disconnect(ctx)

		accounts := client.Database("admin").Collection("accounts")
		events := client.Database("admin").Collection("events")

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return nil, nil
			case <-ticker.C:
				log.Println("Updating metrics")
				count, err := accounts.EstimatedDocumentCount(ctx, nil)
				if err != nil {
					return nil, err
				}
				collectionCount.WithLabelValues("accounts").Set(float64(count))

				count, err = events.EstimatedDocumentCount(ctx, nil)
				if err != nil {
					return nil, err
				}
				collectionCount.WithLabelValues("events").Set(float64(count))
			}
		}
	})

	return &collector, nil
}

func (c Collector) Shutdown() error {
	c.handle.Stop()
	_, err := c.handle.Wait()
	return err
}

func (c Collector) HealthCheck() error {
	return c.handle.Error()
}
