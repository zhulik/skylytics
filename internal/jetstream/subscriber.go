package jetstream

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/samber/do"
	"log"
	"skylytics/internal/core"
)

const (
	url = "wss://jetstream2.us-east.bsky.network/subscribe"
)

type Subscriber struct {
	events chan core.JetstreamEvent
	conn   *websocket.Conn
}

func (s Subscriber) Shutdown() error {
	s.conn.Close()
	close(s.events)
	return nil
}

func (s Subscriber) HealthCheck() error {
	return nil
}

func (s Subscriber) Chan() <-chan core.JetstreamEvent {
	return s.events
}

func NewSubscriber(_ *do.Injector) (core.JetstreamSubscriber, error) {
	events := make(chan core.JetstreamEvent)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	go func() {
		defer log.Println("Subscriber stopped")

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				panic(err)
			}

			var event core.JetstreamEvent

			err = json.Unmarshal(message, &event)
			if err != nil {
				log.Printf("error unmarshalling event: %+v", err)
				continue
			}

			events <- event
		}
	}()

	return Subscriber{events: events, conn: conn}, nil
}
