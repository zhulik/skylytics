package cmd

import (
	"context"
	"log/slog"
	"time"

	"skylytics/internal/aggregator"
	"skylytics/internal/cmd/flags"
	"skylytics/internal/config"
	"skylytics/internal/leaderboard"

	"github.com/urfave/cli/v3"
	"github.com/zhulik/pal"
)

var aggregatorCmd = &cli.Command{
	Name:  "aggregator",
	Usage: "Aggregate raw leaderboard buckets into hourly, daily, and weekly buckets",
	Flags: []cli.Flag{
		flags.RedisAddr,
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		return run(ctx, c,
			aggregator.Provide(),
			pal.Provide(&aggregatorRunner{}),
		)
	},
}

type aggregatorRunner struct {
	Logger     *slog.Logger
	Config     *config.Config
	Summariser *aggregator.Summariser
}

func (r *aggregatorRunner) Run(ctx context.Context) error {
	for {
		timer := time.NewTimer(randomPause(8*time.Minute, 12*time.Minute))
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}

		for _, interaction := range leaderboard.AllInteractions.Members() {
			if err := r.Summariser.SummariseHourly(ctx, interaction); err != nil {
				r.Logger.Error("hourly summarise failed", "interaction", interaction, "error", err)
			}
		}
	}
}
