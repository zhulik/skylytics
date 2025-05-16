package core

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Logger *slog.Logger
	// Common
	NatsURL      string `envconfig:"NATS_URL"`
	NatsConsumer string `envconfig:"NATS_CONSUMER"`
	NatsStream   string `envconfig:"NATS_STREAM"`

	// bsky subscriber
	NatsStateKVBucket string `envconfig:"NATS_STATE_KV_BUCKET"`

	// Accounts cache
	NatsAccountsCacheKVBucket string `envconfig:"NATS_ACCOUNTS_CACHE_KV_BUCKET"`

	// Posgresql
	PostgresHost     string `envconfig:"POSTGRES_HOST"`
	PostgresUser     string `envconfig:"POSTGRES_USER"`
	PostgresPassword string `envconfig:"POSTGRES_PASSWORD"`
	PostgresDB       string `envconfig:"POSTGRES_DB"`
}

func (c *Config) Init(_ context.Context) error {
	err := envconfig.Process("skylytics", c)
	c.Logger.Info("Config loaded", "config", *c, "error", err)
	return err
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s", c.PostgresHost, c.PostgresUser, c.PostgresPassword, c.PostgresDB)
}
