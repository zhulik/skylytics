package core

import (
	"context"
	"time"

	"github.com/bluesky-social/jetstream/pkg/models"
	libredis "github.com/redis/go-redis/v9"
)

type MetricsCollector interface {
	IncJetstreamProcessedEventsTotal(ctx context.Context, kind, operation, collection string)
	IncJetstreamSubscriptionErrorsTotal(ctx context.Context, err error)
	IncBlueskyPostCreated(ctx context.Context, languageCount, imageCount int)
	IncBlueskyPostCreatedInLanguage(ctx context.Context, language string)
	SetRawBucketsTotal(ctx context.Context, content string, count float64)
	IncPostInteracted(ctx context.Context, interation string)
}

type EventAnalyzer interface {
	Analyze(ctx context.Context, event *models.Event) error
}

type Redis interface {
	Get(ctx context.Context, key string) *libredis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *libredis.StatusCmd

	ZIncrBy(ctx context.Context, key string, increment float64, member string) *libredis.FloatCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *libredis.BoolCmd
	CountKeys(ctx context.Context, pattern string) (int64, error)
}
