package updating

import (
	"context"
	"fmt"
	"net/url"
	"skylytics/internal/core"
	"skylytics/pkg/async"
	"skylytics/pkg/stormy"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/samber/do"
	"resty.dev/v3"
)

var (
	apiLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "stormy_request_latency",
			Help:    "Histogram of Stormy API request latency in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"method", "path", "status_code"},
	)
)

type AccountUpdater struct {
	accountRepo core.AccountRepository
	stormy      *stormy.Client

	handle *async.JobHandle[any]
}

func NewAccountUpdater(injector *do.Injector) (core.AccountUpdater, error) {
	updater := AccountUpdater{
		accountRepo: do.MustInvoke[core.AccountRepository](injector),
		stormy: stormy.NewClient(&stormy.ClientConfig{
			TransportSettings: stormy.DefaultConfig.TransportSettings,

			ResponseMiddlewares: []resty.ResponseMiddleware{metricMiddleware},
		}),
	}

	updater.handle = async.Job(func(ctx context.Context) (any, error) {
		js := do.MustInvoke[core.JetstreamClient](injector)

		return nil, js.ConsumeToPipeline(ctx,
			"skylytics", "account-updater",
			pipeline(&updater))
	})

	return &updater, nil
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

func (a AccountUpdater) Update(ctx context.Context, msgs ...jetstream.Msg) error {
	return async.AsyncEach(ctx, msgs, func(_ context.Context, msg jetstream.Msg) error {
		return msg.Ack()
	})
}

func (a AccountUpdater) HealthCheck() error {
	return a.handle.Error()
}

func (a AccountUpdater) Shutdown() error {
	a.handle.Stop()
	_, err := a.handle.Wait()
	return err
}
