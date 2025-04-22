package bluesky

import (
	"context"
	"encoding/json"

	"skylytics/internal/core"
	"skylytics/pkg/async"

	"github.com/gorilla/websocket"
	"github.com/samber/do"

	"github.com/zhulik/pips"
)

const (
	url = "wss://jetstream2.us-east.bsky.network/subscribe"
)

type Subscriber struct {
	conn   *websocket.Conn
	handle *async.JobHandle[any]
}

func (s Subscriber) Shutdown() error {
	_, err := s.handle.StopWait()
	return err
}

func (s Subscriber) HealthCheck() error {
	return s.handle.Error()
}

func (s Subscriber) Subscribe() <-chan pips.D[core.BlueskyEvent] {
	var ch <-chan pips.D[core.BlueskyEvent]

	s.handle, ch = async.Generator(func(_ context.Context, yield async.Yielder[core.BlueskyEvent]) error {
		defer s.conn.Close()

		for {
			_, message, err := s.conn.ReadMessage()
			if err != nil {
				return err
			}

			var event core.BlueskyEvent
			err = json.Unmarshal(message, &event)

			yield(event, err)
		}
	})

	return ch
}

func NewSubscriber(_ *do.Injector) (core.BlueskySubscriber, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return Subscriber{conn: conn}, nil
}
