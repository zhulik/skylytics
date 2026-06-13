package aggregator

import (
	"context"
	"log/slog"
	"skylytics/internal/core"
	"skylytics/internal/leaderboard"
	"time"

	libredis "github.com/redis/go-redis/v9"
)

const hourlyBucketTTL = 14 * 24 * time.Hour

type AggregationSpec struct {
	SourceKeys func(periodStart time.Time) []string
	DestKey    func(periodStart time.Time) string
	TTL        time.Duration
}

type Summariser struct {
	Logger *slog.Logger
	Redis  core.Redis
}

func (s *Summariser) SummariseHourly(ctx context.Context, interaction leaderboard.Interaction) error {
	hour := leaderboard.LastCompleteHour(time.Now())
	spec := AggregationSpec{
		SourceKeys: func(periodStart time.Time) []string {
			return leaderboard.RawKeysForHour(interaction, periodStart)
		},
		DestKey: func(periodStart time.Time) string {
			return leaderboard.HourlyKey(interaction, periodStart)
		},
		TTL: hourlyBucketTTL,
	}

	destKey, err := s.aggregate(ctx, spec, hour)
	if err != nil {
		return err
	}

	s.Logger.Info("summarised hourly bucket", "interaction", interaction, "hour", leaderboard.HourBucket(hour), "dest", destKey)
	return nil
}

func (s *Summariser) aggregate(ctx context.Context, spec AggregationSpec, periodStart time.Time) (string, error) {
	sourceKeys := spec.SourceKeys(periodStart)
	destKey := spec.DestKey(periodStart)

	err := s.Redis.ZUnionStore(ctx, destKey, &libredis.ZStore{
		Keys:      sourceKeys,
		Aggregate: "SUM",
	}).Err()
	if err != nil {
		return "", err
	}

	err = s.Redis.Expire(ctx, destKey, spec.TTL).Err()
	if err != nil {
		return "", err
	}

	return destKey, nil
}
