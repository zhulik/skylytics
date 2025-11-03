package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"skylytics/pkg/stormy"
	"syscall"
	"time"

	"github.com/golang-cz/devslog"
	"github.com/mattn/go-isatty"

	"skylytics/internal/bluesky"
	"skylytics/internal/core"
	"skylytics/internal/db"

	"github.com/zhulik/pal"
)

func main() {
	w := os.Stdout

	var handler slog.Handler
	if isatty.IsTerminal(w.Fd()) {
		handler = devslog.NewHandler(w, nil)
	} else {
		handler = slog.NewJSONHandler(w, nil)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	services := []pal.ServiceDef{
		pal.ProvideFn[*stormy.Client](func(_ context.Context) (*stormy.Client, error) {
			return stormy.NewClient(nil), nil
		}),
		pal.Provide(&core.Config{}),
		db.Provide(),
		pal.Provide(&printer{}),
	}

	command := os.Args[1]
	switch command {
	case "subscriber":
		// services = append(services, inspect.Provide()...)
		services = append(services,
			pal.Provide(&saver{}),
			pal.Provide[core.BlueskySubscriber](&bluesky.Subscriber{}),
		)

	default:
		log.Fatalf("unknown command: %s", command)
	}

	err := pal.New(services...).
		InjectSlog().
		InitTimeout(2*time.Second).
		HealthCheckTimeout(1*time.Second).
		ShutdownTimeout(10*time.Second).
		Run(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	if err != nil {
		log.Fatal(err)
	}
}
