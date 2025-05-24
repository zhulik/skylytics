package api

import (
	"context"
	"encoding/json"
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

func (b *Backend) GetV1Posts(_ context.Context, _ GetV1PostsRequestObject) (GetV1PostsResponseObject, error) {
	return nil, nil
}

func (b *Backend) GetV1PostsId(ctx context.Context, request GetV1PostsIdRequestObject) (GetV1PostsIdResponseObject, error) { // nolint:revive
	post, err := b.PostsRepo.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return GetV1PostsId200JSONResponse(Post{
		Did:  &post.DID,
		Text: &post.Text,
	}), nil
}

func (b *Backend) Init(context.Context) error {
	b.Logger = b.Logger.With("component", "api.Backend")
	return nil
}
