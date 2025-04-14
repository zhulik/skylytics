package main

import (
	"log"
	"os"
	"syscall"

	"skylytics/internal/bluesky"
	"skylytics/internal/commitanalyzer"
	"skylytics/internal/core"
	"skylytics/internal/forwarder"
	"skylytics/internal/metrics"

	"github.com/samber/do"
)

func main() {
	injector := do.New()

	do.Provide[core.MetricsServer](injector, metrics.NewHTTPServer)
	do.MustInvoke[core.MetricsServer](injector)

	command := "subscriber"

	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "subscriber":
		do.Provide[core.BlueskySubscriber](injector, bluesky.NewSubscriber)
		do.Provide[core.Forwarder](injector, forwarder.New)

		do.MustInvoke[core.Forwarder](injector)
	case "commit-analyzer":
		do.Provide[core.CommitAnalyzer](injector, commitanalyzer.New)

		do.MustInvoke[core.CommitAnalyzer](injector)
	default:
		log.Fatalf("unknown command: %s", command)
	}

	if err := injector.ShutdownOnSignals(syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}
}
