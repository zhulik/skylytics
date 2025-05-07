package archiving

import (
	"context"
	"encoding/json"
	"os"

	"skylytics/internal/core"
	"skylytics/pkg/async"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/zhulik/pips"
	"github.com/zhulik/pips/apply"
)

const (
	batchSize = 10
)

type EventsArchiver struct {
	JS              core.JetstreamClient
	EventRepository core.EventRepository
}

func (a *EventsArchiver) Run(ctx context.Context) error {
	return a.JS.ConsumeToPipeline(ctx, os.Getenv("NATS_STREAM"), "events-archiver", pips.New[jetstream.Msg, any]().
		Then(apply.Batch[jetstream.Msg](batchSize)).
		Then(
			apply.Map(func(ctx context.Context, msgs []jetstream.Msg) ([]jetstream.Msg, error) {
				return msgs, a.Archive(ctx, msgs...)
			}),
		))
}

func (a *EventsArchiver) Archive(ctx context.Context, msgs ...jetstream.Msg) error {
	events, err := async.AsyncMap(ctx, msgs, func(_ context.Context, item jetstream.Msg) (core.EventModel, error) {
		var event core.BlueskyEvent

		err := json.Unmarshal(item.Data(), &event)
		return core.EventModel{Event: event}, err
	})
	if err != nil {
		return err
	}

	if err := a.EventRepository.Insert(ctx, events...); err != nil {
		return err
	}

	return async.AsyncEach(ctx, msgs, func(_ context.Context, item jetstream.Msg) error {
		return item.Ack()
	})
}
