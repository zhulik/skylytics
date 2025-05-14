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
	AccountRepo core.AccountRepository
	Config      *core.Config

	stormy *stormy.Client
}

func (a *AccountUpdater) Init(_ context.Context) error {
	a.stormy = stormy.NewClient(&stormy.ClientConfig{
		TransportSettings: stormy.DefaultConfig.TransportSettings,

		ResponseMiddlewares: []resty.ResponseMiddleware{metricMiddleware},
	})

	return nil
}

func (a *AccountUpdater) Run(ctx context.Context) error {
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

func (a *AccountUpdater) Update(ctx context.Context, msgs ...jetstream.Msg) error {
	return async.AsyncEach(ctx, msgs, func(_ context.Context, msg jetstream.Msg) error {
		return msg.Ack()
	})
}
