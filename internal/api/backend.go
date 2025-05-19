package api

import (
	"context"
	"log/slog"
	"net/http"
)

type Backend struct {
	Logger *slog.Logger
}

func (b *Backend) Init(context.Context) error {
	b.Logger = b.Logger.With("component", "api.Backend")
	return nil
}

func (b *Backend) GetV1Posts(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.WriteHeader(200)
}

func (b *Backend) GetV1PostsId(_ http.ResponseWriter, _ *http.Request, _ string) { // nolint:revive
	// TODO implement me
	panic("implement me")
}
