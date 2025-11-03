package main

import (
	"context"
	"log/slog"
	"skylytics/db/models"
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
				count, err := models.Events.Query().Count(ctx, p.DB)
				if err != nil {
					return err
				}
				p.Logger.Info("events", "count", count)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
