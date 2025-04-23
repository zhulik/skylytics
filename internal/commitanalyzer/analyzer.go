package commitanalyzer

import (
	"context"
	"encoding/json"

	"skylytics/pkg/async"

	"github.com/zhulik/pips/apply"

	"skylytics/internal/core"

	"github.com/zhulik/pips"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/samber/do"
)

var (
	commitProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_commit_processed_total",
		Help: "The total number of processed commits",
	}, []string{"commit_type", "operation"})
)

type Analyzer struct {
	handle *async.JobHandle[any]
}

func New(i *do.Injector) (core.CommitAnalyzer, error) {
	analyzer := Analyzer{}

	analyzer.handle = async.Job(func(ctx context.Context) (any, error) {
		js := do.MustInvoke[core.JetstreamClient](i)
		return nil, js.ConsumeToPipeline(ctx,
			"skylytics", "commit-analyzer",
			pips.New[jetstream.Msg, any](apply.Map(analyzer.Analyze)))
	})

	return &analyzer, nil
}

func (a Analyzer) Shutdown() error {
	_, err := a.handle.StopWait()
	return err
}

func (a Analyzer) HealthCheck() error {
	return a.handle.Error()
}

func (a Analyzer) Analyze(_ context.Context, msg jetstream.Msg) (any, error) {
	msg.Ack() //nolint:errcheck

	event := core.BlueskyEvent{}
	err := json.Unmarshal(msg.Data(), &event)
	if err != nil {
		return nil, err
	}

	var commitType = ""
	var commitOperation = ""

	if event.Commit.Record != nil {
		commitType = event.Commit.Collection
		commitOperation = event.Commit.Operation
	}

	commitProcessed.WithLabelValues(commitType, commitOperation).Inc()
	return nil, nil
}
