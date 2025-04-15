package archiving

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do"
	"log"
	"os"
	"skylytics/internal/core"
)

type EventsArchiver struct {
	ctx jetstream.ConsumeContext

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

	consCtx, err := cons.Consume(archiver.Archive)
	if err != nil {
		return nil, err
	}

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

func (a EventsArchiver) Archive(msg jetstream.Msg) {
	msg.Ack()

	log.Println("Archiving event")

	if err := a.eventRepository.SaveRaw(context.TODO(), msg.Data()); err != nil {
		log.Printf("error saving event: %+v", err)
	}
}
