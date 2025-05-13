package core

import (
	"context"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Common
	NatsURL      string `envconfig:"NATS_URL"`
	NatsConsumer string `envconfig:"NATS_CONSUMER"`
	NatsStream   string `envconfig:"NATS_STREAM"`

	// bsky subscriber
	NatsStateKVBucket string `envconfig:"NATS_STATE_KV_BUCKET"`

	// Posgresql
	PostgresHost     string `envconfig:"POSTGRES_HOST"`
	PostgresUser     string `envconfig:"POSTGRES_USER"`
	PostgresPassword string `envconfig:"POSTGRES_PASSWORD"`
	PostgresDB       string `envconfig:"POSTGRES_DB"`
}

func (c *Config) Init(_ context.Context) error {
	return envconfig.Process("skylytics", c)
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s", c.PostgresHost, c.PostgresUser, c.PostgresPassword, c.PostgresDB)
}
