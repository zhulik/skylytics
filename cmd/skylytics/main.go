package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"skylytics/internal/archiving"
	"skylytics/internal/bluesky"
	"skylytics/internal/commitanalyzer"
	"skylytics/internal/core"
	"skylytics/internal/forwarder"
	"skylytics/internal/metrics"
	inats "skylytics/internal/nats"
	"skylytics/internal/persistence"
	"skylytics/internal/persistence/accounts"
	"skylytics/internal/persistence/events"
	"skylytics/internal/updating"

	"github.com/zhulik/pal"
	"golang.org/x/exp/slog"
)

func main() {
	services := []pal.ServiceImpl{
		pal.Provide[core.JetstreamClient, inats.Client](),
		pal.Provide[core.MetricsServer, metrics.HTTPServer](),
	}

	command := os.Args[1]
	switch command {
	case "subscriber":
		services = append(services,

			pal.Provide[core.BlueskySubscriber, bluesky.Subscriber](),
			pal.Provide[core.Forwarder, forwarder.Forwarder](),
		)

	case "commit-analyzer":
		services = append(services,
			pal.Provide[core.CommitAnalyzer, commitanalyzer.Analyzer](),
		)

	case "event-archiver":
		services = append(services,
			pal.Provide[core.EventRepository, events.Repository](),
			pal.Provide[core.EventsArchiver, archiving.EventsArchiver](),
		)

	case "account-updater":
		services = append(services,
			pal.Provide[core.DB, persistence.DB](),
			pal.Provide[core.AccountRepository, accounts.Repository](),
			pal.Provide[core.AccountUpdater, updating.AccountUpdater](),
		)

	case "metrics-server":
		services = append(services, pal.Provide[core.MetricsCollector, metrics.Collector]())

	case "migrate":
		// TODO: extract migration runner.
		// db := do.MustInvoke[core.DB](injector)
		// err := db.Migrate()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// log.Println("Database migrated")
		// return

	default:
		log.Fatalf("unknown command: %s", command)
	}

	err := pal.New(services...).
		SetLogger(func(fm string, args ...any) {
			slog.With("component", "pal").Info(fmt.Sprintf(fm, args...))
		}).
		InitTimeout(300*time.Second).
		HealthCheckTimeout(1*time.Second).
		ShutdownTimeout(3*time.Second).
		Run(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if err != nil {
		log.Fatal(err)
	}
}
