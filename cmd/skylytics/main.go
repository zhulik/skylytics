package main

import (
	"log"
	"os"
	"syscall"

	"skylytics/internal/archiving"
	inats "skylytics/internal/nats"
	"skylytics/internal/persistence"
	"skylytics/internal/persistence/accounts"
	"skylytics/internal/persistence/events"
	"skylytics/internal/updating"

	"skylytics/internal/bluesky"
	"skylytics/internal/commitanalyzer"
	"skylytics/internal/core"
	"skylytics/internal/forwarder"
	"skylytics/internal/metrics"

	"github.com/samber/do"
)

func main() {
	injector := do.New()
	defer injector.Shutdown()

	do.Provide[core.DB](injector, persistence.NewDB)
	do.Provide[core.JetstreamClient](injector, inats.NewClient)
	do.Provide[core.MetricsServer](injector, metrics.NewHTTPServer)
	do.MustInvoke[core.MetricsServer](injector)

	command := os.Args[1]
	switch command {
	case "subscriber":
		do.Provide[core.BlueskySubscriber](injector, bluesky.NewSubscriber)
		do.Provide[core.Forwarder](injector, forwarder.New)
		do.MustInvoke[core.Forwarder](injector)

	case "commit-analyzer":
		do.Provide[core.CommitAnalyzer](injector, commitanalyzer.New)
		do.MustInvoke[core.CommitAnalyzer](injector)

	case "event-archiver":
		do.Provide[core.EventRepository](injector, events.NewRepository)
		do.Provide[core.EventsArchiver](injector, archiving.NewEventsArchiver)
		do.MustInvoke[core.EventsArchiver](injector)

	case "account-updater":
		do.Provide[core.AccountRepository](injector, accounts.NewRepository)
		do.Provide[core.AccountUpdater](injector, updating.NewAccountUpdater)
		do.MustInvoke[core.AccountUpdater](injector)

	case "metrics-server":
		do.Provide[core.MetricsCollector](injector, metrics.NewCollector)
		do.MustInvoke[core.MetricsCollector](injector)

	case "migrate":
		db := do.MustInvoke[core.DB](injector)
		err := db.Migrate()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Database migrated")
		return

	default:
		log.Fatalf("unknown command: %s", command)
	}

	if err := injector.ShutdownOnSignals(syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}
}
