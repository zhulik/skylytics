package cmd

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math/big"
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
	var wg sync.WaitGroup

	wg.Go(func() {
		s.runRawBucketCountLoop(ctx, "likes:*", "like")
	})
	wg.Go(func() {
		s.runRawBucketCountLoop(ctx, "reposts:*", "repost")
	})
	wg.Go(func() {
		s.runRawBucketCountLoop(ctx, "quotes:*", "quote")
	})
	wg.Go(func() {
		s.runRawBucketCountLoop(ctx, "replies:*", "reply")
	})
	wg.Wait()

	return nil
}

func (s *metricsServer) runRawBucketCountLoop(ctx context.Context, pattern, content string) {
	for {
		timer := time.NewTimer(randomPause(3*time.Minute, 7*time.Minute))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}

		s.reportRawBucketCount(ctx, pattern, content)
	}
}

func (s *metricsServer) reportRawBucketCount(ctx context.Context, pattern, content string) {
	count, err := s.Redis.CountKeys(ctx, pattern)
	if err != nil {
		s.Logger.Error("failed to count raw buckets", "content", content, "error", err)
		return
	}

	s.Metrics.SetRawBucketsTotal(ctx, content, float64(count))
	s.Logger.Info("counted raw buckets", "content", content, "count", count)
}

func randomPause(minPause, maxPause time.Duration) time.Duration {
	span := big.NewInt(int64(maxPause - minPause + 1))
	n, err := rand.Int(rand.Reader, span)
	if err != nil {
		return minPause
	}
	return minPause + time.Duration(n.Int64())
}
