package main

import (
	"github.com/samber/do"
	"log"
	"skylytics/internal/core"
	"skylytics/internal/forwarder"
	"skylytics/internal/jetstream"
	"skylytics/internal/metrics"
	"syscall"
)

func main() {
	injector := do.New()

	do.Provide[core.MetricsServer](injector, metrics.NewHTTPServer)
	do.Provide[core.JetstreamSubscriber](injector, jetstream.NewSubscriber)
	do.Provide[core.Forwarder](injector, forwarder.New)

	do.MustInvoke[core.MetricsServer](injector)
	do.MustInvoke[core.Forwarder](injector)

	if err := injector.ShutdownOnSignals(syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}
}
