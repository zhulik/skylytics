package nats

import (
	"context"

	"skylytics/pkg/async"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do"
	"github.com/zhulik/pips"
)

func Consume(ctx context.Context, i *do.Injector, stream, name string, batchSize int) (<-chan pips.D[jetstream.Msg], error) {
	js := do.MustInvoke[jetstream.JetStream](i)

	cons, err := js.Consumer(ctx, stream, name)
	if err != nil {
		return nil, err
	}

	return async.Generator(ctx, func(ctx context.Context, y async.Yielder[jetstream.Msg]) error {
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
	}), nil
}
