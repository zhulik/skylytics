package jetstream

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"go.uber.org/fx"
	"log"
	"skylytics/internal/core"
)

const (
	url = "wss://jetstream2.us-east.bsky.network/subscribe"
)

type Subscriber struct {
	events chan core.JetstreamEvent
}

func (s Subscriber) Chan() <-chan core.JetstreamEvent {
	return s.events
}

func NewSubscriber(lc fx.Lifecycle) core.JetstreamSubscriber {
	events := make(chan core.JetstreamEvent)

	var conn *websocket.Conn

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var err error
			conn, _, err = websocket.DefaultDialer.Dial(url, nil)
			if err != nil {
				return err
			}

			go func() {
				defer log.Println("Subscriber stopped")
				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						if websocket.IsUnexpectedCloseError(err) {
							log.Printf("websocket connection closed unexpectedly: %+v", err)
						}
						return
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
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Stopping subscriber")
			close(events)
			return conn.Close()
		},
	})

	return Subscriber{events}
}
