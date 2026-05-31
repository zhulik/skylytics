package cmd

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"skylytics/internal/cmd/flags"
	"skylytics/internal/config"
	"skylytics/internal/core"

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
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.reportRawBucketCounts(ctx)
		}
	}
}

func (s *metricsServer) reportRawBucketCounts(ctx context.Context) {
	var wg sync.WaitGroup

	wg.Go(func() {
		s.reportRawBucketCount(ctx, "likes:*", "like")
	})
	wg.Go(func() {
		s.reportRawBucketCount(ctx, "reposts:*", "repost")
	})
	wg.Wait()
}

func (s *metricsServer) reportRawBucketCount(ctx context.Context, pattern, content string) {
	count, err := s.Redis.CountKeys(ctx, pattern)
	if err != nil {
		s.Logger.Error("failed to count raw buckets", "content", content, "error", err)
		return
	}
	s.Metrics.SetRawBucketsTotal(ctx, content, float64(count))
}
