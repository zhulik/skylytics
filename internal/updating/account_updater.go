package updating

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/samber/do"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log"
	"net/url"
	"os"
	"resty.dev/v3"
	"skylytics/internal/core"
	"skylytics/pkg/async"
	"skylytics/pkg/stormy"
	"sync"
	"time"
)

type AccountUpdater struct {
	accountRepo core.AccountRepository
	stormy      *stormy.Client
}

func NewAccountUpdater(injector *do.Injector) (core.AccountUpdater, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL)
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

	updater := AccountUpdater{
		accountRepo: do.MustInvoke[core.AccountRepository](injector),
		stormy: stormy.NewClient(&stormy.ClientConfig{
			TransportSettings: stormy.DefaultConfig.TransportSettings,

			ResponseMiddlewares: []resty.ResponseMiddleware{func(client *resty.Client, response *resty.Response) error {
				reqURL, err := url.Parse(response.Request.URL)
				if err != nil {
					return err
				}

				log.Printf("%s %s: %s [%s]", response.Request.Method, reqURL.Path, response.Status(), response.Duration())
				return nil
			}},
		}),
	}

	go func() {
		// TODO: shutdown!
		for {

			batch, err := cons.Fetch(100)
			if err != nil {
				log.Printf("error fetching events: %+v", err)
				continue
			}

			wg := &sync.WaitGroup{}

			for msgs := range async.Batcher(context.TODO(), batch.Messages(), 25, 1*time.Second) {
				wg.Add(1)
				go func(msgs []jetstream.Msg) {
					defer wg.Done()
					err = updater.Update(context.TODO(), msgs...)
					if err != nil {
						log.Printf("error updating accounts: %+v", err)
					}
				}(msgs)
			}

			wg.Wait()
		}
	}()

	return updater, nil
}

func (a AccountUpdater) Update(ctx context.Context, msgs ...jetstream.Msg) error {
	dids, err := async.AsyncMap(ctx, msgs, func(_ context.Context, msg jetstream.Msg) (string, error) {
		var event core.BlueskyEvent

		if err := json.Unmarshal(msg.Data(), &event); err != nil {
			return "", err
		}
		return event.Did, nil
	})
	if err != nil {
		return err
	}

	dids = lo.Uniq(dids)

	profiles, err := a.stormy.GetProfiles(ctx, dids...)
	if err != nil {
		return err
	}

	serializedProfiles, err := async.AsyncMap(nil, profiles, func(_ context.Context, profile *stormy.Profile) ([]byte, error) {
		return json.Marshal(profile)
	})
	if err != nil {
		return err
	}

	_, err = a.accountRepo.InsertRaw(context.TODO(), serializedProfiles...)
	if err != nil {
		if !mongo.IsDuplicateKeyError(err) {
			return err
		}
	}

	return async.AsyncEach(nil, msgs, func(_ context.Context, msg jetstream.Msg) error {
		return msg.Ack()
	})
}

func (a AccountUpdater) HealthCheck() error {
	return nil
}

func (a AccountUpdater) Shutdown() error {
	return nil
}
