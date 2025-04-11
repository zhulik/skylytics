package main

import (
	"go.uber.org/fx"
	"skylytics/internal/core"
	"skylytics/internal/forwarder"
	"skylytics/internal/jetstream"
	"skylytics/internal/metrics"
)

func main() {
	fx.New(
		fx.Provide(metrics.NewHTTPServer),
		fx.Provide(jetstream.NewSubscriber),
		fx.Provide(forwarder.New),

		fx.Invoke(func(server core.MetricsServer) {}),
		fx.Invoke(func(forwarder core.Forwarder) {}),
	).Run()
}
