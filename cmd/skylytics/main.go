package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"skylytics/internal/api"
	"syscall"
	"time"

	"skylytics/internal/bluesky"
	"skylytics/internal/core"
	"skylytics/internal/forwarder"
	"skylytics/internal/metrics"
	inats "skylytics/internal/nats"
	"skylytics/internal/persistence"
	"skylytics/internal/persistence/accounts"
	"skylytics/internal/updating"

	"github.com/zhulik/pal"
	"github.com/zhulik/pal/inspect"
)

func main() {
	services := []pal.ServiceDef{
		pal.ProvideConst[*slog.Logger](slog.New(slog.NewTextHandler(os.Stdout, nil))),
		pal.Provide[core.JetstreamClient, inats.Client](),
		pal.Provide[core.MetricsServer, metrics.HTTPServer](),
		pal.Provide[*core.Config, core.Config](),
	}

	command := os.Args[1]
	switch command {
	case "subscriber":
		services = append(services, inspect.Provide()...)
		services = append(services,
			pal.Provide[core.BlueskySubscriber, bluesky.Subscriber](),
			pal.Provide[core.Forwarder, forwarder.Forwarder](),
		)

	case "account-updater":
		services = append(services,
			pal.Provide[core.AccountRepository, accounts.Repository](),
			pal.Provide[core.AccountUpdater, updating.AccountUpdater](),
		)

	case "metrics-server":
		services = append(services,
			pal.Provide[core.DB, persistence.DB](),
			pal.Provide[core.MetricsCollector, metrics.Collector]())

	case "repl":
		err := pal.New(
			pal.ProvideRunner(inspect.RemoteConsole),
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
			pal.Provide[*api.Server, api.Server](),
		)

	case "migrate":
		services = append(
			services,
			pal.Provide[core.DB, persistence.DB](),
			pal.Provide[core.Migrator, persistence.Migrator](),
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
