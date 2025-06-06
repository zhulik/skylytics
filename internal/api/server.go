package api

import (
	"context"
	"errors"

	"github.com/zhulik/pal"

	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	oapiMiddleware "github.com/oapi-codegen/nethttp-middleware"
	slogchi "github.com/samber/slog-chi"
)

//go:generate jsonnet openapi.jsonnet -o openapi.json
//go:generate go tool oapi-codegen -config oapi-codegen.yaml openapi.json

type Server struct {
	server *http.Server

	Backend StrictServerInterface
	Logger  *slog.Logger
}

func Provide() pal.ServiceDef {
	return pal.ProvideList(
		pal.Provide[StrictServerInterface, Backend](),
		pal.Provide[*Server, Server](),
	)
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

func (s *Server) Init(_ context.Context) error {
	s.Logger = s.Logger.With("component", "api.Server")

	r := chi.NewMux()
	r.Use(
		slogchi.New(s.Logger),
		middleware.Recoverer,
		middleware.RequestID,

		oapiMiddleware.OapiRequestValidatorWithOptions(spec, nil),
	)

	h := HandlerFromMux(NewStrictHandler(s.Backend, []StrictMiddlewareFunc{}), r)

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
