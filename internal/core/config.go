package core

import (
	"context"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Posgresql
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

func (c *Config) Init(_ context.Context) error {
	err := envconfig.Process("skylytics", c)
	return err
}
