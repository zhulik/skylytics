package core

import (
	"context"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Posgresql
	DATABASE_URL string `envconfig:"DATABASE_URL"`
}

func (c *Config) Init(_ context.Context) error {
	err := envconfig.Process("skylytics", c)
	return err
}
