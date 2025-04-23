package bluesky

import (
	"context"
	"encoding/json"
	"time"

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

	s.handle, ch = async.Generator(func(_ context.Context, yield async.Yielder[core.BlueskyEvent]) error {
		defer s.conn.Close()

		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()

		go func() {
			for range timer.C {
				s.conn.Close()
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
