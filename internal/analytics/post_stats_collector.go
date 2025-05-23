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
	"github.com/samber/lo"

	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"
)

var (
	interactionsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "skylytics_post_interactions_processed_total",
		Help: "The total number of processed post interactions",
	}, []string{"operation", "collection"})
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

type pipelineItem struct {
	msg    jetstream.Msg
	event  *models.Event
	record *gabs.Container

	interaction *core.PostInteraction
}

func (i *pipelineItem) Ack() {
	i.msg.Ack() // nolint:errcheck
}

func (i *pipelineItem) Nak() {
	i.msg.Nak() // nolint:errcheck
}

func (p *PostStatsCollector) Run(ctx context.Context) error {
	return p.JS.ConsumeToPipeline(ctx, p.Config.NatsStream, p.Config.NatsConsumer, p.pipeline())
}

func (p *PostStatsCollector) pipeline() *pips.Pipeline[jetstream.Msg, any] {
	return pips.New[jetstream.Msg, any]().
		Then(
			apply.Map(func(_ context.Context, msg jetstream.Msg) (pipelineItem, error) {
				event := &models.Event{}
				err := json.Unmarshal(msg.Data(), event)
				if err != nil {
					return pipelineItem{}, err
				}

				var record *gabs.Container

				if len(event.Commit.Record) > 0 {
					record, err = gabs.ParseJSON(event.Commit.Record)
					if err != nil {
						return pipelineItem{}, err
					}
				}

				return pipelineItem{
					msg:    msg,
					event:  event,
					record: record,
				}, nil
			})).
		Then(
			apply.Map(func(_ context.Context, item pipelineItem) (pipelineItem, error) {
				switch item.event.Commit.Operation {
				case "create":
					var iType string
					var cid string
					switch item.event.Commit.Collection {
					case "app.bsky.feed.like":
						iType = "like"
						cid = item.record.Path("subject.cid").Data().(string)
					case "app.bsky.feed.repost":
						iType = "repost"
						cid = item.record.Path("subject.cid").Data().(string)
					case "app.bsky.feed.post":
						iType = "reply"
						cid = item.record.Path("reply.root.cid").Data().(string)
					default:
						panic(fmt.Sprintf("unknown collection %s", item.event.Commit.Collection))
					}

					interaction := core.PostInteraction{
						CID:       cid,
						DID:       item.event.Did,
						Type:      iType,
						Timestamp: time.UnixMicro(item.event.TimeUS),
					}

					item.interaction = &interaction
					return item, nil

				case "delete":
					return item, nil
				default:
					panic(fmt.Sprintf("unknown operation %s", item.event.Commit.Operation))
				}
			})).
		Then(apply.Batch[pipelineItem](100)).
		Then(
			apply.EachC(2, func(ctx context.Context, items []pipelineItem) error {
				interactions := lo.Compact(
					lo.Map(items, func(item pipelineItem, _ int) *core.PostInteraction {
						return item.interaction
					}),
				)
				err := p.PostRepo.AddInteraction(ctx, interactions...)

				if err != nil {
					p.Logger.Warn("failed to add interactions", "error", err)
					lo.ForEach(items, func(item pipelineItem, _ int) {
						item.Nak()
					})
				}

				return nil
			}),
		).
		Then(apply.Flatten[pipelineItem]()).
		Then(
			apply.Each(func(_ context.Context, item pipelineItem) error {
				interactionsProcessed.WithLabelValues(item.event.Commit.Operation, item.event.Commit.Collection).Inc()
				item.Ack()
				return nil
			}),
		)
}
