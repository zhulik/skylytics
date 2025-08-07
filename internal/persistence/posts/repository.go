package posts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"skylytics/internal/core"
	"skylytics/internal/persistence"
	"skylytics/pkg/stormy"

	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/nats-io/nats.go/jetstream"
)

type Repository struct {
	Logger *slog.Logger
	DB     core.DB
	JS     core.JetstreamClient
	Stormy *stormy.Client
	Config *core.Config
}

func (r *Repository) Get(ctx context.Context, uri string) (*core.Post, error) {
	subject := fmt.Sprintf("event.commit.*.app.bsky.feed.post.%s", base64.RawURLEncoding.EncodeToString([]byte(uri)))

	r.Logger.Info("Fetching events", "subject", subject)
	cons, err := r.JS.OrderedConsumer(ctx, r.Config.NatsStream, jetstream.OrderedConsumerConfig{
		FilterSubjects: []string{subject},
		DeliverPolicy:  jetstream.DeliverAllPolicy,
	})
	if err != nil {
		return nil, err
	}
	batch, err := cons.FetchNoWait(1000)
	if err != nil {
		return nil, err
	}
	if batch.Error() != nil {
		return nil, batch.Error()
	}

	pr := PostReconstructor{}

	for msg := range batch.Messages() {
		event := &models.Event{}
		err := json.Unmarshal(msg.Data(), event)
		if err != nil {
			return nil, err
		}
		if err := pr.AddEvent(event); err != nil {
			return nil, err
		}
	}

	if pr.Post == nil {
		r.Logger.Info("Post could not be found locally, falling back to bluesky API", "uri", uri)

		posts, err := r.Stormy.GetPosts(ctx, uri)
		if err != nil {
			return nil, err
		}
		if len(posts) == 0 {
			return nil, persistence.ErrNotFound
		}
		return &core.Post{
			DID:  "",
			Text: "",
		}, nil
	}

	return pr.Post, nil
}

func (r *Repository) AddInteraction(ctx context.Context, interactions ...*core.PostInteraction) error {
	return r.DB.Model(&core.PostInteraction{}).WithContext(ctx).Create(interactions).Error
}

func (r *Repository) TopN(_ context.Context, _ int) ([]core.Post, error) {
	return nil, nil
}
