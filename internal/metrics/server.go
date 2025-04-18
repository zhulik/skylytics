package metrics

import (
	"context"
	"log"
	"net"
	"net/http"

	"skylytics/internal/core"

	"github.com/samber/do"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPServer struct {
	srv *http.Server
}

func (s HTTPServer) Shutdown() error {
	return s.srv.Shutdown(context.Background())
}

func (s HTTPServer) HealthCheck() error {
	return nil
}

func NewHTTPServer(i *do.Injector) (core.MetricsServer, error) {
	srv := &http.Server{Addr: ":8080"}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := i.HealthCheck(); err != nil {
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
	go srv.Serve(ln)

	return HTTPServer{srv}, nil
}
