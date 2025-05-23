package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"skylytics/internal/core"

	"github.com/Jeffail/gabs"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"
)

var (
	interactionsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_post_interactions_processed_total",
		Help: "The total number of processed post interactions",
	}, []string{"collection"})
)

type PostStatsCollector struct {
	Logger *slog.Logger

	PostRepo core.PostRepository
	JS       core.JetstreamClient

	Config *core.Config
}

func (p *PostStatsCollector) Init(_ context.Context) error {
	p.Logger = p.Logger.With("component", "analytics.PostStatsCollector")
	return nil
}

func (p *PostStatsCollector) Run(ctx context.Context) error {
	return p.JS.ConsumeToPipeline(ctx, p.Config.NatsStream, p.Config.NatsConsumer,
		pips.New[jetstream.Msg, any]().
			Then(
				apply.Each(func(_ context.Context, msg jetstream.Msg) error {
					event := &models.Event{}
					err := json.Unmarshal(msg.Data(), event)
					if err != nil {
						return err
					}

					err = p.processCommit(ctx, event)
					if err != nil {
						return err
					}
					interactionsProcessed.WithLabelValues(event.Commit.Collection).Inc()
				}),
			).
			Then(
				apply.Each(func(_ context.Context, msg jetstream.Msg) error {
					msg.Ack() // nolint:errcheck
					return nil
				}),
			),
	)
}

func (p *PostStatsCollector) processCommit(ctx context.Context, event *models.Event) error {
	switch event.Commit.Operation {
	case "create":
		container, err := gabs.ParseJSON(event.Commit.Record)
		if err != nil {
			return err
		}

		var iType string
		switch event.Commit.Collection {
		case "app.bsky.feed.like":
			iType = "like"
		case "app.bsky.feed.repost":
			iType = "repost"
		default:
			panic(fmt.Sprintf("unknown operation %s", event.Commit.Operation))
		}

		interaction := core.PostInteraction{
			CID:       container.Path("subject.cid").Data().(string),
			DID:       event.Did,
			Type:      iType,
			Timestamp: time.UnixMicro(event.TimeUS),
		}
		return p.PostRepo.AddInteraction(ctx, interaction)

	case "delete":
	default:
		panic(fmt.Sprintf("unknown operation %s", event.Commit.Operation))
	}

	return nil
}
