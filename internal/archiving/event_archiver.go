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

type EventsArchiver struct {
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

	go func() {
		n := 0

		// TODO: shutdown!
		for {
			batch, err := cons.Fetch(100)
			if err != nil {
				log.Printf("error fetching events: %+v", err)
				continue
			}

			log.Printf("Processing batch %d", n)

			var msgs []jetstream.Msg

			for msg := range batch.Messages() {
				if err := batch.Error(); err != nil {
					log.Printf("Error processing batch %d: %+v", n, err)
					break
				}
				msgs = append(msgs, msg)
			}

			err = archiver.Archive(msgs...)
			if err != nil {
				log.Printf("error archiving events: %+v", err)
			}
			log.Printf("Batch %d of %d elements archived", n, len(msgs))
			n++
		}
	}()

	return &archiver, nil
}

func (a EventsArchiver) Shutdown() error {
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
