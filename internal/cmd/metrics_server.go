package cmd

import (
	"context"
	"log/slog"

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
	<-ctx.Done()
	return nil
}
