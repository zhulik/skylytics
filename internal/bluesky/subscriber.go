package bluesky

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	bsky "github.com/bluesky-social/jetstream/pkg/client"
	"github.com/bluesky-social/jetstream/pkg/client/schedulers/sequential"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/nats-io/nats.go/jetstream"

	"skylytics/internal/core"

	"github.com/zhulik/pips"
)

const (
	jetstreamURL = "wss://jetstream2.us-east.bsky.network/subscribe"
)

type Subscriber struct {
	Logger *slog.Logger

	JS core.JetstreamClient

	client *bsky.Client

	KV core.KeyValueClient
	ch chan pips.D[*core.BlueskyEvent]
}

func (s *Subscriber) Init(ctx context.Context) error {
	var err error

	s.ch = make(chan pips.D[*core.BlueskyEvent])
	s.Logger = s.Logger.With("component", "bluesky.Subscriber")
	s.KV, err = s.JS.KV(ctx, os.Getenv("NATS_STATE_KV_BUCKET"))
	if err != nil {
		return err
	}

	s.client, err = bsky.NewClient(
		&bsky.ClientConfig{
			Compress:     true,
			WebsocketURL: jetstreamURL,
			ExtraHeaders: map[string]string{},
		},
		s.Logger.With("component", "bsky-jetstream-client"),
		sequential.NewScheduler("scheduler", s.Logger, func(_ context.Context, event *models.Event) error {
			s.ch <- pips.NewD(event)

			return nil
		}),
	)

	return err
}

func (s *Subscriber) Shutdown(_ context.Context) error {
	defer close(s.ch)
	return nil
}

func (s *Subscriber) C() <-chan pips.D[*core.BlueskyEvent] {
	return s.ch
}

func (s *Subscriber) Run(ctx context.Context) error {
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

	cursor := &lastEventTimestamp

	err = s.client.ConnectAndRead(ctx, cursor)

	// A separate context because the original one will be canceled for shutdown.
	putCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return errors.Join(err, s.KV.Put(putCtx, "last_event_timestamp", SerializeInt64(*cursor)))

	//timer := time.NewTimer(5 * time.Second)
	//
	//defer timer.Stop()
	//
	//go func() {
	//	<-timer.C
	//	s.Shutdown(ctx)
	//}()
	//
	//timer.Reset(5 * time.Second)
	//
	//retry.WrapWithRetry(func() error {
	// TODO:
	//}, func(_ error, _ int) bool {
	//	return true
	//}, 10
}

func SerializeInt64(n int64) []byte {
	return []byte(fmt.Sprintf("%d", n))
}

func DeserializeInt64(b []byte) (int64, error) {
	return strconv.ParseInt(string(b), 10, 64)
}
