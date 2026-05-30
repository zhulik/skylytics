package core

import (
	"context"
	"time"

	"github.com/bluesky-social/jetstream/pkg/models"
	libredis "github.com/redis/go-redis/v9"
)

type MetricsCollector interface {
	IncJetstreamProcessedEventsTotal(ctx context.Context, kind, operation, collection string)
	IncJetstreamSubscriptionErrorsTotal(ctx context.Context)
}

type EventAnalyzer interface {
	Analyze(ctx context.Context, event *models.Event) error
}

type Redis interface {
	Get(ctx context.Context, key string) *libredis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *libredis.StatusCmd
}
