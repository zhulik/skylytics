package archiving

import (
	"context"
	"skylytics/internal/core"
	"skylytics/pkg/async"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do"

	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"
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

	archiver.handle = async.Job(func(ctx context.Context) (any, error) {
		js := do.MustInvoke[core.JetstreamClient](injector)

		return nil, js.ConsumeToPipeline(ctx, "skylytics", "events-archiver", pips.New[jetstream.Msg, any]().
			Then(apply.Batch[jetstream.Msg](batchSize)).
			Then(
				apply.Map(func(ctx context.Context, msgs []jetstream.Msg) ([]jetstream.Msg, error) {
					return msgs, archiver.Archive(ctx, msgs...)
				}),
			))
	})

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

	if _, err := a.eventRepository.InsertRaw(ctx, events...); err != nil {
		return err
	}

	return async.AsyncEach(ctx, msgs, func(_ context.Context, item jetstream.Msg) error {
		return item.Ack()
	})
}
