package bluesky

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/samber/do"
	"github.com/zhulik/pips"
	"skylytics/internal/core"
	"skylytics/pkg/async"
)

const (
	url = "wss://jetstream2.us-east.bsky.network/subscribe"
)

type Subscriber struct {
	events <-chan pips.D[core.BlueskyEvent]
	conn   *websocket.Conn
}

func (s Subscriber) Shutdown() error {
	s.conn.Close()
	return nil
}

func (s Subscriber) HealthCheck() error {
	return nil
}

func (s Subscriber) Chan() <-chan pips.D[core.BlueskyEvent] {
	return s.events
}

func NewSubscriber(_ *do.Injector) (core.BlueskySubscriber, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	events := async.Generator(context.TODO(), func(_ context.Context, yield async.Yielder[core.BlueskyEvent]) error {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return err
			}

			var event core.BlueskyEvent
			err = json.Unmarshal(message, &event)

			yield(event, err)
		}
	})

	return Subscriber{events: events, conn: conn}, nil
}
