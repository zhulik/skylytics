package bluesky

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"skylytics/internal/core"

	"github.com/gorilla/websocket"
	"github.com/samber/lo"

	"github.com/zhulik/pips"
)

const (
	jetstreamURL = "wss://jetstream2.us-east.bsky.network/subscribe"
)

type Subscriber struct {
	JS core.JetstreamClient

	Conn *websocket.Conn
	KV   core.KeyValueClient
	ch   chan pips.D[core.BlueskyEvent]
}

func (s *Subscriber) Init(ctx context.Context) error {
	kv, err := s.JS.KV(ctx, "skylytics")
	if err != nil {
		return err
	}

	lastEventTimestampBytes, err := kv.Get(ctx, "last_event_timestamp")
	if err != nil {
		if !errors.Is(err, jetstream.ErrKeyNotFound) {
			return err
		}
	}

	lastEventTimestamp := DeserializeInt64(lastEventTimestampBytes)

	streamURL, err := url.Parse(jetstreamURL)
	if err != nil {
		return err
	}
	if lastEventTimestamp > 0 {
		params := make(url.Values)
		params.Add("cursor", fmt.Sprintf("%d", lastEventTimestamp))
		streamURL.RawQuery = params.Encode()

		log.Printf("Continuing from last event timestamp: %d, url: %s", lastEventTimestamp, streamURL.String())
	}

	conn, _, err := websocket.DefaultDialer.Dial(streamURL.String(), nil)
	if err != nil {
		return err
	}

	s.Conn = conn
	s.KV = kv
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
	defer log.Println("subscriber stopped")

	go func() {
		<-timer.C
		panic("hanged")
	}()

	for {
		var event core.BlueskyEvent

		_, message, err := s.Conn.ReadMessage()
		if err != nil {
			s.ch <- pips.NewD(event, err)

			return err
		}

		timer.Reset(5 * time.Second)

		err = json.Unmarshal(message, &event)
		if err == nil {
			err = s.KV.Put(ctx, "last_event_timestamp", SerializeInt64(event.TimeUS))
		}

		s.ch <- pips.NewD(event, err)
	}
}

func SerializeInt64(n int64) []byte {
	return []byte(fmt.Sprintf("%d", n))
}

func DeserializeInt64(b []byte) int64 {
	return lo.Must(strconv.ParseInt(string(b), 10, 64))
}
