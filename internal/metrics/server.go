package metrics

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/samber/lo"

	"skylytics/internal/core"

	"github.com/samber/do"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPServer struct {
	srv *http.Server
}

func NewHTTPServer(i *do.Injector) (core.MetricsServer, error) {
	srv := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      time.Second,
		ReadTimeout:       time.Second,
		IdleTimeout:       time.Second,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		checks := i.HealthCheck()
		defer r.Body.Close()

		err := errors.Join(lo.Values(checks)...)
		if err != nil {
			log.Printf("Health check failed: %+v", err)
			w.WriteHeader(500)
		}
	})

	srv.Handler = mux

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return nil, err
	}
	log.Println("Starting HTTP server at", srv.Addr)
	go srv.Serve(ln) //nolint:errcheck

	return HTTPServer{srv}, nil
}

func (s HTTPServer) Shutdown() error {
	return s.srv.Shutdown(context.Background())
}

func (s HTTPServer) HealthCheck() error {
	// If the server is unhealthy, the health endpoint is not accessible anyway
	return nil
}
