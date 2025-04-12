package jetstream

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
	events <-chan async.Result[core.JetstreamEvent]
	conn   *websocket.Conn
}

func (s Subscriber) Shutdown() error {
	s.conn.Close()
	return nil
}

func (s Subscriber) HealthCheck() error {
	return nil
}

func (s Subscriber) Chan() <-chan async.Result[core.JetstreamEvent] {
	return s.events
}

func NewSubscriber(_ *do.Injector) (core.JetstreamSubscriber, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	events := async.Generator(func(yield async.Yielder[core.JetstreamEvent]) error {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return err
			}

			var event core.JetstreamEvent

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
