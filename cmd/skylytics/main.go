package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"skylytics/pkg/stormy"
	"syscall"
	"time"

	"github.com/golang-cz/devslog"
	"github.com/mattn/go-isatty"

	"skylytics/internal/analytics"
	"skylytics/internal/api"
	"skylytics/internal/persistence/posts"

	"skylytics/internal/bluesky"
	"skylytics/internal/core"
	"skylytics/internal/forwarder"
	"skylytics/internal/metrics"
	inats "skylytics/internal/nats"
	"skylytics/internal/persistence"

	"github.com/zhulik/pal"
	// "github.com/zhulik/pal/inspect"
)

func main() {
	w := os.Stdout

	var handler slog.Handler
	if isatty.IsTerminal(w.Fd()) {
		handler = devslog.NewHandler(w, nil)
	} else {
		handler = slog.NewJSONHandler(w, nil)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	services := []pal.ServiceDef{
		pal.ProvideConst[*slog.Logger](logger),
		pal.Provide[core.JetstreamClient, inats.Client](),
		pal.ProvideFn[*stormy.Client](func(_ context.Context) (*stormy.Client, error) {
			return stormy.NewClient(nil), nil
		}),
		pal.Provide[*core.Config, core.Config](),
	}

	command := os.Args[1]
	switch command {
	case "subscriber":
		// services = append(services, inspect.Provide()...)
		services = append(services,
			pal.Provide[core.MetricsServer, metrics.HTTPServer](),
			pal.Provide[core.BlueskySubscriber, bluesky.Subscriber](),
			pal.Provide[core.Forwarder, forwarder.Forwarder](),
		)

	case "metrics-server":
		services = append(services,
			pal.Provide[core.MetricsServer, metrics.HTTPServer](),
			pal.Provide[core.DB, persistence.DB](),
			pal.Provide[core.MetricsCollector, metrics.Collector]())

	case "repl":
		err := pal.New(
		// pal.ProvideRunner(inspect.RemoteConsole),
		).
			InitTimeout(2*time.Second).
			HealthCheckTimeout(2*time.Second).
			ShutdownTimeout(13*time.Second).
			Run(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		if err != nil {
			log.Fatal(err)
		}
		return

	case "api":
		services = append(services,
			pal.Provide[core.MetricsServer, metrics.HTTPServer](),
			pal.Provide[core.DB, persistence.DB](),
			pal.Provide[core.PostRepository, posts.Repository](),

			api.Provide(),
		)

	case "post-stats-collector":
		services = append(services,
			pal.Provide[core.MetricsServer, metrics.HTTPServer](),
			pal.Provide[core.DB, persistence.DB](),
			pal.Provide[core.PostRepository, posts.Repository](),

			pal.Provide[*analytics.PostStatsCollector, analytics.PostStatsCollector](),
		)

	case "migrate-up":
		services = append(
			services,
			pal.Provide[core.DB, persistence.DB](),
			pal.Provide[core.Migrator, persistence.Migrator](),
			pal.Provide[*persistence.MigrationUpRunner, persistence.MigrationUpRunner](),
		)

	case "migrate-down":
		services = append(
			services,
			pal.Provide[core.DB, persistence.DB](),
			pal.Provide[core.Migrator, persistence.Migrator](),
			pal.Provide[*persistence.MigrationDownRunner, persistence.MigrationDownRunner](),
		)

	default:
		log.Fatalf("unknown command: %s", command)
	}

	err := pal.New(services...).
		InitTimeout(2*time.Second).
		HealthCheckTimeout(1*time.Second).
		ShutdownTimeout(10*time.Second).
		Run(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if err != nil {
		log.Fatal(err)
	}
}
