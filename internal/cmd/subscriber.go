package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"skylytics/internal/bluesky"
	"skylytics/internal/cmd/flags"
	"skylytics/internal/config"
	"skylytics/internal/nats"

	"github.com/bluesky-social/jetstream/pkg/models"

	libnats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
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
			pal.Provide(&cursorStore{}),
			pal.Provide(&subscriber{}),
			nats.Provide(),
		)
	},
}

type subscriber struct {
	Logger      *slog.Logger
	Config      *config.Config
	Subscriber  *bluesky.Subscriber
	NATS        *nats.NATS
	CursorStore *cursorStore
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
	cursor, err := s.CursorStore.Get(ctx)
	if err != nil {
		return err
	}

	s.Logger.Info("subscribing to the Bluesky events")
	ch, err := s.Subscriber.Consume(ctx, cursor)
	if err != nil {
		return err
	}

	for event := range ch {
		bytes, err := json.Marshal(event)
		if err != nil {
			return err
		}

		msgID := messageID(event)

		msg := &libnats.Msg{
			Subject: "skylytics.event",
			Data:    bytes,
			Header: libnats.Header{
				libnats.MsgIdHdr: []string{msgID},
			},
		}
		_, err = s.NATS.JS.PublishMsg(ctx, msg)
		if err != nil {
			return err
		}
		err = s.CursorStore.Set(ctx, event.TimeUS)
		if err != nil {
			return err
		}

		s.Logger.Debug("published event", "id", msgID, "cursor", event.TimeUS)
	}

	return nil
}

func messageID(event *models.Event) string {
	return fmt.Sprintf("%s-%d", event.Did, event.TimeUS)
}

type cursorStore struct {
	NATS *nats.NATS
}

func (c *cursorStore) Get(ctx context.Context) (*int64, error) {
	cursor, err := c.NATS.KV.Get(ctx, "cursor")
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, err
	}
	parsed, err := strconv.ParseInt(string(cursor.Value()), 10, 64)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (c *cursorStore) Set(ctx context.Context, cursor int64) error {
	_, err := c.NATS.KV.Put(ctx, "cursor", []byte(fmt.Sprintf("%d", cursor)))
	return err
}
