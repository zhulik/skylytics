package archiving

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do"
	"log"
	"os"
	"skylytics/internal/core"
	"skylytics/pkg/async"
	"time"
)

const (
	batchSize = 10
)

type EventsArchiver struct {
	eventRepository core.EventRepository
	handle          *async.JobHandle[any]
}

func NewEventsArchiver(injector *do.Injector) (core.EventsArchiver, error) {
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

	cons, err := js.Consumer(context.TODO(), "skylytics", "events-archiver")
	if err != nil {
		return nil, err
	}

	archiver := EventsArchiver{
		eventRepository: do.MustInvoke[core.EventRepository](injector),
	}

	handle := async.Job(func(ctx context.Context) (any, error) {
		for {
			select {
			case <-ctx.Done():
				return nil, nil
			default:
				batch, err := cons.Fetch(batchSize * 10)
				if err != nil {
					log.Printf("error fetching events: %+v", err)
					continue
				}

				for msgs := range async.Batcher(ctx, batch.Messages(), batchSize, 1*time.Second) {
					err = archiver.Archive(ctx, msgs...)
					if err != nil {
						log.Printf("error archiving events: %+v", err)
					}
				}
			}

		}
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
	return nil
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
