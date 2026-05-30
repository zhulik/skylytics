package bluesky

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	bsky "github.com/bluesky-social/jetstream/pkg/client"
	"github.com/bluesky-social/jetstream/pkg/client/schedulers/sequential"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

const (
	jetstreamURL = "wss://jetstream2.us-east.bsky.network/subscribe"
)

type Subscriber struct {
	Logger *slog.Logger
}

func (s *Subscriber) Consume(ctx context.Context, cursor *int64) (chan mo.Result[*models.Event], error) {
	ch := make(chan mo.Result[*models.Event], 10)

	client, err := bsky.NewClient(
		&bsky.ClientConfig{
			Compress:     true,
			WebsocketURL: jetstreamURL,
			ReadTimeout:  10 * time.Second,
		}, s.Logger, sequential.NewScheduler("scheduler", s.Logger, func(_ context.Context, event *models.Event) error {
			ch <- mo.Ok(event)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	lo.Async0(func() {
		defer close(ch)

		for {
			err := client.ConnectAndRead(ctx, cursor)

			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				ch <- mo.Err[*models.Event](err)
				return
			}
		}
	})

	return ch, nil
}

func SerializeInt64(n int64) []byte {
	return []byte(fmt.Sprintf("%d", n))
}

func DeserializeInt64(b []byte) (int64, error) {
	return strconv.ParseInt(string(b), 10, 64)
}
