package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"skylytics/internal/core"
	"skylytics/internal/persistence"
)

type Backend struct {
	Logger    *slog.Logger
	PostsRepo core.PostRepository
}

func (b *Backend) GetV1Openapi(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.WriteHeader(200)
	w.Write(openAPISpec) // nolint:errcheck
}

func (b *Backend) Init(context.Context) error {
	b.Logger = b.Logger.With("component", "api.Backend")
	return nil
}

func (b *Backend) GetV1Posts(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.WriteHeader(200)
}

func (b *Backend) GetV1PostsId(w http.ResponseWriter, r *http.Request, cid string) { // nolint:revive
	post, err := b.PostsRepo.Get(r.Context(), cid)
	if err != nil {
		if errors.Is(err, persistence.ErrNotFound) {
			w.WriteHeader(404)
			return
		}
		panic(err)
	}

	data, err := json.Marshal(post)
	if err != nil {
		panic(err)
	}
	w.WriteHeader(200)
	w.Write(data) // nolint:errcheck
}
