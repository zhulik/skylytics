package core

import (
	"context"
	"time"

	"github.com/bluesky-social/jetstream/pkg/models"
	libredis "github.com/redis/go-redis/v9"
)

type MetricsCollector interface {
	IncJetstreamProcessedEventsTotal(ctx context.Context, kind, operation, collection string)
	ObserveJetstreamEventLag(ctx context.Context, lag time.Duration)
	ObserveEventProcessingDuration(ctx context.Context, duration time.Duration)
	IncJetstreamSubscriptionErrorsTotal(ctx context.Context, err error)
	IncBlueskyPostsTotal(ctx context.Context, languageCount, imageCount int)
	IncBlueskyPostsByLanguageTotal(ctx context.Context, language string)
	SetLeaderboardRawBucketKeysTotal(ctx context.Context, content string, count float64)
	SetLeaderboardRawBucketMembersTotal(ctx context.Context, content string, count float64)
	SetLeaderboardRawBucketTopScore(ctx context.Context, content string, score float64)
	IncPostInteractionsTotal(ctx context.Context, interaction Interaction)
}

type EventAnalyzer interface {
	Analyze(ctx context.Context, event *models.Event) error
}

type Redis interface {
	Get(ctx context.Context, key string) *libredis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *libredis.StatusCmd

	ZIncrBy(ctx context.Context, key string, increment float64, member string) *libredis.FloatCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *libredis.BoolCmd

	Scan(ctx context.Context, cursor uint64, match string, count int64) *libredis.ScanCmd
	ZCard(ctx context.Context, key string) *libredis.IntCmd
	ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) *libredis.ZSliceCmd
	ZUnionStore(ctx context.Context, dest string, store *libredis.ZStore) *libredis.IntCmd
}
