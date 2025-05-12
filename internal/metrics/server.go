package metrics

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/zhulik/pal"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPServer struct {
	*http.Server

	Logger *slog.Logger
}

func (s *HTTPServer) Init(ctx context.Context) error {
	s.Logger = s.Logger.With("component", s)

	s.Server = &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      time.Second,
		ReadTimeout:       time.Second,
		IdleTimeout:       time.Second,
	}

	p := pal.FromContext(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		err := p.HealthCheck(r.Context())
		defer r.Body.Close()

		if err != nil {
			s.Logger.Warn("Health check failed", "error", err)
			w.WriteHeader(500)
		}
	})

	s.Handler = mux

	return nil
}

func (s *HTTPServer) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	s.Logger.Info("Starting HTTP server", "addr", s.Addr)

	go func() {
		<-ctx.Done()
		// TODO: figure out a good context here, Run's ctx is cancelled.
		s.Shutdown(context.TODO()) //nolint:errcheck
	}()

	err = s.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}
