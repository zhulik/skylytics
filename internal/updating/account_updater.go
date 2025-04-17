package updating

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

type AccountUpdater struct {
}

func NewAccountUpdater(_ *do.Injector) (core.AccountUpdater, error) {
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

	cons, err := js.Consumer(context.TODO(), "skylytics", "account-updater")
	if err != nil {
		return nil, err
	}

	updater := AccountUpdater{}

	go func() {
		// TODO: shutdown!
		for {
			batch, err := cons.Fetch(1000)
			if err != nil {
				log.Printf("error fetching events: %+v", err)
				continue
			}

			for msgs := range async.Batcher(context.TODO(), batch.Messages(), 25, 1*time.Second) {
				err = updater.Update(msgs...)
				if err != nil {
					log.Printf("error updateing accounts: %+v", err)
					continue
				}
			}
		}
	}()

	return updater, nil
}

func (a AccountUpdater) Update(msgs ...jetstream.Msg) error {
	async.Each(msgs, func(msg jetstream.Msg) {
		msg.Ack()
	})
	// TODO: fetch account from bluesky, update it in mongo
	return nil

}

func (a AccountUpdater) HealthCheck() error {
	return nil
}

func (a AccountUpdater) Shutdown() error {
	return nil
}
