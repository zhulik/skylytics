package archiving

import (
	"context"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do"
	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"
	"skylytics/internal/core"
	inats "skylytics/internal/nats"
	"skylytics/pkg/async"
)

const (
	batchSize = 10
)

type EventsArchiver struct {
	eventRepository core.EventRepository
	handle          *async.JobHandle[any]
}

func NewEventsArchiver(injector *do.Injector) (core.EventsArchiver, error) {
	archiver := EventsArchiver{
		eventRepository: do.MustInvoke[core.EventRepository](injector),
	}

	handle := async.Job(func(ctx context.Context) (any, error) {
		ch, err := inats.Consume(ctx, "skylytics", "events-archiver", batchSize*10)
		if err != nil {
			return nil, err
		}

		input := pips.MapInputChan(ctx, ch, func(ctx context.Context, a async.Result[jetstream.Msg]) (pips.D[jetstream.Msg], error) {
			msg, err := a.Unpack()
			if err != nil {
				return nil, err
			}
			return pips.NewD(msg), nil
		})

		pips.New[jetstream.Msg, any]().
			Then(apply.Batch(batchSize)).
			Then(apply.Map(func(ctx context.Context, msgs []jetstream.Msg) (any, error) {
				return nil, archiver.Archive(ctx, msgs...)
			})).Run(ctx, input)

		return nil, nil
	})

	archiver.handle = handle

	return &archiver, nil
}

func (a EventsArchiver) Shutdown() error {
	a.handle.Stop()
	_, err := a.handle.Wait()
	return err
}

func (a EventsArchiver) HealthCheck() error {
	return a.handle.Error()
}

func (a EventsArchiver) Archive(ctx context.Context, msgs ...jetstream.Msg) error {
	events := async.Map(msgs, func(item jetstream.Msg) []byte {
		return item.Data()
	})

	if _, err := a.eventRepository.InsertRaw(context.TODO(), events...); err != nil {
		return err
	}

	return async.AsyncEach(ctx, msgs, func(_ context.Context, item jetstream.Msg) error {
		return item.Ack()
	})
}
