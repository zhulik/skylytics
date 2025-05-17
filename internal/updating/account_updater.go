package updating

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"skylytics/internal/core"
	"skylytics/pkg/stormy"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"resty.dev/v3"
)

var (
	apiLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "stormy_request_latency",
			Help:    "Histogram of stormy API request latency in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"method", "path", "status_code"},
	)
)

type AccountUpdater struct {
	JS          core.JetstreamClient
	Logger      *slog.Logger
	AccountRepo core.AccountRepository
	Config      *core.Config

	stormy *stormy.Client
}

func (a *AccountUpdater) Init(_ context.Context) error {
	a.stormy = stormy.NewClient(&stormy.ClientConfig{
		TransportSettings: stormy.DefaultConfig.TransportSettings,

		ResponseMiddlewares: []resty.ResponseMiddleware{metricMiddleware},
		RequestMiddlewares:  []resty.RequestMiddleware{},
	})

	return nil
}

func (a *AccountUpdater) Run(ctx context.Context) error {
	go func() {
		timer := time.NewTicker(2 * time.Second)
		defer timer.Stop()

		var prev int64

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				total := eventsProcessedCounter.Load()
				a.Logger.Info("Events processed", "total", total, "diff", total-prev)
				prev = total
			}
		}
	}()

	return a.JS.ConsumeToPipeline(
		ctx,
		a.Config.NatsStream,
		a.Config.NatsConsumer,
		pipeline(a),
	)
}

func metricMiddleware(_ *resty.Client, response *resty.Response) error {
	reqURL, err := url.Parse(response.Request.URL)
	if err != nil {
		return err
	}

	statusCode := response.StatusCode()
	apiLatency.WithLabelValues(
		response.Request.Method,
		reqURL.Path,
		fmt.Sprintf("%d", statusCode),
	).Observe(response.Duration().Seconds())

	return nil
}
