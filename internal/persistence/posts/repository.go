package posts

import (
	"context"
	"log/slog"
	"time"

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

func (r *Repository) AddStats(_ context.Context, _ string, _ int64, _ int64, _ int64) error {
	return nil
}

// Cleanup deletes records older than d.
func (r *Repository) Cleanup(_ context.Context, _ time.Duration) error {
	return nil
}

func (r *Repository) TopN(_ context.Context, _ int) ([]core.Post, error) {
	return nil, nil
}
