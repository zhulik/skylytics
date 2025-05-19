package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

//go:generate jsonnet openapi.jsonnet -o openapi.json
//go:generate go tool oapi-codegen -config oapi-codegen.yaml openapi.json

type Server struct {
	server *http.Server

	Logger *slog.Logger
}

func (s *Server) Run(ctx context.Context) error {
	s.Logger.Info("Starting API server", "addr", s.server.Addr)

	go func() {
		<-ctx.Done()
		// TODO: figure out a good context here, Run's ctx is cancelled.
		s.server.Shutdown(context.TODO()) //nolint:errcheck
	}()

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Init(context.Context) error {
	s.Logger = s.Logger.With("component", "api.Server")

	r := chi.NewMux()

	h := HandlerFromMux(s, r)

	s.server = &http.Server{
		Handler:           h,
		Addr:              ":8888",
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      time.Second,
		ReadTimeout:       time.Second,
		IdleTimeout:       time.Second,
	}
	return nil
}

func (s *Server) GetV1Posts(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.WriteHeader(200)
}
