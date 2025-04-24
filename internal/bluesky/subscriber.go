package bluesky

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/nats-io/nats.go/jetstream"

	"time"

	"skylytics/internal/core"
	"skylytics/pkg/async"

	"github.com/gorilla/websocket"
	"github.com/samber/do"
	"github.com/samber/lo"

	"github.com/zhulik/pips"
)

const (
	jetstreamURL = "wss://jetstream2.us-east.bsky.network/subscribe"
)

type Subscriber struct {
	conn   *websocket.Conn
	kv     core.KeyValueClient
	handle *async.JobHandle[any]
}

func NewSubscriber(i *do.Injector) (core.BlueskySubscriber, error) {
	kv := lo.Must(do.MustInvoke[core.JetstreamClient](i).KV(context.Background(), "skylytics"))

	lastEventTimestampBytes, err := kv.Get(context.Background(), "last_event_timestamp")
	if err != nil {
		if !errors.Is(err, jetstream.ErrKeyNotFound) {
			return nil, err
		}
	}

	lastEventTimestamp := DeserializeInt64(lastEventTimestampBytes)

	streamURL, err := url.Parse(jetstreamURL)
	if err != nil {
		return nil, err
	}
	if lastEventTimestamp > 0 {
		params := make(url.Values)
		params.Add("cursor", fmt.Sprintf("%d", lastEventTimestamp))
		streamURL.RawQuery = params.Encode()

		log.Printf("Continuing from last event timestamp: %d, url: %s", lastEventTimestamp, streamURL.String())
	}

	conn, _, err := websocket.DefaultDialer.Dial(streamURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return Subscriber{
		conn: conn,
		kv:   kv,
	}, nil
}

func (s Subscriber) Shutdown() error {
	if s.handle == nil {
		return nil
	}
	_, err := s.handle.StopWait()
	return err
}

func (s Subscriber) HealthCheck() error {
	if s.handle == nil {
		return nil
	}
	return s.handle.Error()
}

func (s Subscriber) Subscribe() <-chan pips.D[core.BlueskyEvent] {
	var ch <-chan pips.D[core.BlueskyEvent]

	s.handle, ch = async.Generator(func(ctx context.Context, yield async.Yielder[core.BlueskyEvent]) error {
		defer s.conn.Close()

		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()

		go func() {
			for range timer.C {
				panic("hanged")
			}
		}()

		for {
			_, message, err := s.conn.ReadMessage()
			if err != nil {
				return err
			}

			timer.Reset(5 * time.Second)

			var event core.BlueskyEvent
			err = json.Unmarshal(message, &event)
			if err == nil {
				err = s.kv.Put(ctx, "last_event_timestamp", SerializeInt64(event.TimeUS))
				if err == nil {
					log.Println("timer reset, last_event_timestamp updated")
				}
			}

			yield(event, err)
		}
	})

	return ch
}

func SerializeInt64(n int64) []byte {
	return []byte(fmt.Sprintf("%d", n))
}

func DeserializeInt64(b []byte) int64 {
	return lo.Must(strconv.ParseInt(string(b), 10, 64))
}
