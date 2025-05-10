package nats

import (
	"context"
	"os"

	"skylytics/internal/core"
	"skylytics/pkg/async"

	"github.com/zhulik/pips"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Client struct {
	jetstream.JetStream
	Handle *async.JobHandle[any]
}

func (c *Client) KV(ctx context.Context, bucket string) (core.KeyValueClient, error) {
	return NewKV(ctx, c, bucket)
}

func (c *Client) Init(_ context.Context) error {
	url := os.Getenv("NATS_URL")
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
	ch, err := c.Consume(ctx, stream, name)
	if err != nil {
		return err
	}

	return pipeline.
		Run(ctx, ch).
		Wait(ctx)
}

func (c *Client) Consume(ctx context.Context, stream, name string) (<-chan pips.D[jetstream.Msg], error) {
	cons, err := c.Consumer(ctx, stream, name)
	if err != nil {
		return nil, err
	}

	var ch <-chan pips.D[jetstream.Msg]

	c.Handle, ch = async.Generator(func(ctx context.Context, y async.Yielder[jetstream.Msg]) error {
		for {
			select {
			case <-ctx.Done():
				return nil

			default:
				batch, err := cons.FetchNoWait(100)
				if err != nil {
					y(nil, err)
				}

				if batch.Error() != nil {
					return batch.Error()
				}

				for msg := range batch.Messages() {
					y(msg, nil)
				}
			}
		}
	})

	return ch, nil
}

func (c *Client) HealthCheck(_ context.Context) error {
	if c.Handle == nil {
		return nil
	}
	return c.Handle.Error()
}

func (c *Client) Shutdown(_ context.Context) error {
	if c.Handle == nil {
		return nil
	}
	_, err := c.Handle.StopWait()
	return err
}
