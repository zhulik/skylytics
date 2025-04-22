package archiving

import (
	"context"
	"encoding/json"

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
	_, err := a.handle.StopWait()
	return err
}

func (a EventsArchiver) HealthCheck() error {
	return a.handle.Error()
}

func (a EventsArchiver) Archive(ctx context.Context, msgs ...jetstream.Msg) error {
	events, err := async.AsyncMap(ctx, msgs, func(_ context.Context, item jetstream.Msg) (core.EventModel, error) {
		var event core.BlueskyEvent

		err := json.Unmarshal(item.Data(), &event)
		return core.EventModel{Event: event}, err
	})
	if err != nil {
		return err
	}

	if err := a.eventRepository.Insert(ctx, events...); err != nil {
		return err
	}

	return async.AsyncEach(ctx, msgs, func(_ context.Context, item jetstream.Msg) error {
		return item.Ack()
	})
}
