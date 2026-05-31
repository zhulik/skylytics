package metrics

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
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

func (c *Collector) Init(_ context.Context) error {
	c.reg = prometheus.NewRegistry()
	c.reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			Namespace: "skylytics",
		}),

		jetstreamProcessedEventsTotal,
		jetstreamSubscriptionErrorsTotal,
		blueskyPostCreatedTotal,
		blueskyPostCreatedInLanguageTotal,
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

func (c *Collector) RunConfig() *pal.RunConfig {
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

func (c *Collector) IncJetstreamProcessedEventsTotal(_ context.Context, kind, operation, collection string) {
	jetstreamProcessedEventsTotal.WithLabelValues(kind, operation, collection).Inc()
}

func (c *Collector) IncJetstreamSubscriptionErrorsTotal(_ context.Context, err error) {
	jetstreamSubscriptionErrorsTotal.WithLabelValues(err.Error()).Inc()
}

func (c *Collector) IncBlueskyPostCreated(_ context.Context, languageCount, imageCount int) {
	blueskyPostCreatedTotal.WithLabelValues(
		strconv.Itoa(languageCount),
		strconv.Itoa(imageCount),
	).Inc()
}

func (c *Collector) IncBlueskyPostCreatedInLanguage(_ context.Context, language string) {
	blueskyPostCreatedInLanguageTotal.WithLabelValues(language).Inc()
}
