package metrics

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"skylytics/internal/core"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zhulik/pal"
)

const addr = ":2112"

type Collector struct {
	Logger *slog.Logger

	reg    *prometheus.Registry
	server *http.Server
}

func (c *Collector) Increment(_ context.Context, name core.MetricName, tags map[string]string) {
	counters[name].With(tags).Inc()
}

func (c *Collector) Init(_ context.Context) error {
	c.reg = prometheus.NewRegistry()
	c.reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		counters[JetstreamProcessedEventsTotal],
	)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(c.reg, promhttp.HandlerOpts{}))

	c.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	return nil
}

func (c *Collector) RunConfig(_ context.Context) *pal.RunConfig {
	return &pal.RunConfig{
		Wait: false,
	}
}

func (c *Collector) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		err := c.server.Shutdown(context.WithoutCancel(ctx))
		if err != nil {
			c.Logger.Error("error shutting down metrics server", "error", err)
		}
	}()

	c.Logger.Info("metrics server started", "addr", addr)

	if err := c.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}

	return nil
}
