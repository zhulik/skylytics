package redis

import (
	"context"
	"log/slog"
	"skylytics/internal/config"

	libredis "github.com/redis/go-redis/v9"
)

type Client struct {
	*libredis.Client

	Logger *slog.Logger

	Config *config.Config
}

func (c *Client) Init(ctx context.Context) error {
	c.Client = libredis.NewFailoverClient(&libredis.FailoverOptions{
		MasterName:    "mymaster",
		SentinelAddrs: []string{c.Config.RedisAddr},
	})

	c.Logger.Info("connecting to Redis Sentinel", "addr", c.Config.RedisAddr)

	return c.Ping(ctx).Err()
}

func (c *Client) HealthCheck(ctx context.Context) error {
	return c.Ping(ctx).Err()
}

func (c *Client) Shutdown(_ context.Context) error {
	return c.Close()
}

func (c *Client) CountKeys(ctx context.Context, pattern string) (int64, error) {
	var count int64
	var cursor uint64

	for {
		keys, nextCursor, err := c.Client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return 0, err
		}

		count += int64(len(keys))
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return count, nil
}
