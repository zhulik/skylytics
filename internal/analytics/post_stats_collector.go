package analytics

import (
	"context"
	"log/slog"

	"skylytics/internal/core"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"
)

type PostStatsCollector struct {
	Logger *slog.Logger
	DB     core.DB
	JS     core.JetstreamClient

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
					msg.Ack() // nolint:errcheck
					return nil
				}),
			),
	)
}
