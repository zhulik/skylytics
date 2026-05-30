package redis

import (
	"context"
	"skylytics/internal/config"

	libredis "github.com/redis/go-redis/v9"
)

type Client struct {
	*libredis.Client

	Config *config.Config
}

func (c *Client) Init(ctx context.Context) error {
	c.Client = libredis.NewFailoverClient(&libredis.FailoverOptions{
		MasterName:    "master",
		SentinelAddrs: []string{c.Config.RedisAddr},
	})

	return c.Ping(ctx).Err()
}

func (c *Client) HealthCheck(ctx context.Context) error {
	return c.Ping(ctx).Err()
}

func (c *Client) Shutdown(_ context.Context) error {
	return c.Close()
}
