package posts

import (
	"context"
	"log/slog"

	"skylytics/internal/core"
)

type Repository struct {
	Logger *slog.Logger
	DB     core.DB
	JS     core.JetstreamClient
}

func (r *Repository) Init(_ context.Context) error {
	r.Logger = r.Logger.With("component", "posts.Repository")
	return nil
}

func (r *Repository) Get(_ context.Context, _ string) (core.Post, error) {
	// TODO: using JS fetch all events related the CID and reconstruct the state of the post.
	return core.Post{}, nil
}

func (r *Repository) AddInteraction(ctx context.Context, interaction core.PostInteraction) error {
	return r.DB.Model(&core.PostInteraction{}).WithContext(ctx).Create(&interaction).Error
}

func (r *Repository) TopN(_ context.Context, _ int) ([]core.Post, error) {
	return nil, nil
}
