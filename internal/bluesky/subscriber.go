package bluesky

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/samber/do"
	"log"
	"skylytics/internal/core"
	"skylytics/pkg/async"
)

const (
	url = "wss://jetstream2.us-east.bsky.network/subscribe"
)

type Subscriber struct {
	events <-chan async.Result[core.BlueskyEvent]
	conn   *websocket.Conn
}

func (s Subscriber) Shutdown() error {
	s.conn.Close()
	return nil
}

func (s Subscriber) HealthCheck() error {
	return nil
}

func (s Subscriber) Chan() <-chan async.Result[core.BlueskyEvent] {
	return s.events
}

func NewSubscriber(_ *do.Injector) (core.BlueskySubscriber, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	events := async.Generator(func(yield async.Yielder[core.BlueskyEvent]) error {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return err
			}

			var event core.BlueskyEvent

			err = json.Unmarshal(message, &event)
			if err != nil {
				log.Printf("error unmarshalling event: %+v", err)
				continue
			}

			yield(event)
		}
	})

	return Subscriber{events: events, conn: conn}, nil
}
