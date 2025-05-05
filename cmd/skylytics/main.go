package main

import (
	"context"
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
)

func main() {
	services := []pal.ServiceImpl{
		pal.Provide[core.DB, persistence.DB](),
		pal.Provide[core.JetstreamClient, inats.Client](),
		pal.Provide[core.MetricsServer, metrics.HTTPServer](),
		pal.Provide[core.BlueskySubscriber, bluesky.Subscriber](),
		pal.Provide[core.EventRepository, events.Repository](),
		pal.Provide[core.AccountRepository, accounts.Repository](),
	}

	command := os.Args[1]
	switch command {
	case "subscriber":
		services = append(services, pal.Provide[core.Forwarder, forwarder.Forwarder]())

	case "commit-analyzer":
		services = append(services, pal.Provide[core.CommitAnalyzer, commitanalyzer.Analyzer]())

	case "event-archiver":
		services = append(services, pal.Provide[core.EventsArchiver, archiving.EventsArchiver]())

	case "account-updater":
		services = append(services, pal.Provide[core.AccountUpdater, updating.AccountUpdater]())

	case "metrics-server":
		services = append(services, pal.Provide[core.MetricsServer, metrics.Collector]())

	case "migrate":
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

	err := pal.New(
		services...,
	).
		InitTimeout(3*time.Second).
		HealthCheckTimeout(1*time.Second).
		ShutdownTimeout(3*time.Second).Run(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if err != nil {
		log.Fatal(err)
	}
}
