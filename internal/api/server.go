package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/zhulik/pal"
)

type contextKey string

const loggerContextKey = contextKey("logger")

//go:generate jsonnet openapi.jsonnet -o openapi.json
//go:generate go tool oapi-codegen -config oapi-codegen.yaml openapi.json

type Server struct {
	server *http.Server

	Logger *slog.Logger
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
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

func (s *Server) Init(ctx context.Context) error {
	s.Logger = s.Logger.With("component", "api.Server")

	r := chi.NewMux()

	logger := func(ctx context.Context) *slog.Logger {
		return ctx.Value(loggerContextKey).(*slog.Logger)
	}

	r.Use(
		middleware.OapiRequestValidatorWithOptions(spec, nil),

		// json content type
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				next.ServeHTTP(w, r)
			})
		},

		// Logging
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				logger := s.Logger.With("method", r.Method, "path", r.URL.Path)
				ctx := context.WithValue(r.Context(), loggerContextKey, logger)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		},

		// Logging
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				sw := &statusWriter{ResponseWriter: w}

				next.ServeHTTP(sw, r)

				duration := time.Since(start)
				logger(r.Context()).Info("request", "duration", duration, "status", sw.status)
			})
		},

		// Recovering panics and logging
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if err := recover(); err != nil {
						logger(r.Context()).Error("panic recovered", "error", err)
						http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
					}
				}()
				next.ServeHTTP(w, r)
			})
		},
	)

	p := pal.FromContext(ctx)

	backend, err := pal.Build[Backend](ctx, p)
	if err != nil {
		return err
	}

	h := HandlerFromMux(backend, r)

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
