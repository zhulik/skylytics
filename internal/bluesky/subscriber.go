package bluesky

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"skylytics/pkg/async"

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

var (
	ErrStuckSubscriber = errors.New("subscriber is stuck")
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

	watchdogTimer := time.NewTimer(3 * time.Second)
	defer watchdogTimer.Stop()

	handler := sequential.NewScheduler("scheduler", s.Logger, func(_ context.Context, event *models.Event) error {
		watchdogTimer.Reset(3 * time.Second)
		ch <- pips.NewD(event)
		return nil
	})

	client, err := bsky.NewClient(
		&bsky.ClientConfig{
			Compress:     true,
			WebsocketURL: jetstreamURL,
			ReadTimeout:  10 * time.Second,
			MaxSize:      1000,
		}, s.Logger, handler,
	)

	if err != nil {
		return err
	}

	watchdogJob := async.Job[any](func(ctx context.Context) (any, error) {
		for {
			select {
			case <-ctx.Done():
				return nil, nil
			case <-watchdogTimer.C:
				s.Logger.Warn("Stuck subscriber detected")
				return nil, ErrStuckSubscriber
			}
		}
	})

	defer watchdogJob.Stop()

	listenerJob := async.Job[any](func(ctx context.Context) (any, error) {
		return nil, retry.WrapWithRetry(func() error {
			for {
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

				err = client.ConnectAndRead(ctx, &lastEventTimestamp)
				if err != nil {
					return err
				}
			}
		}, func(_ error, _ int) bool {
			return true
		}, 10)()
	})

	defer func() {
		go func() {
			listenerJob.StopWait() // nolint:errcheck
			close(ch)
		}()
	}()

	cursorUpdaterJob := async.Job[any](func(ctx context.Context) (any, error) {
		for d := range pipeline.Run(ctx, ch) {
			event, err := d.Unpack()
			if err != nil {
				return nil, err
			}
			err = s.KV.Put(ctx, "last_event_timestamp", SerializeInt64(event.TimeUS))
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})

	defer cursorUpdaterJob.Stop()

	err = async.FirstFailed(ctx, listenerJob, cursorUpdaterJob, watchdogJob)

	return err
}

func SerializeInt64(n int64) []byte {
	return []byte(fmt.Sprintf("%d", n))
}

func DeserializeInt64(b []byte) (int64, error) {
	return strconv.ParseInt(string(b), 10, 64)
}
