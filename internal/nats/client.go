package nats

import (
	"context"
	"log/slog"
	"time"

	"skylytics/internal/core"

	"github.com/zhulik/pips"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Client struct {
	jetstream.JetStream

	Config *core.Config
	Logger *slog.Logger
}

func (c *Client) KV(ctx context.Context, bucket string) (core.KeyValueClient, error) {
	return NewKV(ctx, c, bucket)
}

func (c *Client) Init(_ context.Context) error {
	c.Logger = c.Logger.With("component", "nats.Client")

	url := c.Config.NatsURL
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)
	if err != nil {
		return err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return err
	}

	c.JetStream = js

	return nil
}

func (c *Client) ConsumeToPipeline(ctx context.Context, stream, name string, pipeline *pips.Pipeline[jetstream.Msg, any]) error {
	cons, err := c.Consumer(ctx, stream, name)
	if err != nil {
		return err
	}

	ch := make(chan pips.D[jetstream.Msg])

	consCtx, err := cons.Consume(func(msg jetstream.Msg) {
		ch <- pips.NewD(msg)
	},
		jetstream.PullExpiry(5*time.Second),
		jetstream.PullMaxMessages(10000),
		jetstream.PullThresholdMessages(8000),
	)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()

		consCtx.Drain()
		<-consCtx.Closed()

		close(ch)
	}()

	return pipeline.
		Run(ctx, ch).
		Wait(ctx)
}

func (c *Client) HealthCheck(_ context.Context) error {
	return nil
}

func (c *Client) Shutdown(_ context.Context) error {
	return nil
}
