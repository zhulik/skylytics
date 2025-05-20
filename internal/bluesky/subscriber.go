package bluesky

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"

	"skylytics/pkg/retry"

	"skylytics/internal/core"

	bsky "github.com/bluesky-social/jetstream/pkg/client"
	"github.com/bluesky-social/jetstream/pkg/client/schedulers/sequential"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/zhulik/pips"
)

const (
	jetstreamURL = "wss://jetstream2.us-east.bsky.network/subscribe"
)

type Subscriber struct {
	Logger *slog.Logger
	JS     core.JetstreamClient
	KV     core.KeyValueClient
	Config *core.Config
}

func (s *Subscriber) Init(ctx context.Context) error {
	var err error

	s.Logger = s.Logger.With("component", "bluesky.Subscriber")

	s.KV, err = s.JS.KV(ctx, s.Config.NatsStateKVBucket)
	return err
}

func (s *Subscriber) ConsumeToPipeline(ctx context.Context, pipeline *pips.Pipeline[*core.BlueskyEvent, *core.BlueskyEvent]) error {
	ch := make(chan pips.D[*core.BlueskyEvent])
	defer close(ch)

	handler := sequential.NewScheduler("scheduler", s.Logger, func(_ context.Context, event *models.Event) error {
		ch <- pips.NewD(event)
		return nil
	})

	client, err := bsky.NewClient(
		&bsky.ClientConfig{
			Compress:     true,
			WebsocketURL: jetstreamURL,
			ReadTimeout:  10 * time.Second,
		}, s.Logger, handler,
	)

	if err != nil {
		return err
	}

	wg := errgroup.Group{}

	wg.Go(retry.WrapWithRetry(func() error {
		lastEventTimestampBytes, err := s.KV.Get(ctx, "last_event_timestamp")
		if err != nil {
			if !errors.Is(err, jetstream.ErrKeyNotFound) {
				return err
			}
		}

		lastEventTimestamp, err := DeserializeInt64(lastEventTimestampBytes)
		if err != nil {
			lastEventTimestamp = 0
		}

		return client.ConnectAndRead(ctx, &lastEventTimestamp)
	}, func(_ error, _ int) bool {
		return true
	}, 10))

	wg.Go(func() error {
		for d := range pipeline.Run(ctx, ch) {
			event, err := d.Unpack()
			if err != nil {
				return err
			}
			err = s.KV.Put(ctx, "last_event_timestamp", SerializeInt64(event.TimeUS))
			if err != nil {
				return err
			}
		}
		return nil
	})

	return wg.Wait()
}

func SerializeInt64(n int64) []byte {
	return []byte(fmt.Sprintf("%d", n))
}

func DeserializeInt64(b []byte) (int64, error) {
	return strconv.ParseInt(string(b), 10, 64)
}
