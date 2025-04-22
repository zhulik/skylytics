package nats

import (
	"context"
	"os"

	"skylytics/internal/core"
	"skylytics/pkg/async"

	"github.com/zhulik/pips"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do"
)

type Client struct {
	jetstream.JetStream
}

func NewClient(_ *do.Injector) (core.JetstreamClient, error) {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	c, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	return Client{c}, nil
}

func (c Client) ConsumeToPipeline(ctx context.Context, stream, name string, pipeline *pips.Pipeline[jetstream.Msg, any]) error {
	ch, err := c.Consume(ctx, stream, name)
	if err != nil {
		return err
	}

	return pipeline.
		Run(ctx, ch).
		Wait(ctx)
}

func (c Client) Consume(ctx context.Context, stream, name string) (<-chan pips.D[jetstream.Msg], error) {
	cons, err := c.Consumer(ctx, stream, name)
	if err != nil {
		return nil, err
	}

	return async.Generator(ctx, func(ctx context.Context, y async.Yielder[jetstream.Msg]) error {
		for {
			select {
			case <-ctx.Done():
				return nil

			default:
				batch, err := cons.Fetch(1000)
				if err != nil {
					y(nil, err)
				}

				for msg := range batch.Messages() {
					y(msg, nil)
				}
			}
		}
	}), nil
}

func (c Client) HealthCheck() error {
	return nil
}

func (c Client) Shutdown() error {
	c.Conn().Close()
	return nil
}
