package metrics

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/zhulik/pal"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPServer struct {
	*http.Server
}

func (s *HTTPServer) Init(ctx context.Context) error {
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
			log.Printf("Health check failed: %+v", err)
			w.WriteHeader(500)
		}
	})

	s.Handler = mux

	return nil
}

func (s *HTTPServer) Run(_ context.Context) error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	log.Println("Starting HTTP server at", s.Addr)

	return s.Serve(ln)
}
