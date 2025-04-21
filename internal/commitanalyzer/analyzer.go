package commitanalyzer

import (
	"context"
	"encoding/json"
	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"
	"skylytics/pkg/async"

	"skylytics/internal/core"
	inats "skylytics/internal/nats"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/samber/do"
)

var (
	commitProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_commit_processed_total",
		Help: "The total number of processed commits",
	}, []string{"commit_type"})
)

type Analyzer struct {
	handle *async.JobHandle[any]
}

func New(i *do.Injector) (core.CommitAnalyzer, error) {
	analyzer := Analyzer{}

	analyzer.handle = async.Job(func(ctx context.Context) (any, error) {
		ch, err := inats.Consume(ctx, "skylytics", "commit-analyzer", 1000)
		if err != nil {
			return nil, err
		}

		out := pips.New[jetstream.Msg, any]().
			Then(apply.Map(analyzer.Analyze)).
			Run(ctx, ch)

		for r := range out {
			if err := r.Error(); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return analyzer, nil
}

func (a Analyzer) Shutdown() error {
	a.handle.Stop()
	_, err := a.handle.Wait()
	return err
}

func (a Analyzer) HealthCheck() error {
	return a.handle.Error()
}

func (a Analyzer) Analyze(_ context.Context, msg jetstream.Msg) (any, error) {
	msg.Ack()

	event := core.BlueskyEvent{}
	err := json.Unmarshal(msg.Data(), &event)
	if err != nil {
		return false, err
	}

	var commitType = ""

	if event.Commit.Record != nil {
		var commit core.Commit
		err = json.Unmarshal(event.Commit.Record, &commit)
		if err != nil {
			return false, err
		}
		commitType = commit.Type
	}

	commitProcessed.WithLabelValues(commitType).Inc()
	return true, nil
}
