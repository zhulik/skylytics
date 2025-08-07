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

type Subscriber struct {
	Logger *slog.Logger
	JS     core.JetstreamClient
	KV     core.KeyValueClient
	Config *core.Config

	ch            chan pips.D[*core.BlueskyEvent]
	watchdogTimer *time.Timer
}

func (s *Subscriber) Init(ctx context.Context) error {
	var err error

	s.ch = make(chan pips.D[*core.BlueskyEvent], 10)
	s.watchdogTimer = time.NewTimer(3 * time.Second)
	s.KV, err = s.JS.KV(ctx, s.Config.NatsStateKVBucket)
	return err
}

func (s *Subscriber) Shutdown(context.Context) error {
	close(s.ch)
	s.watchdogTimer.Stop()
	return nil
}

func (s *Subscriber) ConsumeToPipeline(ctx context.Context, pipeline *pips.Pipeline[*core.BlueskyEvent, *core.BlueskyEvent]) error {
	listenerJob := async.Job(s.listen)
	defer listenerJob.Stop()

	watchdogJob := async.Job(func(ctx context.Context) (any, error) {
		for {
			select {
			case <-ctx.Done():
				return nil, nil
			case <-s.watchdogTimer.C:
				s.Logger.Warn("Stuck subscriber detected, restarting")
				listenerJob.Stop()
				listenerJob = async.Job(s.listen)
			}
		}
	})
	defer watchdogJob.Stop()

	runnerJob := async.Job(func(ctx context.Context) (any, error) {
		for d := range pipeline.Run(ctx, s.ch) {
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

	defer runnerJob.Stop()

	return async.FirstFailed(ctx, listenerJob, runnerJob, watchdogJob)
}

func (s *Subscriber) listen(ctx context.Context) (any, error) {
	client, err := bsky.NewClient(
		&bsky.ClientConfig{
			Compress:     true,
			WebsocketURL: jetstreamURL,
			ReadTimeout:  10 * time.Second,
		}, s.Logger, sequential.NewScheduler("scheduler", s.Logger, func(_ context.Context, event *models.Event) error {
			s.watchdogTimer.Reset(3 * time.Second)
			s.ch <- pips.NewD(event)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return nil, retry.WrapWithRetry(func() error {
		for {
			lastEventTimestampBytes, err := s.KV.Get(ctx, "last_event_timestamp")
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				if !errors.Is(err, jetstream.ErrKeyNotFound) {
					return err
				}
			}

			lastEventTimestamp, err := DeserializeInt64(lastEventTimestampBytes)
			if err != nil {
				lastEventTimestamp = time.Now().Add(-3 * time.Second).UnixMicro()
			}

			cursor := time.UnixMicro(lastEventTimestamp)
			if time.Since(cursor) > time.Hour {
				s.Logger.Info("Resetting cursor as it's stale.")
				lastEventTimestamp = time.Now().Add(-3 * time.Second).UnixMicro()
			}

			err = client.ConnectAndRead(ctx, &lastEventTimestamp)
			if err != nil {
				return err
			}
		}
	}, func(_ error, _ int) bool {
		return true
	}, 10)()
}

func SerializeInt64(n int64) []byte {
	return []byte(fmt.Sprintf("%d", n))
}

func DeserializeInt64(b []byte) (int64, error) {
	return strconv.ParseInt(string(b), 10, 64)
}
