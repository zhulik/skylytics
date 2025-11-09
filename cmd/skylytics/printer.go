package main

import (
	"context"
	"log/slog"
	"skylytics/internal/core"
	"time"

	"github.com/zhulik/pal"
)

type printer struct {
	Logger *slog.Logger

	DB core.DB
}

func (p *printer) RunConfig() pal.RunConfig {
	return pal.RunConfig{
		Wait: false,
	}
}

func (p *printer) Run(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			select {
			case <-ticker.C:
				res, err := p.DB.QueryContext(ctx, "SELECT reltuples::bigint AS estimate FROM pg_class WHERE relname = 'events'")
				if err != nil {
					return err
				}
				defer res.Close()
				for res.Next() {
					var estimate int64
					err = res.Scan(&estimate)
					if err != nil {
						return err
					}
					p.Logger.Info("events", "count", estimate)
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
