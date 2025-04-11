package main

import (
	"go.uber.org/fx"
	"skylytics/internal/core"
	"skylytics/internal/metrics"
)

func main() {
	fx.New(
		fx.Provide(metrics.NewHTTPServer),
		fx.Invoke(func(server core.MetricsServer) {}),
	).Run()
}
