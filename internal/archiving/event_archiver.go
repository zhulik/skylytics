package archiving

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do"
	"github.com/samber/lo"
	"log"
	"os"
	"skylytics/internal/core"
)

const (
	batchSize = 1000
)

type EventsArchiver struct {
	ctx             jetstream.ConsumeContext
	eventRepository core.EventRepository
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

	cons, err := js.Consumer(context.Background(), "skylytics", "events-archiver")
	if err != nil {
		return nil, err
	}

	archiver := EventsArchiver{
		eventRepository: do.MustInvoke[core.EventRepository](injector),
	}

	consCtx, err := cons.Consume(func(msg jetstream.Msg) {
		err = archiver.Archive(msg)
		if err != nil {
			log.Printf("error archiving event: %+v", err)
		}
	})

	archiver.ctx = consCtx

	return &archiver, nil
}

func (a EventsArchiver) Shutdown() error {
	a.ctx.Stop()
	return nil
}

func (a EventsArchiver) HealthCheck() error {
	return nil
}

func (a EventsArchiver) Archive(msgs ...jetstream.Msg) error {
	events := lo.Map(msgs, func(item jetstream.Msg, _ int) []byte {
		return item.Data()
	})

	if err := a.eventRepository.SaveRaw(context.TODO(), events...); err != nil {
		return err
	}

	lo.ForEach(msgs, func(item jetstream.Msg, _ int) {
		item.Ack()
	})

	return nil
}
