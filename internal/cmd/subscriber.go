package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync/atomic"
	"time"

	"skylytics/internal/bluesky"
	"skylytics/internal/cmd/flags"
	"skylytics/internal/config"
	"skylytics/internal/core"

	"github.com/bluesky-social/jetstream/pkg/models"

	libredis "github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v3"
	"github.com/zhulik/pal"
)

var subscriberCmd = &cli.Command{
	Name:  "subscriber",
	Usage: "Subscribe to the Bluesky events, forward them to NATS JetStream",
	Flags: []cli.Flag{
		flags.RedisAddr,
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		return run(ctx, c,
			pal.Provide(&bluesky.Subscriber{}),
			pal.Provide(&subscriber{}),
		)
	},
}

type subscriber struct {
	Logger     *slog.Logger
	Config     *config.Config
	Subscriber *bluesky.Subscriber
	Metrics    core.MetricsCollector

	Redis core.Redis

	processedEvents atomic.Uint64
}

func (s *subscriber) Run(ctx context.Context) error {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			s.Logger.Info("processed events", "count", s.processedEvents.Load())
		}
	}()

	for {
		err := s.run(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			s.Logger.Error("error running subscriber, retrying in 1 second", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}
		return nil
	}
}

func (s *subscriber) run(ctx context.Context) error {
	cursor, err := s.getCursor(ctx)
	if err != nil {
		return err
	}
	if cursor == nil {
		s.Logger.Info("no cursor found in Redis, starting from the beginning")
	}

	s.Logger.Info("subscribing to the Bluesky events", "cursor", cursor)
	ch, err := s.Subscriber.Consume(ctx, cursor)
	if err != nil {
		return err
	}

	for eventRes := range ch {
		event, err := eventRes.Get()
		if err != nil {
			return err
		}

		err = s.publishEvent(ctx, event)
		if err != nil {
			return err
		}
		s.processedEvents.Add(1)
		err = s.setCursor(ctx, event.TimeUS)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *subscriber) publishEvent(ctx context.Context, event *models.Event) error {
	kind := event.Kind

	tags := map[string]string{
		"operation":  "",
		"collection": "",
		"kind":       kind,
	}

	switch kind {
	case models.EventKindCommit:
		tags["operation"] = event.Commit.Operation
		tags["collection"] = event.Commit.Collection
	case models.EventKindAccount:
	case models.EventKindIdentity:
	}

	s.Metrics.Increment(ctx, "jetstream_processed_events_total", tags)

	return nil
}

func (s *subscriber) getCursor(ctx context.Context) (*int64, error) {
	cursor, err := s.Redis.Get(ctx, "cursor").Result()
	if err != nil {
		if errors.Is(err, libredis.Nil) {
			return nil, nil
		}

		return nil, fmt.Errorf("error getting cursor from Redis: %w", err)
	}

	var cursorValueInt *int64
	if cursor != "" {
		c, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing cursor: %w", err)
		}
		cursorValueInt = &c
	}

	return cursorValueInt, nil
}

func (s *subscriber) setCursor(ctx context.Context, cursor int64) error {
	return s.Redis.Set(ctx, "cursor", cursor, 0).Err()
}
