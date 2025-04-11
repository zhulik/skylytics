package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"skylytics/internal/core"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

func NewHTTPServer(lc fx.Lifecycle) core.MetricsServer {
	srv := &http.Server{Addr: ":8080"}

	mux := http.NewServeMux()

	//http.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	srv.Handler = mux

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			fmt.Println("Starting HTTP server at", srv.Addr)
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
	return srv
}
