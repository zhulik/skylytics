package metrics

import (
	"context"
	"github.com/samber/do"
	"log"
	"net"
	"net/http"
	"skylytics/internal/core"

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

func NewHTTPServer(_ *do.Injector) (core.MetricsServer, error) {
	srv := &http.Server{Addr: ":8080"}

	mux := http.NewServeMux()

	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	srv.Handler = mux

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return nil, err
	}
	log.Println("Starting HTTP server at", srv.Addr)
	go srv.Serve(ln)

	return HTTPServer{srv}, nil
}
