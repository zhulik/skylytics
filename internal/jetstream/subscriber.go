package jetstream

import (
	"context"
	"github.com/gorilla/websocket"
	"go.uber.org/fx"
	"log"
	"skylytics/internal/core"
)

const (
	url = "wss://jetstream2.us-east.bsky.network/subscribe"
)

type Subscriber struct {
	events chan string
}

func (s Subscriber) Chan() <-chan string {
	return s.events
}

func NewSubscriber(lc fx.Lifecycle) core.JetstreamSubscriber {
	events := make(chan string)

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
					events <- string(message)
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
