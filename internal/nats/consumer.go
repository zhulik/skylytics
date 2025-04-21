package nats

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/zhulik/pips"
	"os"
	"skylytics/pkg/async"
)

func Consume(ctx context.Context, stream, name string, batchSize int) (<-chan pips.D[jetstream.Msg], error) {
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

	cons, err := js.Consumer(ctx, stream, name)
	if err != nil {
		return nil, err
	}

	msgs := async.Generator(ctx, func(ctx context.Context, y async.Yielder[jetstream.Msg]) error {
		for {
			select {
			case <-ctx.Done():
				return nil

			default:
				batch, err := cons.Fetch(batchSize)
				if err != nil {
					y(nil, err)
				}

				for msg := range batch.Messages() {
					y(msg, nil)
				}
			}
		}
	})

	return msgs, nil
}
