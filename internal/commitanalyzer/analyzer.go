package commitanalyzer

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/zhulik/pips/apply"

	"skylytics/internal/core"

	"github.com/zhulik/pips"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	commitProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_commit_processed_total",
		Help: "The total number of processed commits",
	}, []string{"commit_type", "operation"})
)

type Analyzer struct {
	JS core.JetstreamClient
}

func (a *Analyzer) Run(ctx context.Context) error {
	return a.JS.ConsumeToPipeline(ctx,
		os.Getenv("NATS_STREAM"), "commit-analyzer",
		pips.New[jetstream.Msg, any](apply.Map(a.Analyze)))
}

func (a *Analyzer) Analyze(_ context.Context, msg jetstream.Msg) (any, error) {
	msg.Ack() //nolint:errcheck

	event := core.BlueskyEvent{}
	err := json.Unmarshal(msg.Data(), &event)
	if err != nil {
		return nil, err
	}

	slog.Info("Processing", "event", event)

	commitProcessed.WithLabelValues(event.Commit.Collection, event.Commit.Operation).Inc()
	return nil, nil
}
