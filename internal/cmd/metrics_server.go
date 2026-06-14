package cmd

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"skylytics/internal/cmd/flags"
	"skylytics/internal/config"
	"skylytics/internal/core"
	"skylytics/internal/leaderboard"
	"skylytics/pkg/randomtick"

	"github.com/urfave/cli/v3"
	"github.com/zhulik/pal"
)

var metricsServerCmd = &cli.Command{
	Name:  "metrics-server",
	Usage: "Expose metrics to Prometheus",
	Flags: []cli.Flag{
		flags.RedisAddr,
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		return run(ctx, c,
			pal.Provide(&metricsServer{}),
		)
	},
}

type metricsServer struct {
	Logger  *slog.Logger
	Config  *config.Config
	Metrics core.MetricsCollector

	Redis core.Redis
}

func (s *metricsServer) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	for _, interaction := range core.AllInteractions.Members() {
		wg.Go(func() {
			s.runRawBucketCountLoop(ctx, interaction)
		})
	}
	wg.Wait()

	return nil
}

func (s *metricsServer) runRawBucketCountLoop(ctx context.Context, interaction core.Interaction) {
	randomtick.Loop(ctx, 3*time.Minute, 7*time.Minute, func(ctx context.Context) {
		s.reportRawBucketCount(ctx, interaction)
	})
}

func (s *metricsServer) reportRawBucketCount(ctx context.Context, interaction core.Interaction) {
	content := interaction.Value
	prefix := leaderboard.InteractionKeyPrefix(interaction)
	pattern := prefix + "*"

	keys, members, err := s.leaderboardRawBucketStats(ctx, pattern)
	if err != nil {
		s.Logger.Error("failed to count raw buckets", "content", content, "error", err)
		return
	}

	topScore, err := s.leaderboardPenultimateTopScore(ctx, prefix)
	if err != nil {
		s.Logger.Error("failed to read penultimate bucket top score", "content", content, "error", err)
		return
	}

	s.Metrics.SetLeaderboardRawBucketKeysTotal(ctx, content, float64(keys))
	s.Metrics.SetLeaderboardRawBucketMembersTotal(ctx, content, float64(members))
	s.Metrics.SetLeaderboardRawBucketTopScore(ctx, content, topScore)
	s.Logger.Info("counted raw buckets", "content", content, "keys", keys, "members", members, "top_score", topScore)
}

func penultimateRawBucketKey(prefix string) string {
	t := time.Now().UTC().Truncate(5 * time.Minute).Add(-5 * time.Minute)
	return prefix + t.Format("2006-01-02T15:04")
}

func (s *metricsServer) leaderboardPenultimateTopScore(ctx context.Context, prefix string) (float64, error) {
	top, err := s.Redis.ZRevRangeWithScores(ctx, penultimateRawBucketKey(prefix), 0, 0).Result()
	if err != nil {
		return 0, err
	}
	if len(top) == 0 {
		return 0, nil
	}
	return top[0].Score, nil
}

func (s *metricsServer) leaderboardRawBucketStats(ctx context.Context, pattern string) (keys, members int64, err error) {
	var cursor uint64

	for {
		bucketKeys, nextCursor, err := s.Redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return 0, 0, err
		}

		keys += int64(len(bucketKeys))
		for _, key := range bucketKeys {
			card, err := s.Redis.ZCard(ctx, key).Result()
			if err != nil {
				return 0, 0, err
			}
			members += card
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return keys, members, nil
}
