package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"skylytics/internal/core"
)

type Backend struct {
	Logger    *slog.Logger
	PostsRepo core.PostRepository
}

func (b *Backend) GetV1Openapi(_ context.Context, _ GetV1OpenapiRequestObject) (GetV1OpenapiResponseObject, error) {
	s := map[string]any{}

	err := json.Unmarshal(openAPISpec, &s)
	if err != nil {
		return nil, err
	}

	return GetV1Openapi200JSONResponse(s), nil
}

func (b *Backend) GetV1UsersDidAppBskyFeedPostRkey(ctx context.Context, request GetV1UsersDidAppBskyFeedPostRkeyRequestObject) (GetV1UsersDidAppBskyFeedPostRkeyResponseObject, error) {
	uri := fmt.Sprintf("at://%s/app.bsky.feed.post/%s", request.Did, request.Rkey)
	post, err := b.PostsRepo.Get(ctx, uri)
	if err != nil {
		return nil, err
	}

	return GetV1UsersDidAppBskyFeedPostRkey200JSONResponse(Post{
		Did:  &post.DID,
		Text: &post.Text,
	}), nil
}

func (b *Backend) Init(context.Context) error {
	b.Logger = b.Logger.With("component", "api.Backend")
	return nil
}
