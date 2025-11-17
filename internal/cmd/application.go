package cmd

import (
	"context"
	"fmt"
	"os"
	"skylytics/internal/cmd/flags"
	"skylytics/internal/config"
	"skylytics/pkg/clicfg"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/zhulik/pal"
)

const VERSION = "0.1.0"

var cmd = &cli.Command{
	Name:    "skylytics",
	Usage:   "Skylytics is a tool for analyzing the Bluesky network",
	Version: VERSION,
	Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
		if err := initLogger(c.String("log-level")); err != nil {
			return ctx, err
		}
		return ctx, nil
	},
	Flags: []cli.Flag{
		flags.LogLevel,
	},
	Commands: []*cli.Command{
		subscriberCmd,
	},
}

func Run() {
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, c *cli.Command, services ...pal.ServiceDef) error {
	cfg := config.Config{}
	if err := clicfg.ParseFlags(c, &cfg); err != nil {
		return err
	}
	services = append(services, pal.Provide(&cfg))

	return pal.New(services...).
		InjectSlog().
		InitTimeout(2*time.Second).
		HealthCheckTimeout(1*time.Second).
		ShutdownTimeout(10*time.Second).
		Run(ctx, syscall.SIGINT, syscall.SIGTERM)
}
