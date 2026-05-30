package cmd

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"skylytics/internal/bluesky"
	"skylytics/internal/cmd/flags"
	"skylytics/internal/config"
	"skylytics/internal/core"

	"github.com/bluesky-social/jetstream/pkg/models"

	"github.com/urfave/cli/v3"
	"github.com/zhulik/pal"
)

var subscriberCmd = &cli.Command{
	Name:  "subscriber",
	Usage: "Subscribe to the Bluesky events, forward them to NATS JetStream",
	Flags: []cli.Flag{
		flags.NATSUrl,
		flags.InitNATS,
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		return run(ctx, c,
			pal.Provide(&bluesky.Subscriber{}),
			// pal.Provide(&cursorStore{}),
			pal.Provide(&subscriber{}),
			// nats.Provide(),
		)
	},
}

type subscriber struct {
	Logger     *slog.Logger
	Config     *config.Config
	Subscriber *bluesky.Subscriber
	Metrics    core.MetricsCollector
	// NATS        *nats.NATS
	// CursorStore *cursorStore
}

func (s *subscriber) Run(ctx context.Context) error {
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
	// cursor, err := s.CursorStore.Get(ctx)
	// if err != nil {
	// 	return err
	// }

	s.Logger.Info("subscribing to the Bluesky events")
	ch, err := s.Subscriber.Consume(ctx, nil)
	if err != nil {
		return err
	}

	for event := range ch {
		// bytes, err := json.Marshal(event)
		// if err != nil {
		// 	return err
		// }

		// msgID := messageID(event)

		// msg := &libnats.Msg{
		// 	Subject: "skylytics.event",
		// 	Data:    bytes,
		// 	Header: libnats.Header{
		// 		libnats.MsgIdHdr: []string{msgID},
		// 	},
		// }
		// _, err = s.NATS.JS.PublishMsg(ctx, msg)
		// if err != nil {
		// 	return err
		// }
		// err = s.CursorStore.Set(ctx, event.TimeUS)
		// if err != nil {
		// 	return err
		// }

		s.publishEvent(ctx, event)
	}

	return nil
}

func (s *subscriber) publishEvent(ctx context.Context, event *models.Event) {
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
	// s.Logger.Debug("published event", "id", msgID, "cursor", event.TimeUS)
}
