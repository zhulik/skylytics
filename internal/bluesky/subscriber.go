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
	conn *websocket.Conn
}

func (s Subscriber) Shutdown() error {
	return s.conn.Close()
}

func (s Subscriber) HealthCheck() error {
	return nil
}

func (s Subscriber) Chan(ctx context.Context) <-chan pips.D[core.BlueskyEvent] {
	return async.Generator(ctx, func(_ context.Context, yield async.Yielder[core.BlueskyEvent]) error {
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
}

func NewSubscriber(_ *do.Injector) (core.BlueskySubscriber, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return Subscriber{conn: conn}, nil
}
