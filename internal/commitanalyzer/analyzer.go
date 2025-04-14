package commitanalyzer

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"skylytics/internal/core"

	"github.com/nats-io/nats.go"
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
	ctx jetstream.ConsumeContext
}

func New(i *do.Injector) (core.CommitAnalyzer, error) {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}
	cons, err := js.Consumer(context.Background(), "skylytics", "commit-analyzer")
	if err != nil {
		return nil, err
	}

	analyzer := Analyzer{}
	consCtx, err := cons.Consume(analyzer.Analyze)
	if err != nil {
		return nil, err
	}

	analyzer.ctx = consCtx

	return analyzer, nil
}

func (a Analyzer) Shutdown() error {
	a.ctx.Stop()
	return nil
}

func (a Analyzer) HealthCheck() error {
	return nil
}

func (a Analyzer) Analyze(msg jetstream.Msg) {
	msg.Ack()

	event := core.BlueskyEvent{}
	err := json.Unmarshal(msg.Data(), &event)
	if err != nil {
		log.Println("error unmarshalling event:", err)
		return
	}

	var commitType = ""

	if event.Commit.Record != nil {
		var commit core.Commit
		err = json.Unmarshal(event.Commit.Record, &commit)
		if err != nil {
			log.Printf("error parsing commit record: %+v", err)
			return
		}
		commitType = commit.Type
	}

	commitProcessed.WithLabelValues(commitType).Inc()
}
