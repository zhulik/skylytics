package bluesky

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"skylytics/pkg/retry"

	"strconv"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"skylytics/internal/core"

	"github.com/gorilla/websocket"

	"github.com/zhulik/pips"
)

const (
	jetstreamURL = "wss://jetstream2.us-east.bsky.network/subscribe"
)

type Subscriber struct {
	Logger *slog.Logger

	JS core.JetstreamClient

	Conn *websocket.Conn
	KV   core.KeyValueClient
	ch   chan pips.D[core.BlueskyEvent]
}

func (s *Subscriber) Init(ctx context.Context) error {
	s.Logger = s.Logger.With("component", s)
	var err error
	s.KV, err = s.JS.KV(ctx, os.Getenv("NATS_STATE_KV_BUCKET"))
	if err != nil {
		return err
	}

	s.ch = make(chan pips.D[core.BlueskyEvent])

	return nil
}

func (s *Subscriber) Shutdown(_ context.Context) error {
	return s.Conn.Close()
}

func (s *Subscriber) C() <-chan pips.D[core.BlueskyEvent] {
	return s.ch
}

func (s *Subscriber) Run(ctx context.Context) error {
	defer close(s.ch)
	timer := time.NewTimer(5 * time.Second)

	defer timer.Stop()
	defer s.Logger.Info("Subscriber stopped")

	go func() {
		<-timer.C
		s.Conn.Close()
	}()

	return retry.WrapWithRetry(func() error {
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

		streamURL, err := url.Parse(jetstreamURL)
		if err != nil {
			return err
		}
		if lastEventTimestamp > 0 {
			params := make(url.Values)
			params.Add("cursor", fmt.Sprintf("%d", lastEventTimestamp))
			streamURL.RawQuery = params.Encode()
		}

		s.Logger.Info("Listening to bluesky jetstream", "url", streamURL.String())

		conn, _, err := websocket.DefaultDialer.Dial(streamURL.String(), nil)
		if err != nil {
			return err
		}

		s.Conn = conn

		for {
			var event core.BlueskyEvent

			_, message, err := s.Conn.ReadMessage()
			if err != nil {
				// We do not forward errors to the channel, if something fails, we close the channel instead
				return err
			}

			timer.Reset(5 * time.Second)

			err = json.Unmarshal(message, &event)
			if err == nil {
				err = s.KV.Put(ctx, "last_event_timestamp", SerializeInt64(event.TimeUS))
			}

			s.ch <- pips.NewD(event, err)
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
